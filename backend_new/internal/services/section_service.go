package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/cohesion-dev/GNX/ai/gnxaigc"
	"github.com/cohesion-dev/GNX/backend_new/internal/models"
	"github.com/cohesion-dev/GNX/backend_new/internal/repositories"
	"github.com/cohesion-dev/GNX/backend_new/pkg/imageutil"
	"github.com/cohesion-dev/GNX/backend_new/pkg/logger"
	"github.com/cohesion-dev/GNX/backend_new/pkg/storage"
)

type SectionService struct {
	comicRepo    *repositories.ComicRepository
	roleRepo     *repositories.RoleRepository
	sectionRepo  *repositories.SectionRepository
	pageRepo     *repositories.PageRepository
	storage      *storage.Storage
	aigc         *gnxaigc.GnxAIGC
	charService  *CharacterService
}

func NewSectionService(
	comicRepo *repositories.ComicRepository,
	roleRepo *repositories.RoleRepository,
	sectionRepo *repositories.SectionRepository,
	pageRepo *repositories.PageRepository,
	storage *storage.Storage,
	aigc *gnxaigc.GnxAIGC,
	charService *CharacterService,
) *SectionService {
	return &SectionService{
		comicRepo:    comicRepo,
		roleRepo:     roleRepo,
		sectionRepo:  sectionRepo,
		pageRepo:     pageRepo,
		storage:      storage,
		aigc:         aigc,
		charService:  charService,
	}
}

func (s *SectionService) CreateSection(ctx context.Context, comicID uint, title, content string) (*models.ComicSection, error) {
	logger.Info("[Section Creation] Starting section creation: comicID=%d, title=%s", comicID, title)
	comic, err := s.comicRepo.FindByID(comicID)
	if err != nil {
		logger.Error("[Section Creation] Comic not found: comicID=%d, error=%v", comicID, err)
		return nil, fmt.Errorf("comic not found: %w", err)
	}

	count, err := s.sectionRepo.CountByComicID(comicID)
	if err != nil {
		return nil, fmt.Errorf("failed to count sections: %w", err)
	}

	section := &models.ComicSection{
		ComicID: comicID,
		Title:   title,
		Index:   int(count) + 1,
		Content: content,
		Status:  "pending",
	}

	if err := s.sectionRepo.Create(section); err != nil {
		logger.Error("[Section Creation] Failed to create section: %v", err)
		return nil, fmt.Errorf("failed to create section: %w", err)
	}
	logger.Info("[Section Creation] Section created: ID=%d, index=%d, status=pending", section.ID, section.Index)

	logger.Info("[Section Processing] Starting AI processing for section ID=%d", section.ID)
	if err := s.processSectionSync(ctx, comic, section); err != nil {
		logger.Error("[Section Processing] Failed to process section: %v", err)
		return nil, fmt.Errorf("failed to process section: %w", err)
	}

	logger.Info("[Section Creation] Section ID=%d created successfully", section.ID)
	return section, nil
}

func (s *SectionService) GetSectionDetail(comicID, sectionID uint) (*models.ComicSection, error) {
	section, err := s.sectionRepo.FindByID(sectionID)
	if err != nil {
		return nil, fmt.Errorf("section not found: %w", err)
	}

	if section.ComicID != comicID {
		return nil, fmt.Errorf("section does not belong to comic")
	}

	return section, nil
}

func (s *SectionService) processSectionSync(ctx context.Context, comic *models.Comic, section *models.ComicSection) error {
	logger.Info("[Section Processing] Loading character roles for section ID=%d", section.ID)
	roles, err := s.roleRepo.FindByComicID(comic.ID)
	if err != nil {
		logger.Error("[Section Processing] Failed to get roles for section %d: %v", section.ID, err)
		s.updateSectionStatus(section.ID, "failed")
		return fmt.Errorf("failed to get roles for section %d: %w", section.ID, err)
	}
	logger.Info("[Section Processing] Loaded %d character roles", len(roles))

	logger.Info("[Section Processing] Fetching available voice list")
	voices, err := s.aigc.GetVoiceList(ctx)
	if err != nil {
		logger.Error("[Section Processing] Failed to get voice list: %v", err)
		s.updateSectionStatus(section.ID, "failed")
		return fmt.Errorf("failed to get voice list for section %d: %w", section.ID, err)
	}

	voiceItems := make([]gnxaigc.TTSVoiceItem, 0, len(voices))
	for _, v := range voices {
		voiceItems = append(voiceItems, gnxaigc.TTSVoiceItem{
			VoiceName: v.VoiceName,
			VoiceType: v.VoiceType,
		})
	}

	charFeatures := make([]gnxaigc.CharacterFeature, 0, len(roles))
	for _, role := range roles {
		charFeatures = append(charFeatures, gnxaigc.CharacterFeature{
			Basic: gnxaigc.CharacterBasicProfile{
				Name:   role.Name,
				Gender: role.Gender,
				Age:    role.Age,
			},
			TTS: gnxaigc.CharacterTTSProfile{
				VoiceName:  role.VoiceName,
				VoiceType:  role.VoiceType,
				SpeedRatio: 1.0,
			},
			Comment: role.Brief,
		})
	}

	logger.Info("[Section Processing] Generating AI summary for section ID=%d", section.ID)
	summary, err := s.aigc.SummaryChapter(ctx, gnxaigc.SummaryChapterInput{
		NovelTitle:           comic.Title,
		ChapterTitle:         section.Title,
		Content:              section.Content,
		AvailableVoiceStyles: voiceItems,
		CharacterFeatures:    charFeatures,
		MaxPanelsPerPage:     4,
	})
	if err != nil {
		logger.Error("[Section Processing] Failed to generate AI summary: %v", err)
		s.updateSectionStatus(section.ID, "failed")
		return fmt.Errorf("failed to generate summary for section %d: %w", section.ID, err)
	}
	logger.Info("[Section Processing] AI summary generated: %d storyboard pages", len(summary.StoryboardPages))

	logger.Info("[Section Processing] Creating %d pages for section ID=%d", len(summary.StoryboardPages), section.ID)
	for pageIndex, storyboardPage := range summary.StoryboardPages {
		page := &models.ComicPage{
			SectionID:   section.ID,
			Index:       pageIndex + 1,
			ImagePrompt: storyboardPage.ImagePrompt,
		}

		if err := s.pageRepo.Create(page); err != nil {
			logger.Error("[Section Processing] Failed to create page %d: %v", pageIndex+1, err)
			continue
		}
		logger.Info("[Section Processing] Created page %d (ID=%d) with %d panels", pageIndex+1, page.ID, len(storyboardPage.Panels))

		for panelIndex, panel := range storyboardPage.Panels {
			for segmentIndex, segment := range panel.SourceTextSegments {
				var roleID *uint
				if len(segment.CharacterNames) > 0 {
					for _, role := range roles {
						if role.Name == segment.CharacterNames[0] {
							roleID = &role.ID
							break
						}
					}
				}

				detail := &models.ComicPageDetail{
					PageID:  page.ID,
					Index:   (panelIndex * 100) + segmentIndex,
					Content: segment.Text,
					RoleID:  roleID,
				}

				if err := s.pageRepo.CreateDetail(detail); err != nil {
					logger.Error("[Section Processing] Failed to create page detail: %v", err)
				}
			}
		}
	}

	s.updateSectionStatus(section.ID, "completed")
	logger.Info("[Section Processing] Section ID=%d marked as completed", section.ID)

	logger.Info("[Section Image Processing] Starting image generation for section ID=%d", section.ID)
	go s.processSectionImages(context.Background(), comic, section.ID)

	return nil
}

func (s *SectionService) processSectionImages(ctx context.Context, comic *models.Comic, sectionID uint) {
	logger.Info("[Section Image Processing] Loading section data for ID=%d", sectionID)
	section, err := s.sectionRepo.FindByID(sectionID)
	if err != nil {
		logger.Error("[Section Image Processing] Failed to get section %d: %v", sectionID, err)
		return
	}

	roles, err := s.roleRepo.FindByComicID(comic.ID)
	if err != nil {
		logger.Error("[Section Image Processing] Failed to get roles: %v", err)
		return
	}

	voices, err := s.aigc.GetVoiceList(ctx)
	if err != nil {
		logger.Error("[Section Image Processing] Failed to get voice list: %v", err)
		return
	}

	voiceItems := make([]gnxaigc.TTSVoiceItem, 0, len(voices))
	for _, v := range voices {
		voiceItems = append(voiceItems, gnxaigc.TTSVoiceItem{
			VoiceName: v.VoiceName,
			VoiceType: v.VoiceType,
		})
	}

	charFeatures := make([]gnxaigc.CharacterFeature, 0, len(roles))
	for _, role := range roles {
		charFeatures = append(charFeatures, gnxaigc.CharacterFeature{
			Basic: gnxaigc.CharacterBasicProfile{
				Name:   role.Name,
				Gender: role.Gender,
				Age:    role.Age,
			},
			TTS: gnxaigc.CharacterTTSProfile{
				VoiceName:  role.VoiceName,
				VoiceType:  role.VoiceType,
				SpeedRatio: 1.0,
			},
			Comment: role.Brief,
		})
	}

	summary, err := s.aigc.SummaryChapter(ctx, gnxaigc.SummaryChapterInput{
		NovelTitle:           comic.Title,
		ChapterTitle:         section.Title,
		Content:              section.Content,
		AvailableVoiceStyles: voiceItems,
		CharacterFeatures:    charFeatures,
		MaxPanelsPerPage:     4,
	})
	if err != nil {
		logger.Error("[Section Image Processing] Failed to generate summary: %v", err)
		return
	}

	logger.Info("[Section Image Processing] Syncing character assets for section ID=%d", sectionID)
	characterAssets, err := s.charService.SyncCharacterAssets(ctx, comic.ID, comic.UserPrompt, summary.CharacterFeatures)
	if err != nil {
		logger.Error("[Section Image Processing] Failed to sync character assets: %v", err)
		characterAssets = make(map[string]*CharacterAsset)
	}

	pages, err := s.pageRepo.FindBySectionID(sectionID)
	if err != nil {
		logger.Error("[Section Image Processing] Failed to get pages: %v", err)
		return
	}
	logger.Info("[Section Image Processing] Starting parallel image generation for %d pages", len(pages))

	var wg sync.WaitGroup

	totalPages := len(summary.StoryboardPages)
	logger.Info("[Section Image Processing] Processing %d pages in parallel", totalPages)
	for pageIndex, storyboardPage := range summary.StoryboardPages {
		if pageIndex >= len(pages) {
			break
		}

		wg.Add(1)
		go func(pageIndex int, page models.ComicPage, storyboardPage gnxaigc.StoryboardPage) {
			defer wg.Done()

			logger.Info("[Section Image Processing] Page %d/%d: Generating image", pageIndex+1, totalPages)

			fullPrompt := gnxaigc.ComposePageImagePrompt(comic.UserPrompt, storyboardPage)

			referenceKeys := s.collectPageCharacterKeys(storyboardPage, summary.CharacterFeatures)
			var referenceImages [][]byte
			for _, key := range referenceKeys {
				asset := characterAssets[key]
				if asset == nil || len(asset.ImageData) == 0 {
					continue
				}
				referenceImages = append(referenceImages, asset.ImageData)
			}

			var (
				imageData []byte
				err       error
			)

			switch len(referenceImages) {
			case 0:
				imageData, err = s.aigc.GenerateImageByText(ctx, fullPrompt)
			case 1:
				logger.Info("[Section Image Processing] Page %d/%d: Using single reference image", pageIndex+1, totalPages)
				imageData, err = s.aigc.GenerateImageByImage(ctx, referenceImages[0], fullPrompt)
				if err != nil {
					logger.Warn("[Section Image Processing] Page %d/%d: img2img failed (%v), falling back to text-to-image", pageIndex+1, totalPages, err)
					imageData, err = s.aigc.GenerateImageByText(ctx, fullPrompt)
				}
			default:
				composite, mergeErr := imageutil.MergeImagesSideBySide(referenceImages)
				if mergeErr != nil {
					logger.Warn("[Section Image Processing] Page %d/%d: Failed to merge %d reference images (%v), using text-to-image", pageIndex+1, totalPages, len(referenceImages), mergeErr)
					imageData, err = s.aigc.GenerateImageByText(ctx, fullPrompt)
					break
				}

				logger.Info("[Section Image Processing] Page %d/%d: Using %d merged reference images", pageIndex+1, totalPages, len(referenceImages))
				imageData, err = s.aigc.GenerateImageByImage(ctx, composite, fullPrompt)
				if err != nil {
					logger.Warn("[Section Image Processing] Page %d/%d: merged img2img failed (%v), falling back to text-to-image", pageIndex+1, totalPages, err)
					imageData, err = s.aigc.GenerateImageByText(ctx, fullPrompt)
				}
			}

			if err != nil {
				logger.Error("[Section Image Processing] Page %d/%d: Failed to generate image: %v", pageIndex+1, totalPages, err)
			} else {
				imageID := fmt.Sprintf("%d", page.ID)
				if err := s.storage.UploadBytes(imageData, imageID); err != nil {
					logger.Error("[Section Image Processing] Page %d/%d: Failed to upload image: %v", pageIndex+1, totalPages, err)
				} else {
					logger.Info("[Section Image Processing] Page %d/%d: Image uploaded successfully (imageID=%s)", pageIndex+1, totalPages, imageID)
				}
			}
		}(pageIndex, pages[pageIndex], storyboardPage)
	}

	wg.Wait()
	logger.Info("[Section Image Processing] Completed image generation for section ID=%d", sectionID)
}

func (s *SectionService) collectPageCharacterKeys(page gnxaigc.StoryboardPage, features []gnxaigc.CharacterFeature) []string {
	nameSet := make(map[string]bool)
	for _, panel := range page.Panels {
		for _, segment := range panel.SourceTextSegments {
			for _, name := range segment.CharacterNames {
				if name != "" {
					nameSet[name] = true
				}
			}
		}
	}

	var keys []string
	for _, feature := range features {
		if nameSet[feature.Basic.Name] {
			keys = append(keys, feature.Basic.Name)
		}
	}
	return keys
}

func (s *SectionService) updateSectionStatus(sectionID uint, status string) {
	section, err := s.sectionRepo.FindByID(sectionID)
	if err != nil {
		return
	}
	section.Status = status
	s.sectionRepo.Update(section)
}
