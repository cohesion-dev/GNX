package services

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/cohesion-dev/GNX/ai/gnxaigc"
	"github.com/cohesion-dev/GNX/backend_new/internal/models"
	"github.com/cohesion-dev/GNX/backend_new/internal/repositories"
	"github.com/cohesion-dev/GNX/backend_new/pkg/imageutil"
	"github.com/cohesion-dev/GNX/backend_new/pkg/logger"
	"github.com/cohesion-dev/GNX/backend_new/pkg/storage"
	"github.com/google/uuid"
)

var chapterHeadingPattern = regexp.MustCompile(`^第[零〇一二三四五六七八九十百千万0-9]+章.*$`)

type CharacterAsset struct {
	Feature   gnxaigc.CharacterFeature
	ImageData []byte
	Prompt    string
}

type ComicService struct {
	comicRepo   *repositories.ComicRepository
	roleRepo    *repositories.RoleRepository
	sectionRepo *repositories.SectionRepository
	pageRepo    *repositories.PageRepository
	storage     *storage.Storage
	aigc        *gnxaigc.GnxAIGC
}

func NewComicService(
	comicRepo *repositories.ComicRepository,
	roleRepo *repositories.RoleRepository,
	sectionRepo *repositories.SectionRepository,
	pageRepo *repositories.PageRepository,
	storage *storage.Storage,
	aigc *gnxaigc.GnxAIGC,
) *ComicService {
	return &ComicService{
		comicRepo:   comicRepo,
		roleRepo:    roleRepo,
		sectionRepo: sectionRepo,
		pageRepo:    pageRepo,
		storage:     storage,
		aigc:        aigc,
	}
}

func (s *ComicService) CreateComic(ctx context.Context, title, userPrompt string, file io.Reader) (*models.Comic, error) {
	logger.Info("[Comic Creation] Starting comic creation: title=%s", title)

	comic := &models.Comic{
		Title:      title,
		UserPrompt: userPrompt,
		Status:     "pending",
	}

	if err := s.comicRepo.Create(comic); err != nil {
		logger.Error("[Comic Creation] Failed to create comic: %v", err)
		return nil, fmt.Errorf("failed to create comic: %w", err)
	}
	logger.Info("[Comic Creation] Comic created with ID=%d, status=pending", comic.ID)

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	logger.Info("[Comic Creation] Processing novel content for comic ID=%d", comic.ID)
	if err := s.processComicSync(ctx, comic.ID, string(content)); err != nil {
		logger.Error("[Comic Creation] Failed to process comic: %v", err)
		return nil, fmt.Errorf("failed to process comic: %w", err)
	}

	logger.Info("[Comic Creation] Comic ID=%d created successfully", comic.ID)
	return comic, nil
}

func (s *ComicService) GetComicList(page, limit int, status string) ([]models.Comic, int64, error) {
	return s.comicRepo.List(page, limit, status)
}

func (s *ComicService) GetComicDetail(id uint) (*models.Comic, error) {
	ret, err := s.comicRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	// 对 sections 按 index 排序
	slices.SortFunc(ret.Sections, func(a, b models.ComicSection) int {
		return a.Index - b.Index
	})
	return ret, nil
}

type novelChapter struct {
	Title   string
	Content string
}

func splitChaptersFromText(raw string) []novelChapter {
	lines := strings.Split(raw, "\n")
	var chapters []novelChapter
	var currentTitle string
	var buffer []string

	flush := func() {
		if currentTitle == "" && len(buffer) == 0 {
			return
		}
		content := strings.Trim(strings.Join(buffer, "\n"), "\n")
		chapters = append(chapters, novelChapter{Title: currentTitle, Content: content})
	}

	for _, line := range lines {
		trimmed := normalizeHeadingCandidate(line)
		if chapterHeadingPattern.MatchString(trimmed) {
			flush()
			currentTitle = trimmed
			buffer = buffer[:0]
			continue
		}
		if currentTitle == "" && len(strings.TrimSpace(line)) == 0 {
			continue
		}
		buffer = append(buffer, line)
	}

	flush()
	return chapters
}

// normalizeHeadingCandidate strips zero-width characters that often wrap headings in web-crawled novels.
func normalizeHeadingCandidate(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return trimmed
	}

	var b strings.Builder
	b.Grow(len(trimmed))
	for _, r := range trimmed {
		switch r {
		case '\u200B', '\u200C', '\u200D', '\uFEFF':
			continue
		default:
			b.WriteRune(r)
		}
	}

	return b.String()
}

func (s *ComicService) processComicSync(ctx context.Context, comicID uint, content string) error {
	comic, err := s.comicRepo.FindByID(comicID)
	if err != nil {
		return fmt.Errorf("failed to get comic %d: %w", comicID, err)
	}

	chapters := splitChaptersFromText(content)
	logger.Info("[Comic Processing] Split novel into %d chapters for comic ID=%d", len(chapters), comicID)
	if len(chapters) == 0 {
		logger.Error("[Comic Processing] No chapters found in novel for comic ID=%d", comicID)
		s.updateComicStatus(comicID, "failed")
		return fmt.Errorf("no chapters found in the novel")
	}

	logger.Info("[Comic Processing] Synchronously inserting %d sections for comic ID=%d", len(chapters), comicID)
	for i, chapter := range chapters {
		title := chapter.Title
		if title == "" {
			title = fmt.Sprintf("第%d章", i+1)
		}

		section := &models.ComicSection{
			ComicID: comicID,
			Title:   title,
			Index:   i + 1,
			Content: chapter.Content,
			Status:  "pending",
		}

		if err := s.sectionRepo.Create(section); err != nil {
			logger.Error("[Comic Processing] Failed to create section %d (%s): %v", i+1, title, err)
			return fmt.Errorf("failed to create section %d: %w", i+1, err)
		}
		logger.Info("[Comic Processing] Created section %d/%d: %s (ID=%d)", i+1, len(chapters), title, section.ID)
	}

	comic.Status = "completed"
	if err := s.comicRepo.Update(comic); err != nil {
		logger.Error("[Comic Processing] Failed to update comic status: %v", err)
	} else {
		logger.Info("[Comic Processing] Comic ID=%d status updated to completed", comicID)
	}

	logger.Info("[Comic AI Processing] Starting AI processing for comic ID=%d", comicID)
	go s.processComicAI(context.Background(), comicID)

	return nil
}

func (s *ComicService) processComicAI(ctx context.Context, comicID uint) {
	logger.Info("[Comic AI Processing] Loading comic data for ID=%d", comicID)
	comic, err := s.comicRepo.FindByID(comicID)
	if err != nil {
		logger.Error("[Comic AI Processing] Failed to get comic %d: %v", comicID, err)
		return
	}

	logger.Info("[Comic AI Processing] Querying all sections from database for comic ID=%d", comicID)
	sections, err := s.sectionRepo.FindByComicID(comicID)
	if err != nil {
		logger.Error("[Comic AI Processing] Failed to get sections for comic %d: %v", comicID, err)
		return
	}

	if len(sections) == 0 {
		logger.Error("[Comic AI Processing] No sections found for comic ID=%d", comicID)
		return
	}
	logger.Info("[Comic AI Processing] Found %d sections for comic ID=%d", len(sections), comicID)

	logger.Info("[Comic AI Processing] Fetching available voice list for comic ID=%d", comicID)
	voices, err := s.aigc.GetVoiceList(ctx)
	if err != nil {
		logger.Error("[Comic AI Processing] Failed to get voice list for comic %d: %v", comicID, err)
		return
	}

	voiceItems := make([]gnxaigc.TTSVoiceItem, 0, len(voices))
	for _, v := range voices {
		voiceItems = append(voiceItems, gnxaigc.TTSVoiceItem{
			VoiceName: v.VoiceName,
			VoiceType: v.VoiceType,
		})
	}

	firstSection := sections[0]
	logger.Info("[Comic AI Processing] Generating AI summary for first section: %s", firstSection.Title)
	summary, err := s.aigc.SummaryChapter(ctx, gnxaigc.SummaryChapterInput{
		NovelTitle:           comic.Title,
		ChapterTitle:         firstSection.Title,
		Content:              firstSection.Content,
		AvailableVoiceStyles: voiceItems,
		CharacterFeatures:    []gnxaigc.CharacterFeature{},
		MaxPanelsPerPage:     4,
	})
	if err != nil {
		logger.Error("[Comic AI Processing] Failed to generate AI summary for comic %d: %v", comicID, err)
		return
	}
	logger.Info("[Comic AI Processing] AI summary generated: %d characters, %d pages", len(summary.CharacterFeatures), len(summary.StoryboardPages))

	logger.Info("[Comic AI Processing] Creating %d character roles for comic ID=%d", len(summary.CharacterFeatures), comicID)
	for _, charFeature := range summary.CharacterFeatures {
		role := &models.ComicRole{
			ComicID:   comicID,
			Name:      charFeature.Basic.Name,
			Brief:     charFeature.Comment,
			Gender:    charFeature.Basic.Gender,
			Age:       charFeature.Basic.Age,
			VoiceName: charFeature.TTS.VoiceName,
			VoiceType: charFeature.TTS.VoiceType,
		}

		if err := s.roleRepo.Create(role); err != nil {
			logger.Error("[Comic AI Processing] Failed to create role %s: %v", charFeature.Basic.Name, err)
		} else {
			logger.Info("[Comic AI Processing] Created role: name=%s, gender=%s, age=%s", role.Name, role.Gender, role.Age)
		}
	}

	logger.Info("[Comic Image Processing] Starting image generation for comic ID=%d", comicID)
	s.processComicImages(ctx, comicID)

	logger.Info("[Comic AI Processing] Serially processing %d sections for comic ID=%d", len(sections), comicID)
	for i, section := range sections {
		logger.Info("[Comic AI Processing] Processing section %d/%d: %s (ID=%d)", i+1, len(sections), section.Title, section.ID)
		if err := s.processSectionSync(ctx, comic, &section); err != nil {
			logger.Error("[Comic AI Processing] Failed to process section %d (%s): %v", i+1, section.Title, err)
		} else {
			logger.Info("[Comic AI Processing] Section %d/%d processed successfully: %s", i+1, len(sections), section.Title)
		}
	}

	logger.Info("[Comic AI Processing] All sections processed for comic ID=%d", comicID)
}

func (s *ComicService) processComicImages(ctx context.Context, comicID uint) {
	logger.Info("[Comic Image Processing] Loading comic data for ID=%d", comicID)
	comic, err := s.comicRepo.FindByID(comicID)
	if err != nil {
		logger.Error("[Comic Image Processing] Failed to get comic %d: %v", comicID, err)
		return
	}

	roles, err := s.roleRepo.FindByComicID(comicID)
	if err != nil {
		logger.Error("[Comic Image Processing] Failed to get roles for comic %d: %v", comicID, err)
		return
	}
	logger.Info("[Comic Image Processing] Generating concept art for %d characters", len(roles))

	for _, role := range roles {
		if role.ImageID != "" {
			logger.Info("[Comic Image Processing] Character %s already has image, skipping", role.Name)
			continue
		}

		conceptArtPrompt := fmt.Sprintf("Character concept art for %s: %s", role.Name, role.Brief)
		logger.Info("[Comic Image Processing] Generating concept art for character: %s", role.Name)
		imageData, err := s.aigc.GenerateImageByText(ctx, conceptArtPrompt)
		if err != nil {
			logger.Error("[Comic Image Processing] Failed to generate role image for %s: %v", role.Name, err)
			continue
		}

		imageID := uuid.New().String()
		if err := s.storage.UploadBytes(imageData, imageID); err != nil {
			logger.Error("[Comic Image Processing] Failed to upload role image for %s: %v", role.Name, err)
			continue
		}

		role.ImageID = imageID
		if err := s.roleRepo.Update(&role); err != nil {
			logger.Error("[Comic Image Processing] Failed to update role image ID for %s: %v", role.Name, err)
		} else {
			logger.Info("[Comic Image Processing] Character %s concept art uploaded: imageID=%s", role.Name, imageID)
		}
	}

	if comic.IconImageID == "" {
		logger.Info("[Comic Image Processing] Generating cover image for comic ID=%d", comicID)
		iconImageData, err := s.aigc.GenerateImageByText(ctx, fmt.Sprintf("Comic book cover for: %s, %s", comic.Title, comic.UserPrompt))
		if err == nil {
			iconImageID := uuid.New().String()
			if err := s.storage.UploadBytes(iconImageData, iconImageID); err != nil {
				logger.Error("[Comic Image Processing] Failed to upload icon image: %v", err)
			} else {
				comic.IconImageID = iconImageID
				s.comicRepo.Update(comic)
				logger.Info("[Comic Image Processing] Cover image uploaded: imageID=%s", iconImageID)
			}
		} else {
			logger.Error("[Comic Image Processing] Failed to generate cover image: %v", err)
		}
	}

	if comic.BackgroundImageID == "" {
		logger.Info("[Comic Image Processing] Generating background image for comic ID=%d", comicID)
		bgImageData, err := s.aigc.GenerateImageByText(ctx, fmt.Sprintf("Comic background scene for: %s, %s", comic.Title, comic.UserPrompt))
		if err == nil {
			bgImageID := uuid.New().String()
			if err := s.storage.UploadBytes(bgImageData, bgImageID); err != nil {
				logger.Error("[Comic Image Processing] Failed to upload background image: %v", err)
			} else {
				comic.BackgroundImageID = bgImageID
				s.comicRepo.Update(comic)
				logger.Info("[Comic Image Processing] Background image uploaded: imageID=%s", bgImageID)
			}
		} else {
			logger.Error("[Comic Image Processing] Failed to generate background image: %v", err)
		}
	}
	logger.Info("[Comic Image Processing] Image processing completed for comic ID=%d", comicID)
}

func (s *ComicService) updateComicStatus(comicID uint, status string) {
	comic, err := s.comicRepo.FindByID(comicID)
	if err != nil {
		return
	}
	comic.Status = status
	s.comicRepo.Update(comic)
}

func (s *ComicService) CreateSection(ctx context.Context, comicID uint, title, content string) (*models.ComicSection, error) {
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

func (s *ComicService) GetSectionDetail(comicID, sectionID uint) (*models.ComicSection, error) {
	section, err := s.sectionRepo.FindByID(sectionID)
	if err != nil {
		return nil, fmt.Errorf("section not found: %w", err)
	}

	if section.ComicID != comicID {
		return nil, fmt.Errorf("section does not belong to comic")
	}

	// 对 pages 按 index 排序
	slices.SortFunc(section.Pages, func(a, b models.ComicPage) int {
		return a.Index - b.Index
	})
	// 对每个 page 的 details 按 index 排序
	for i := range section.Pages {
		slices.SortFunc(section.Pages[i].Details, func(a, b models.ComicPageDetail) int {
			return a.Index - b.Index
		})
	}
	return section, nil
}

func (s *ComicService) processSectionSync(ctx context.Context, comic *models.Comic, section *models.ComicSection) error {
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

func (s *ComicService) processSectionImages(ctx context.Context, comic *models.Comic, sectionID uint) {
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
	characterAssets, err := s.SyncCharacterAssets(ctx, comic.ID, comic.UserPrompt, summary.CharacterFeatures)
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

func (s *ComicService) collectPageCharacterKeys(page gnxaigc.StoryboardPage, features []gnxaigc.CharacterFeature) []string {
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

func (s *ComicService) updateSectionStatus(sectionID uint, status string) {
	section, err := s.sectionRepo.FindByID(sectionID)
	if err != nil {
		return
	}
	section.Status = status
	s.sectionRepo.Update(section)
}

func (s *ComicService) SyncCharacterAssets(
	ctx context.Context,
	comicID uint,
	imageStyle string,
	features []gnxaigc.CharacterFeature,
) (map[string]*CharacterAsset, error) {
	roles, err := s.roleRepo.FindByComicID(comicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	roleMap := make(map[string]*models.ComicRole)
	for i := range roles {
		roleMap[roles[i].Name] = &roles[i]
	}

	assets := make(map[string]*CharacterAsset)
	trimmedStyle := strings.TrimSpace(imageStyle)

	logger.Info("[Character Assets] Syncing assets for %d characters", len(features))
	for _, feature := range features {
		name := strings.TrimSpace(feature.Basic.Name)
		if name == "" {
			logger.Warn("[Character Assets] Skipping character with empty name")
			continue
		}

		role, exists := roleMap[name]
		if !exists {
			logger.Warn("[Character Assets] Character %s not found in roles", name)
			continue
		}

		prompt := strings.TrimSpace(feature.ConceptArtPrompt)
		if prompt == "" {
			logger.Warn("[Character Assets] Character %s missing concept_art_prompt, skipping image generation", name)
			assets[name] = &CharacterAsset{
				Feature: feature,
			}
			continue
		}

		fullPrompt := prompt
		if trimmedStyle != "" {
			fullPrompt = fmt.Sprintf("%s %s", trimmedStyle, prompt)
		}
		fullPrompt = strings.TrimSpace(fullPrompt)

		var imageData []byte
		shouldGenerate := role.ImageID == ""

		if !shouldGenerate {
			existingImageData, err := s.storage.DownloadBytes(role.ImageID)
			if err != nil || len(existingImageData) == 0 {
				shouldGenerate = true
			} else {
				imageData = existingImageData
			}
		}

		if shouldGenerate {
			if role.ImageID != "" {
				baseData, err := s.storage.DownloadBytes(role.ImageID)
				if err != nil {
					logger.Warn("[Character Assets] Character %s: cannot read existing concept art (%v), using text generation", name, err)
					imageData, err = s.aigc.GenerateImageByText(ctx, fullPrompt)
					if err != nil {
						logger.Error("[Character Assets] Character %s: failed to generate concept art: %v", name, err)
						continue
					}
				} else {
					logger.Info("[Character Assets] Character %s: Refining concept art via img2img", name)
					imageData, err = s.aigc.GenerateImageByImage(ctx, baseData, fullPrompt)
					if err != nil {
						logger.Warn("[Character Assets] Character %s: img2img refinement failed (%v), falling back to text generation", name, err)
						imageData, err = s.aigc.GenerateImageByText(ctx, fullPrompt)
						if err != nil {
							logger.Error("[Character Assets] Character %s: failed to generate concept art: %v", name, err)
							continue
						}
					}
				}
			} else {
				logger.Info("[Character Assets] Character %s: Generating concept art from scratch", name)
				imageData, err = s.aigc.GenerateImageByText(ctx, fullPrompt)
				if err != nil {
					logger.Error("[Character Assets] Character %s: failed to generate concept art: %v", name, err)
					continue
				}
			}

			imageID := fmt.Sprintf("character_%d_%s", role.ID, name)
			if err := s.storage.UploadBytes(imageData, imageID); err != nil {
				logger.Error("[Character Assets] Character %s: failed to upload concept art: %v", name, err)
				continue
			}

			role.ImageID = imageID
			if err := s.roleRepo.Update(role); err != nil {
				logger.Error("[Character Assets] Character %s: failed to update role with imageID: %v", name, err)
			} else {
				logger.Info("[Character Assets] Character %s: Concept art uploaded (imageID=%s)", name, imageID)
			}
		} else {
			logger.Info("[Character Assets] Character %s: Reusing existing concept art", name)
		}

		assets[name] = &CharacterAsset{
			Feature:   feature,
			ImageData: imageData,
			Prompt:    fullPrompt,
		}
	}

	return assets, nil
}

func (s *ComicService) LoadCharacterAssets(
	ctx context.Context,
	comicID uint,
) (map[string]*CharacterAsset, error) {
	roles, err := s.roleRepo.FindByComicID(comicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	assets := make(map[string]*CharacterAsset)
	for i := range roles {
		role := &roles[i]
		if role.ImageID == "" {
			continue
		}

		imageData, err := s.storage.DownloadBytes(role.ImageID)
		if err != nil {
			logger.Warn("[Character Assets] Failed to download image for character %s: %v", role.Name, err)
			continue
		}

		assets[role.Name] = &CharacterAsset{
			ImageData: imageData,
		}
	}

	return assets, nil
}
