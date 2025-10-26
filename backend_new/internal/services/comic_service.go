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
	"github.com/cohesion-dev/GNX/backend_new/pkg/storage"
	"github.com/google/uuid"
)

var chapterHeadingPattern = regexp.MustCompile(`^第[零〇一二三四五六七八九十百千万0-9]+章.*$`)

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
	comic := &models.Comic{
		Title:      title,
		UserPrompt: userPrompt,
		Status:     "pending",
	}

	if err := s.comicRepo.Create(comic); err != nil {
		return nil, fmt.Errorf("failed to create comic: %w", err)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if err := s.processComicSync(ctx, comic.ID, string(content)); err != nil {
		return nil, fmt.Errorf("failed to process comic: %w", err)
	}

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
	if len(chapters) == 0 {
		s.updateComicStatus(comicID, "failed")
		return fmt.Errorf("no chapters found in the novel")
	}

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
			fmt.Printf("Failed to create section %d: %v\n", i+1, err)
			continue
		}
	}

	comic.Status = "completed"
	if err := s.comicRepo.Update(comic); err != nil {
		fmt.Printf("Failed to update comic status: %v\n", err)
	}

	go s.processComicAI(context.Background(), comicID)

	return nil
}

func (s *ComicService) processComicAI(ctx context.Context, comicID uint) {
	comic, err := s.comicRepo.FindByID(comicID)
	if err != nil {
		fmt.Printf("Failed to get comic %d for AI processing: %v\n", comicID, err)
		return
	}

	sections, err := s.sectionRepo.FindByComicID(comicID)
	if err != nil || len(sections) == 0 {
		fmt.Printf("Failed to get sections for comic %d: %v\n", comicID, err)
		return
	}

	voices, err := s.aigc.GetVoiceList(ctx)
	if err != nil {
		fmt.Printf("Failed to get voice list for comic %d: %v\n", comicID, err)
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
	summary, err := s.aigc.SummaryChapter(ctx, gnxaigc.SummaryChapterInput{
		NovelTitle:           comic.Title,
		ChapterTitle:         firstSection.Title,
		Content:              firstSection.Content,
		AvailableVoiceStyles: voiceItems,
		CharacterFeatures:    []gnxaigc.CharacterFeature{},
		MaxPanelsPerPage:     4,
	})
	if err != nil {
		fmt.Printf("Failed to generate AI summary for comic %d: %v\n", comicID, err)
		return
	}

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
			fmt.Printf("Failed to create role: %v\n", err)
		}
	}

	roles, err := s.roleRepo.FindByComicID(comicID)
	if err != nil {
		fmt.Printf("Failed to get roles for first section processing: %v\n", err)
		roles = []models.ComicRole{}
	}

	for pageIndex, storyboardPage := range summary.StoryboardPages {
		page := &models.ComicPage{
			SectionID:   firstSection.ID,
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

	firstSection.Status = "completed"
	if err := s.sectionRepo.Update(&firstSection); err != nil {
		fmt.Printf("Failed to update first section status: %v\n", err)
	}

	s.processComicImages(ctx, comicID)
}

func (s *ComicService) processComicImages(ctx context.Context, comicID uint) {
	comic, err := s.comicRepo.FindByID(comicID)
	if err != nil {
		fmt.Printf("Failed to get comic %d for image processing: %v\n", comicID, err)
		return
	}

	roles, err := s.roleRepo.FindByComicID(comicID)
	if err != nil {
		fmt.Printf("Failed to get roles for comic %d: %v\n", comicID, err)
		return
	}

	for _, role := range roles {
		if role.ImageID != "" {
			continue
		}

		conceptArtPrompt := fmt.Sprintf("Character concept art for %s: %s", role.Name, role.Brief)
		imageData, err := s.aigc.GenerateImageByText(ctx, conceptArtPrompt)
		if err != nil {
			fmt.Printf("Failed to generate role image for %s: %v\n", role.Name, err)
			continue
		}

		imageID := uuid.New().String()
		if err := s.storage.UploadBytes(imageData, imageID); err != nil {
			fmt.Printf("Failed to upload role image for %s: %v\n", role.Name, err)
			continue
		}

		role.ImageID = imageID
		if err := s.roleRepo.Update(&role); err != nil {
			fmt.Printf("Failed to update role image ID for %s: %v\n", role.Name, err)
		}
	}

	if comic.IconImageID == "" {
		iconImageData, err := s.aigc.GenerateImageByText(ctx, fmt.Sprintf("Comic book cover for: %s, %s", comic.Title, comic.UserPrompt))
		if err == nil {
			iconImageID := uuid.New().String()
			if err := s.storage.UploadBytes(iconImageData, iconImageID); err != nil {
				fmt.Printf("Failed to upload icon image: %v\n", err)
			} else {
				comic.IconImageID = iconImageID
				s.comicRepo.Update(comic)
			}
		}
	}

	if comic.BackgroundImageID == "" {
		bgImageData, err := s.aigc.GenerateImageByText(ctx, fmt.Sprintf("Comic background scene for: %s, %s", comic.Title, comic.UserPrompt))
		if err == nil {
			bgImageID := uuid.New().String()
			if err := s.storage.UploadBytes(bgImageData, bgImageID); err != nil {
				fmt.Printf("Failed to upload background image: %v\n", err)
			} else {
				comic.BackgroundImageID = bgImageID
				s.comicRepo.Update(comic)
			}
		}
	}
}

func (s *ComicService) updateComicStatus(comicID uint, status string) {
	comic, err := s.comicRepo.FindByID(comicID)
	if err != nil {
		return
	}
	comic.Status = status
	s.comicRepo.Update(comic)
}
