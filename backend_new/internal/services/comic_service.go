package services

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/cohesion-dev/GNX/ai/gnxaigc"
	"github.com/cohesion-dev/GNX/backend_new/internal/models"
	"github.com/cohesion-dev/GNX/backend_new/internal/repositories"
	"github.com/cohesion-dev/GNX/backend_new/pkg/logger"
	"github.com/cohesion-dev/GNX/backend_new/pkg/storage"
	"github.com/google/uuid"
)

var chapterHeadingPattern = regexp.MustCompile(`^第[零〇一二三四五六七八九十百千万0-9]+章.*$`)

type ComicService struct {
	comicRepo      *repositories.ComicRepository
	roleRepo       *repositories.RoleRepository
	sectionRepo    *repositories.SectionRepository
	pageRepo       *repositories.PageRepository
	storage        *storage.Storage
	aigc           *gnxaigc.GnxAIGC
	sectionService *SectionService
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

func (s *ComicService) SetSectionService(sectionService *SectionService) {
	s.sectionService = sectionService
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
	return s.comicRepo.FindByID(id)
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
		trimmed := strings.TrimSpace(line)
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

	comic.Status = "completed"
	if err := s.comicRepo.Update(comic); err != nil {
		logger.Error("[Comic Processing] Failed to update comic status: %v", err)
	} else {
		logger.Info("[Comic Processing] Comic ID=%d status updated to completed", comicID)
	}

	logger.Info("[Comic AI Processing] Starting AI processing for comic ID=%d", comicID)
	go s.processComicAI(context.Background(), comicID, chapters)

	return nil
}

func (s *ComicService) processComicAI(ctx context.Context, comicID uint, chapters []novelChapter) {
	logger.Info("[Comic AI Processing] Loading comic data for ID=%d", comicID)
	comic, err := s.comicRepo.FindByID(comicID)
	if err != nil {
		logger.Error("[Comic AI Processing] Failed to get comic %d: %v", comicID, err)
		return
	}

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

	if len(chapters) == 0 {
		logger.Error("[Comic AI Processing] No chapters to process for comic ID=%d", comicID)
		return
	}

	firstChapter := chapters[0]
	title := firstChapter.Title
	if title == "" {
		title = "第1章"
	}

	logger.Info("[Comic AI Processing] Generating AI summary for first chapter: %s", title)
	summary, err := s.aigc.SummaryChapter(ctx, gnxaigc.SummaryChapterInput{
		NovelTitle:           comic.Title,
		ChapterTitle:         title,
		Content:              firstChapter.Content,
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
			logger.Info("[Comic AI Processing] Created role: name=%s, gender=%s, age=%d", role.Name, role.Gender, role.Age)
		}
	}

	logger.Info("[Comic Image Processing] Starting image generation for comic ID=%d", comicID)
	s.processComicImages(ctx, comicID)

	logger.Info("[Comic AI Processing] Serially creating %d sections for comic ID=%d", len(chapters), comicID)
	for i, chapter := range chapters {
		title := chapter.Title
		if title == "" {
			title = fmt.Sprintf("第%d章", i+1)
		}

		logger.Info("[Comic AI Processing] Processing section %d/%d: %s", i+1, len(chapters), title)
		if s.sectionService != nil {
			_, err := s.sectionService.CreateSection(ctx, comicID, title, chapter.Content)
			if err != nil {
				logger.Error("[Comic AI Processing] Failed to create section %d (%s): %v", i+1, title, err)
			} else {
				logger.Info("[Comic AI Processing] Section %d/%d created successfully: %s", i+1, len(chapters), title)
			}
		} else {
			logger.Error("[Comic AI Processing] SectionService not set, cannot create section %d", i+1)
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
