package services

import (
	"context"
	"fmt"

	"github.com/cohesion-dev/GNX/ai/gnxaigc"
	"github.com/cohesion-dev/GNX/backend_new/internal/models"
	"github.com/cohesion-dev/GNX/backend_new/internal/repositories"
	"github.com/cohesion-dev/GNX/backend_new/pkg/storage"
	"github.com/google/uuid"
)

type SectionService struct {
	comicRepo   *repositories.ComicRepository
	roleRepo    *repositories.RoleRepository
	sectionRepo *repositories.SectionRepository
	pageRepo    *repositories.PageRepository
	storage     *storage.Storage
	aigc        *gnxaigc.GnxAIGC
}

func NewSectionService(
	comicRepo *repositories.ComicRepository,
	roleRepo *repositories.RoleRepository,
	sectionRepo *repositories.SectionRepository,
	pageRepo *repositories.PageRepository,
	storage *storage.Storage,
	aigc *gnxaigc.GnxAIGC,
) *SectionService {
	return &SectionService{
		comicRepo:   comicRepo,
		roleRepo:    roleRepo,
		sectionRepo: sectionRepo,
		pageRepo:    pageRepo,
		storage:     storage,
		aigc:        aigc,
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

	go s.processSection(context.Background(), comic, section)

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

func (s *SectionService) processSection(ctx context.Context, comic *models.Comic, section *models.ComicSection) {
	roles, err := s.roleRepo.FindByComicID(comic.ID)
	if err != nil {
		fmt.Printf("Failed to get roles for section %d: %v\n", section.ID, err)
		s.updateSectionStatus(section.ID, "failed")
		return
	}

	voices, err := s.aigc.GetVoiceList(ctx)
	if err != nil {
		fmt.Printf("Failed to get voice list for section %d: %v\n", section.ID, err)
		s.updateSectionStatus(section.ID, "failed")
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
		fmt.Printf("Failed to generate summary for section %d: %v\n", section.ID, err)
		s.updateSectionStatus(section.ID, "failed")
		return
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

		fullPrompt := gnxaigc.ComposePageImagePrompt(comic.UserPrompt, storyboardPage)
		imageData, err := s.aigc.GenerateImageByText(ctx, fullPrompt)
		if err != nil {
			fmt.Printf("Failed to generate page image: %v\n", err)
		} else {
			imageID := uuid.New().String()
			if err := s.storage.UploadBytes(imageData, imageID); err != nil {
				fmt.Printf("Failed to upload page image: %v\n", err)
			}
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
}

func (s *SectionService) updateSectionStatus(sectionID uint, status string) {
	section, err := s.sectionRepo.FindByID(sectionID)
	if err != nil {
		return
	}
	section.Status = status
	s.sectionRepo.Update(section)
}
