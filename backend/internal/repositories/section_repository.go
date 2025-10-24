package repositories

import (
	"gorm.io/gorm"

	"github.com/cohesion-dev/GNX/backend/internal/models"
)

type SectionRepository struct {
	db *gorm.DB
}

func NewSectionRepository(db *gorm.DB) *SectionRepository {
	return &SectionRepository{
		db: db,
	}
}

func (r *SectionRepository) Create(section *models.ComicSection) error {
	return r.db.Create(section).Error
}

func (r *SectionRepository) GetByID(id uint) (*models.ComicSection, error) {
	var section models.ComicSection
	err := r.db.First(&section, id).Error
	return &section, err
}

func (r *SectionRepository) GetByComicID(comicID uint) ([]models.ComicSection, error) {
	var sections []models.ComicSection
	err := r.db.Where("comic_id = ?", comicID).Order("index ASC").Find(&sections).Error
	return sections, err
}

func (r *SectionRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.ComicSection{}).Where("id = ?", id).Update("status", status).Error
}
