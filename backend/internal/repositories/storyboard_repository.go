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

func (r *StoryboardRepository) CreatePage(page *models.ComicStoryboardPage) error {
	return r.db.Create(page).Error
}

func (r *StoryboardRepository) CreatePanel(panel *models.ComicStoryboardPanel) error {
	return r.db.Create(panel).Error
}

func (r *StoryboardRepository) CreateSegment(segment *models.SourceTextSegment) error {
	return r.db.Create(segment).Error
}

func (r *StoryboardRepository) GetPagesBySectionID(sectionID uint) ([]models.ComicStoryboardPage, error) {
	var pages []models.ComicStoryboardPage
	err := r.db.
		Where("section_id = ?", sectionID).
		Order("index ASC").
		Preload("Panels", func(db *gorm.DB) *gorm.DB {
			return db.Order("index ASC").
				Preload("SourceTextSegments", func(db *gorm.DB) *gorm.DB {
					return db.Order("index ASC").Preload("Role")
				})
		}).
		Find(&pages).Error
	return pages, err
}

func (r *StoryboardRepository) GetPageByID(pageID uint) (*models.ComicStoryboardPage, error) {
	var page models.ComicStoryboardPage
	err := r.db.First(&page, pageID).Error
	return &page, err
}

func (r *StoryboardRepository) GetPanelByID(panelID uint) (*models.ComicStoryboardPanel, error) {
	var panel models.ComicStoryboardPanel
	err := r.db.First(&panel, panelID).Error
	return &panel, err
}

func (r *StoryboardRepository) GetSegmentByID(segmentID uint) (*models.SourceTextSegment, error) {
	var segment models.SourceTextSegment
	err := r.db.First(&segment, segmentID).Error
	return &segment, err
}

func (r *StoryboardRepository) UpdatePageImageURL(pageID uint, imageURL string) error {
	return r.db.Model(&models.ComicStoryboardPage{}).Where("id = ?", pageID).Update("image_url", imageURL).Error
}

func (r *StoryboardRepository) UpdatePageStatus(pageID uint, status string) error {
	return r.db.Model(&models.ComicStoryboardPage{}).Where("id = ?", pageID).Update("status", status).Error
}

func (r *StoryboardRepository) UpdatePanelImageURL(panelID uint, imageURL string) error {
	return r.db.Model(&models.ComicStoryboardPanel{}).Where("id = ?", panelID).Update("image_url", imageURL).Error
}

func (r *StoryboardRepository) UpdatePanelStatus(panelID uint, status string) error {
	return r.db.Model(&models.ComicStoryboardPanel{}).Where("id = ?", panelID).Update("status", status).Error
}

func (r *StoryboardRepository) UpdateSegmentTTSURL(segmentID uint, ttsURL string) error {
	return r.db.Model(&models.SourceTextSegment{}).Where("id = ?", segmentID).Update("tts_url", ttsURL).Error
}
