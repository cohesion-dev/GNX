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
	err := r.db.First(&detail, id).Error
	return &detail, err
}

func (r *StoryboardRepository) UpdateImageURL(id uint, imageURL string) error {
	return r.db.Model(&models.ComicStoryboard{}).Where("id = ?", id).Update("image_url", imageURL).Error
}

func (r *StoryboardRepository) UpdateDetailTTSURL(id uint, ttsURL string) error {
	return r.db.Model(&models.ComicStoryboardDetail{}).Where("id = ?", id).Update("tts_url", ttsURL).Error
}

func (r *StoryboardRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.ComicStoryboard{}).Where("id = ?", id).Update("status", status).Error
}

func (r *StoryboardRepository) GetByID(id uint) (*models.ComicStoryboard, error) {
	var storyboard models.ComicStoryboard
	err := r.db.First(&storyboard, id).Error
	return &storyboard, err
}

func (r *StoryboardRepository) CreatePage(page *models.ComicStoryboardPage) error {
	return r.db.Create(page).Error
}

func (r *StoryboardRepository) GetPageByID(id uint) (*models.ComicStoryboardPage, error) {
	var page models.ComicStoryboardPage
	err := r.db.First(&page, id).Error
	return &page, err
}

func (r *StoryboardRepository) GetPagesBySectionID(sectionID uint) ([]models.ComicStoryboardPage, error) {
	var pages []models.ComicStoryboardPage
	err := r.db.
		Where("section_id = ?", sectionID).
		Order("index ASC").
		Preload("Panels", func(db *gorm.DB) *gorm.DB {
			return db.Order("index ASC").
				Preload("Segments", func(db *gorm.DB) *gorm.DB {
					return db.Order("index ASC").Preload("Role")
				})
		}).
		Find(&pages).Error
	return pages, err
}

func (r *StoryboardRepository) UpdatePageStatus(pageID uint, status string) error {
	return r.db.Model(&models.ComicStoryboardPage{}).
		Where("id = ?", pageID).
		Update("status", status).Error
}

func (r *StoryboardRepository) CreatePanel(panel *models.ComicStoryboardPanel) error {
	return r.db.Create(panel).Error
}

func (r *StoryboardRepository) GetPanelByID(id uint) (*models.ComicStoryboardPanel, error) {
	var panel models.ComicStoryboardPanel
	err := r.db.First(&panel, id).Error
	return &panel, err
}

func (r *StoryboardRepository) UpdatePanelStatus(panelID uint, status string) error {
	return r.db.Model(&models.ComicStoryboardPanel{}).
		Where("id = ?", panelID).
		Update("status", status).Error
}

func (r *StoryboardRepository) UpdatePanelImageURL(panelID uint, imageURL string) error {
	return r.db.Model(&models.ComicStoryboardPanel{}).
		Where("id = ?", panelID).
		Update("image_url", imageURL).Error
}

func (r *StoryboardRepository) CreateSegment(segment *models.ComicStoryboardSegment) error {
	return r.db.Create(segment).Error
}

func (r *StoryboardRepository) UpdateSegmentTTSURL(segmentID uint, ttsURL string) error {
	return r.db.Model(&models.ComicStoryboardSegment{}).
		Where("id = ?", segmentID).
		Update("tts_url", ttsURL).Error
}
