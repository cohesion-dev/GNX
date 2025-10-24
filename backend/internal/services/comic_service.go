package services

import (
	"fmt"
	"io"
	"io/ioutil"

	"gorm.io/gorm"

	"github.com/cohesion-dev/GNX/backend/internal/models"
	"github.com/cohesion-dev/GNX/backend/internal/repositories"
	"github.com/cohesion-dev/GNX/backend/pkg/ai"
	"github.com/cohesion-dev/GNX/backend/pkg/storage"
)

type ComicService struct {
	comicRepo      *repositories.ComicRepository
	roleRepo       *repositories.RoleRepository
	sectionRepo    *repositories.SectionRepository
	aiService      *ai.OpenAIClient
	storageService *storage.QiniuClient
	db             *gorm.DB
}

func NewComicService(
	comicRepo *repositories.ComicRepository,
	roleRepo *repositories.RoleRepository,
	sectionRepo *repositories.SectionRepository,
	aiService *ai.OpenAIClient,
	storageService *storage.QiniuClient,
	db *gorm.DB,
) *ComicService {
	return &ComicService{
		comicRepo:      comicRepo,
		roleRepo:       roleRepo,
		sectionRepo:    sectionRepo,
		aiService:      aiService,
		storageService: storageService,
		db:             db,
	}
}

func (s *ComicService) GetComicList(page, limit int, status string) ([]models.Comic, int64, error) {
	offset := (page - 1) * limit
	return s.comicRepo.GetListWithFilter(limit, offset, status)
}

func (s *ComicService) CreateComic(title, userPrompt string, fileContent io.Reader) (*models.Comic, error) {
	content, err := ioutil.ReadAll(fileContent)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	comic := &models.Comic{
		Title:      title,
		UserPrompt: userPrompt,
		Status:     "pending",
	}

	if err := s.comicRepo.Create(comic); err != nil {
		return nil, fmt.Errorf("failed to create comic: %w", err)
	}

	go s.processComicGeneration(comic.ID, string(content), userPrompt)

	return comic, nil
}

func (s *ComicService) GetComicDetail(id uint) (*models.Comic, error) {
	return s.comicRepo.GetByIDWithRelations(id)
}

func (s *ComicService) GetComicSections(id uint) ([]models.ComicSection, error) {
	return s.sectionRepo.GetByComicID(id)
}

func (s *ComicService) processComicGeneration(comicID uint, content, userPrompt string) {
	defer func() {
		if r := recover(); r != nil {
			s.comicRepo.UpdateStatus(comicID, "failed")
		}
	}()

	analysis, err := s.aiService.AnalyzeNovel(content, userPrompt)
	if err != nil {
		s.comicRepo.UpdateStatus(comicID, "failed")
		return
	}

	if err := s.comicRepo.UpdateBrief(comicID, analysis.Brief); err != nil {
		s.comicRepo.UpdateStatus(comicID, "failed")
		return
	}

	for _, roleInfo := range analysis.Roles {
		role := &models.ComicRole{
			ComicID: comicID,
			Name:    roleInfo.Name,
			Brief:   roleInfo.Brief,
			Voice:   roleInfo.Voice,
		}
		if err := s.roleRepo.Create(role); err != nil {
			continue
		}

		go s.generateRoleImage(role.ID, roleInfo.ImagePrompt, userPrompt)
	}

	go s.generateComicImages(comicID, analysis.IconPrompt, analysis.BgPrompt, userPrompt)

	if err := s.comicRepo.UpdateStatus(comicID, "completed"); err != nil {
		return
	}
}

func (s *ComicService) generateRoleImage(roleID uint, imagePrompt, userPrompt string) {
	imageData, err := s.aiService.GenerateImage(imagePrompt, userPrompt)
	if err != nil {
		return
	}

	imageURL, err := s.storageService.UploadImage(fmt.Sprintf("roles/%d.png", roleID), imageData)
	if err != nil {
		return
	}

	s.roleRepo.UpdateImageURL(roleID, imageURL)
}

func (s *ComicService) generateComicImages(comicID uint, iconPrompt, bgPrompt, userPrompt string) {
	iconData, err := s.aiService.GenerateImage(iconPrompt, userPrompt)
	if err == nil {
		iconURL, err := s.storageService.UploadImage(fmt.Sprintf("comics/%d/icon.png", comicID), iconData)
		if err == nil {
			s.comicRepo.UpdateIcon(comicID, iconURL)
		}
	}

	bgData, err := s.aiService.GenerateImage(bgPrompt, userPrompt)
	if err == nil {
		bgURL, err := s.storageService.UploadImage(fmt.Sprintf("comics/%d/bg.png", comicID), bgData)
		if err == nil {
			s.comicRepo.UpdateBg(comicID, bgURL)
		}
	}
}
