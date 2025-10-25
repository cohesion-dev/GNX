package services

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/cohesion-dev/GNX/backend/internal/models"
	"github.com/cohesion-dev/GNX/backend/internal/repositories"
	"github.com/cohesion-dev/GNX/backend/pkg/ai"
	"github.com/cohesion-dev/GNX/backend/pkg/storage"
)

type SectionService struct {
	sectionRepo    *repositories.SectionRepository
	storyboardRepo *repositories.StoryboardRepository
	roleRepo       *repositories.RoleRepository
	comicRepo      *repositories.ComicRepository
	aiService      *ai.OpenAIClient
	storageService *storage.QiniuClient
	db             *gorm.DB
}

func NewSectionService(
	sectionRepo *repositories.SectionRepository,
	storyboardRepo *repositories.StoryboardRepository,
	roleRepo *repositories.RoleRepository,
	comicRepo *repositories.ComicRepository,
	aiService *ai.OpenAIClient,
	storageService *storage.QiniuClient,
	db *gorm.DB,
) *SectionService {
	return &SectionService{
		sectionRepo:    sectionRepo,
		storyboardRepo: storyboardRepo,
		roleRepo:       roleRepo,
		comicRepo:      comicRepo,
		aiService:      aiService,
		storageService: storageService,
		db:             db,
	}
}

func (s *SectionService) CreateSection(comicID uint, index int, detail string) (*models.ComicSection, error) {
	comic, err := s.comicRepo.GetByID(comicID)
	if err != nil {
		return nil, fmt.Errorf("comic not found: %w", err)
	}

	section := &models.ComicSection{
		ComicID: comicID,
		Index:   index,
		Detail:  detail,
		Status:  "pending",
	}

	if err := s.sectionRepo.Create(section); err != nil {
		return nil, fmt.Errorf("failed to create section: %w", err)
	}

	go s.processSectionGeneration(section.ID, comicID, detail, comic.UserPrompt)

	return section, nil
}

func (s *SectionService) GetSectionContent(comicID, sectionID uint) (map[string]interface{}, error) {
	section, err := s.sectionRepo.GetByID(sectionID)
	if err != nil {
		return nil, fmt.Errorf("section not found: %w", err)
	}

	if section.ComicID != comicID {
		return nil, fmt.Errorf("section does not belong to comic")
	}

	storyboards, err := s.storyboardRepo.GetBySectionIDWithDetails(sectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get storyboards: %w", err)
	}

	return map[string]interface{}{
		"section":     section,
		"storyboards": storyboards,
	}, nil
}

func (s *SectionService) GetStoryboards(comicID, sectionID uint) ([]models.ComicStoryboard, error) {
	section, err := s.sectionRepo.GetByID(sectionID)
	if err != nil {
		return nil, fmt.Errorf("section not found: %w", err)
	}

	if section.ComicID != comicID {
		return nil, fmt.Errorf("section does not belong to comic")
	}

	return s.storyboardRepo.GetBySectionIDWithDetails(sectionID)
}

func (s *SectionService) processSectionGeneration(sectionID, comicID uint, content, userPrompt string) {
	defer func() {
		if r := recover(); r != nil {
			s.sectionRepo.UpdateStatus(sectionID, "failed")
		}
	}()

	roles, err := s.roleRepo.GetByComicID(comicID)
	if err != nil {
		s.sectionRepo.UpdateStatus(sectionID, "failed")
		return
	}

	storyboards, err := s.aiService.GenerateStoryboards(content, roles, userPrompt)
	if err != nil {
		s.sectionRepo.UpdateStatus(sectionID, "failed")
		return
	}

	for _, sbInfo := range storyboards {
		storyboard := &models.ComicStoryboard{
			SectionID:   sectionID,
			ImagePrompt: sbInfo.ImagePrompt,
		}

		if err := s.storyboardRepo.Create(storyboard); err != nil {
			continue
		}

		for idx, detailInfo := range sbInfo.Details {
			detail := &models.ComicStoryboardDetail{
				StoryboardID: storyboard.ID,
				Index:        idx + 1,
				Text:         detailInfo.Text,
				VoiceName:    "",
				VoiceType:    "",
				SpeedRatio:   1.0,
			}

			if detailInfo.RoleName != "" {
				for _, role := range roles {
					if role.Name == detailInfo.RoleName {
						detail.VoiceName = role.VoiceName
						detail.VoiceType = role.VoiceType
						detail.SpeedRatio = role.SpeedRatio
						break
					}
				}
			}

			if err := s.storyboardRepo.CreateDetail(detail); err != nil {
				continue
			}

			if detail.VoiceType != "" {
				go s.generateTTS(detail.ID, detailInfo.Text, detail.VoiceType)
			}
		}

		go s.generateStoryboardImage(storyboard.ID, sbInfo.ImagePrompt, userPrompt)
	}

	if err := s.sectionRepo.UpdateStatus(sectionID, "completed"); err != nil {
		return
	}
}

func (s *SectionService) generateStoryboardImage(storyboardID uint, imagePrompt, userPrompt string) {
	imageData, err := s.aiService.GenerateImage(imagePrompt, userPrompt)
	if err != nil {
		return
	}

	imageURL, err := s.storageService.UploadImage(fmt.Sprintf("storyboards/%d.png", storyboardID), imageData)
	if err != nil {
		return
	}

	s.storyboardRepo.UpdateImageURL(storyboardID, imageURL)
}

func (s *SectionService) generateTTS(detailID uint, text, voice string) {
	audioData, err := s.storageService.GenerateTTS(text, voice)
	if err != nil {
		return
	}

	audioURL, err := s.storageService.UploadAudio(fmt.Sprintf("tts/%d.mp3", detailID), audioData)
	if err != nil {
		return
	}

	s.storyboardRepo.UpdateDetailTTSURL(detailID, audioURL)
}
