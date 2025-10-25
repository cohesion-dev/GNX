package repositories

import (
	"gorm.io/gorm"

	"github.com/cohesion-dev/GNX/backend/internal/models"
)

type StoryboardRepository struct {
	db *gorm.DB
}

func NewStoryboardRepository(db *gorm.DB) *StoryboardRepository {
	return &StoryboardRepository{
		db: db,
	}
}

func (r *StoryboardRepository) Create(storyboard *models.ComicStoryboard) error {
	return r.db.Create(storyboard).Error
}

func (r *StoryboardRepository) CreateDetail(detail *models.ComicStoryboardDetail) error {
	return r.db.Create(detail).Error
}

func (r *StoryboardRepository) GetBySectionID(sectionID uint) ([]models.ComicStoryboard, error) {
	var storyboards []models.ComicStoryboard
	err := r.db.Where("section_id = ?", sectionID).Find(&storyboards).Error
	return storyboards, err
}

func (r *StoryboardRepository) GetBySectionIDWithDetails(sectionID uint) ([]models.ComicStoryboard, error) {
	var storyboards []models.ComicStoryboard
	err := r.db.Preload("Details.Role").Where("section_id = ?", sectionID).Find(&storyboards).Error
	return storyboards, err
}

func (r *StoryboardRepository) GetDetailByID(id uint) (*models.ComicStoryboardDetail, error) {
	var detail models.ComicStoryboardDetail
	err := r.db.Preload("Role").First(&detail, id).Error
	return &detail, err
}

func (r *StoryboardRepository) UpdateImageURL(id uint, imageURL string) error {
	return r.db.Model(&models.ComicStoryboard{}).Where("id = ?", id).Update("image_url", imageURL).Error
}

func (r *StoryboardRepository) UpdateDetailTTSURL(id uint, ttsURL string) error {
	return r.db.Model(&models.ComicStoryboardDetail{}).Where("id = ?", id).Update("tts_url", ttsURL).Error
}
