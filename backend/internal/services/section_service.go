package services

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/cohesion-dev/GNX/backend/internal/models"
	"github.com/cohesion-dev/GNX/backend/internal/repositories"
	"github.com/cohesion-dev/GNX/backend/pkg/storage"
)

type SectionService struct {
	sectionRepo    *repositories.SectionRepository
	storyboardRepo *repositories.StoryboardRepository
	roleRepo       *repositories.RoleRepository
	comicRepo      *repositories.ComicRepository
	storageService *storage.QiniuClient
	db             *gorm.DB
}

func NewSectionService(
	sectionRepo *repositories.SectionRepository,
	storyboardRepo *repositories.StoryboardRepository,
	roleRepo *repositories.RoleRepository,
	comicRepo *repositories.ComicRepository,
	storageService *storage.QiniuClient,
	db *gorm.DB,
) *SectionService {
	return &SectionService{
		sectionRepo:    sectionRepo,
		storyboardRepo: storyboardRepo,
		roleRepo:       roleRepo,
		comicRepo:      comicRepo,
		storageService: storageService,
		db:             db,
	}
}

func (s *SectionService) CreateSection(comicID uint, index int, detail string) (*models.ComicSection, error) {
	_, err := s.comicRepo.GetByID(comicID)
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

	pages, err := s.storyboardRepo.GetPagesBySectionID(sectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pages: %w", err)
	}

	return map[string]interface{}{
		"section": section,
		"pages":   pages,
	}, nil
}

func (s *SectionService) GetStoryboards(comicID, sectionID uint) ([]models.ComicStoryboardPage, error) {
	section, err := s.sectionRepo.GetByID(sectionID)
	if err != nil {
		return nil, fmt.Errorf("section not found: %w", err)
	}

	if section.ComicID != comicID {
		return nil, fmt.Errorf("section does not belong to comic")
	}

	return s.storyboardRepo.GetPagesBySectionID(sectionID)
}
