package services

import (
	"context"
	"fmt"
	"io"

	"github.com/cohesion-dev/GNX/ai/gnxaigc"
	"github.com/cohesion-dev/GNX/backend_new/internal/models"
	"github.com/cohesion-dev/GNX/backend_new/internal/repositories"
	"github.com/cohesion-dev/GNX/backend_new/pkg/storage"
	"github.com/google/uuid"
)

type ComicService struct {
	comicRepo   *repositories.ComicRepository
	roleRepo    *repositories.RoleRepository
	sectionRepo *repositories.SectionRepository
	storage     *storage.Storage
	aigc        *gnxaigc.GnxAIGC
}

func NewComicService(
	comicRepo *repositories.ComicRepository,
	roleRepo *repositories.RoleRepository,
	sectionRepo *repositories.SectionRepository,
	storage *storage.Storage,
	aigc *gnxaigc.GnxAIGC,
) *ComicService {
	return &ComicService{
		comicRepo:   comicRepo,
		roleRepo:    roleRepo,
		sectionRepo: sectionRepo,
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

	go s.processComic(context.Background(), comic.ID, string(content))

	return comic, nil
}

func (s *ComicService) GetComicList(page, limit int, status string) ([]models.Comic, int64, error) {
	return s.comicRepo.List(page, limit, status)
}

func (s *ComicService) GetComicDetail(id uint) (*models.Comic, error) {
	return s.comicRepo.FindByID(id)
}

func (s *ComicService) processComic(ctx context.Context, comicID uint, content string) {
	comic, err := s.comicRepo.FindByID(comicID)
	if err != nil {
		fmt.Printf("Failed to get comic %d: %v\n", comicID, err)
		return
	}

	voices, err := s.aigc.GetVoiceList(ctx)
	if err != nil {
		fmt.Printf("Failed to get voice list for comic %d: %v\n", comicID, err)
		s.updateComicStatus(comicID, "failed")
		return
	}

	voiceItems := make([]gnxaigc.TTSVoiceItem, 0, len(voices))
	for _, v := range voices {
		voiceItems = append(voiceItems, gnxaigc.TTSVoiceItem{
			VoiceName: v.VoiceName,
			VoiceType: v.VoiceType,
		})
	}

	summary, err := s.aigc.SummaryChapter(ctx, gnxaigc.SummaryChapterInput{
		NovelTitle:           comic.Title,
		ChapterTitle:         "第一章",
		Content:              content,
		AvailableVoiceStyles: voiceItems,
		CharacterFeatures:    []gnxaigc.CharacterFeature{},
		MaxPanelsPerPage:     4,
	})
	if err != nil {
		fmt.Printf("Failed to generate summary for comic %d: %v\n", comicID, err)
		s.updateComicStatus(comicID, "failed")
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

		if charFeature.ConceptArtPrompt != "" {
			imageData, err := s.aigc.GenerateImageByText(ctx, charFeature.ConceptArtPrompt)
			if err != nil {
				fmt.Printf("Failed to generate role image: %v\n", err)
			} else {
				imageID := uuid.New().String()
				if err := s.storage.UploadBytes(imageData, imageID); err != nil {
					fmt.Printf("Failed to upload role image: %v\n", err)
				} else {
					role.ImageID = imageID
				}
			}
		}

		if err := s.roleRepo.Create(role); err != nil {
			fmt.Printf("Failed to create role: %v\n", err)
		}
	}

	section := &models.ComicSection{
		ComicID: comicID,
		Title:   "第一章",
		Index:   1,
		Content: content,
		Status:  "pending",
	}
	if err := s.sectionRepo.Create(section); err != nil {
		fmt.Printf("Failed to create section: %v\n", err)
		s.updateComicStatus(comicID, "failed")
		return
	}

	iconImageData, err := s.aigc.GenerateImageByText(ctx, fmt.Sprintf("Comic book cover for: %s, %s", comic.Title, comic.UserPrompt))
	if err == nil {
		iconImageID := uuid.New().String()
		if err := s.storage.UploadBytes(iconImageData, iconImageID); err != nil {
			fmt.Printf("Failed to upload icon image: %v\n", err)
		} else {
			comic.IconImageID = iconImageID
		}
	}

	bgImageData, err := s.aigc.GenerateImageByText(ctx, fmt.Sprintf("Comic background scene for: %s, %s", comic.Title, comic.UserPrompt))
	if err == nil {
		bgImageID := uuid.New().String()
		if err := s.storage.UploadBytes(bgImageData, bgImageID); err != nil {
			fmt.Printf("Failed to upload background image: %v\n", err)
		} else {
			comic.BackgroundImageID = bgImageID
		}
	}

	comic.Status = "completed"
	if err := s.comicRepo.Update(comic); err != nil {
		fmt.Printf("Failed to update comic status: %v\n", err)
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
