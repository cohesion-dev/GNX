package repositories

import (
	"github.com/cohesion-dev/GNX/backend_new/internal/models"
	"gorm.io/gorm"
)

type SectionRepository struct {
	db *gorm.DB
}

func NewSectionRepository(db *gorm.DB) *SectionRepository {
	return &SectionRepository{db: db}
}

func (r *SectionRepository) Create(section *models.ComicSection) error {
	return r.db.Create(section).Error
}

func (r *SectionRepository) FindByID(id uint) (*models.ComicSection, error) {
	var section models.ComicSection
	err := r.db.Preload("Pages.Details").First(&section, id).Error
	return &section, err
}

func (r *SectionRepository) FindByComicID(comicID uint) ([]models.ComicSection, error) {
	var sections []models.ComicSection
	err := r.db.Where("comic_id = ?", comicID).Order("index ASC").Find(&sections).Error
	return sections, err
}

func (r *SectionRepository) Update(section *models.ComicSection) error {
	return r.db.Save(section).Error
}

func (r *SectionRepository) CountByComicID(comicID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.ComicSection{}).Where("comic_id = ?", comicID).Count(&count).Error
	return count, err
}
