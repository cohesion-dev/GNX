package repositories

import (
	"github.com/cohesion-dev/GNX/backend_new/internal/models"
	"gorm.io/gorm"
)

type ComicRepository struct {
	db *gorm.DB
}

func NewComicRepository(db *gorm.DB) *ComicRepository {
	return &ComicRepository{db: db}
}

func (r *ComicRepository) Create(comic *models.Comic) error {
	return r.db.Create(comic).Error
}

func (r *ComicRepository) FindByID(id uint) (*models.Comic, error) {
	var comic models.Comic
	err := r.db.Preload("Roles").Preload("Sections").First(&comic, id).Error
	return &comic, err
}

func (r *ComicRepository) List(page, limit int, status string) ([]models.Comic, int64, error) {
	var comics []models.Comic
	var total int64

	query := r.db.Model(&models.Comic{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&comics).Error
	return comics, total, err
}

func (r *ComicRepository) Update(comic *models.Comic) error {
	return r.db.Save(comic).Error
}
