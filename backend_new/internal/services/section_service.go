package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/cohesion-dev/GNX/ai/gnxaigc"
	"github.com/cohesion-dev/GNX/backend_new/internal/models"
	"github.com/cohesion-dev/GNX/backend_new/internal/repositories"
	"github.com/cohesion-dev/GNX/backend_new/pkg/imageutil"
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
	comic, err := s.comicRepo.FindByID(comicID)
	if err != nil {
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
		return nil, fmt.Errorf("failed to create section: %w", err)
	}

	if err := s.processSectionSync(ctx, comic, section); err != nil {
		return nil, fmt.Errorf("failed to process section: %w", err)
	}

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
	roles, err := s.roleRepo.FindByComicID(comic.ID)
	if err != nil {
		s.updateSectionStatus(section.ID, "failed")
		return fmt.Errorf("failed to get roles for section %d: %w", section.ID, err)
	}

	voices, err := s.aigc.GetVoiceList(ctx)
	if err != nil {
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

	summary, err := s.aigc.SummaryChapter(ctx, gnxaigc.SummaryChapterInput{
		NovelTitle:           comic.Title,
		ChapterTitle:         section.Title,
		Content:              section.Content,
		AvailableVoiceStyles: voiceItems,
		CharacterFeatures:    charFeatures,
		MaxPanelsPerPage:     4,
	})
	if err != nil {
		s.updateSectionStatus(section.ID, "failed")
		return fmt.Errorf("failed to generate summary for section %d: %w", section.ID, err)
	}

	for pageIndex, storyboardPage := range summary.StoryboardPages {
		page := &models.ComicPage{
			SectionID:   section.ID,
			Index:       pageIndex + 1,
			ImagePrompt: storyboardPage.ImagePrompt,
		}

		if err := s.pageRepo.Create(page); err != nil {
			fmt.Printf("Failed to create page: %v\n", err)
			continue
		}

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
					fmt.Printf("Failed to create page detail: %v\n", err)
				}
			}
		}
	}

	s.updateSectionStatus(section.ID, "completed")

	go s.processSectionImages(context.Background(), comic, section.ID)

	return nil
}

func (s *SectionService) processSectionImages(ctx context.Context, comic *models.Comic, sectionID uint) {
	section, err := s.sectionRepo.FindByID(sectionID)
	if err != nil {
		fmt.Printf("Failed to get section %d for image processing: %v\n", sectionID, err)
		return
	}

	roles, err := s.roleRepo.FindByComicID(comic.ID)
	if err != nil {
		fmt.Printf("Failed to get roles for section %d: %v\n", sectionID, err)
		return
	}

	voices, err := s.aigc.GetVoiceList(ctx)
	if err != nil {
		fmt.Printf("Failed to get voice list for section %d: %v\n", sectionID, err)
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
		fmt.Printf("Failed to generate summary for section %d images: %v\n", sectionID, err)
		return
	}

	characterAssets, err := s.charService.SyncCharacterAssets(ctx, comic.ID, comic.UserPrompt, summary.CharacterFeatures)
	if err != nil {
		fmt.Printf("Failed to sync character assets for section %d: %v\n", sectionID, err)
		characterAssets = make(map[string]*CharacterAsset)
	}

	pages, err := s.pageRepo.FindBySectionID(sectionID)
	if err != nil {
		fmt.Printf("Failed to get pages for section %d: %v\n", sectionID, err)
		return
	}

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
	)

	totalPages := len(summary.StoryboardPages)
	for pageIndex, storyboardPage := range summary.StoryboardPages {
		if pageIndex >= len(pages) {
			break
		}

		wg.Add(1)
		go func(pageIndex int, page models.ComicPage, storyboardPage gnxaigc.StoryboardPage) {
			defer wg.Done()

			mu.Lock()
			fmt.Printf("  [Page %d/%d] Generating image...\n", pageIndex+1, totalPages)
			mu.Unlock()

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
				mu.Lock()
				fmt.Printf("    Using single reference image for page %d\n", pageIndex+1)
				mu.Unlock()
				imageData, err = s.aigc.GenerateImageByImage(ctx, referenceImages[0], fullPrompt)
				if err != nil {
					mu.Lock()
					fmt.Printf("    Error generating page image via img2img: %v\n", err)
					fmt.Printf("    Falling back to text-to-image for page %d\n", pageIndex+1)
					mu.Unlock()
					imageData, err = s.aigc.GenerateImageByText(ctx, fullPrompt)
				}
			default:
				composite, mergeErr := imageutil.MergeImagesSideBySide(referenceImages)
				if mergeErr != nil {
					mu.Lock()
					fmt.Printf("    Warning: failed to merge %d reference images for page %d: %v\n", len(referenceImages), pageIndex+1, mergeErr)
					mu.Unlock()
					imageData, err = s.aigc.GenerateImageByText(ctx, fullPrompt)
					break
				}

				mu.Lock()
				fmt.Printf("    Merged %d reference images for page %d\n", len(referenceImages), pageIndex+1)
				mu.Unlock()
				imageData, err = s.aigc.GenerateImageByImage(ctx, composite, fullPrompt)
				if err != nil {
					mu.Lock()
					fmt.Printf("    Error generating page image via merged img2img: %v\n", err)
					fmt.Printf("    Falling back to text-to-image for page %d\n", pageIndex+1)
					mu.Unlock()
					imageData, err = s.aigc.GenerateImageByText(ctx, fullPrompt)
				}
			}

			if err != nil {
				mu.Lock()
				fmt.Printf("    Error generating image for page %d: %v\n", pageIndex+1, err)
				mu.Unlock()
			} else {
				imageID := fmt.Sprintf("%d", page.ID)
				if err := s.storage.UploadBytes(imageData, imageID); err != nil {
					mu.Lock()
					fmt.Printf("    Error uploading image for page %d: %v\n", pageIndex+1, err)
					mu.Unlock()
				} else {
					mu.Lock()
					fmt.Printf("    Successfully uploaded image for page %d\n", pageIndex+1)
					mu.Unlock()
				}
			}
		}(pageIndex, pages[pageIndex], storyboardPage)
	}

	wg.Wait()
	fmt.Printf("Completed image generation for section %d\n", sectionID)
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
