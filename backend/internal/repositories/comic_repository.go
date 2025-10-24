package repositories

import (
	"gorm.io/gorm"

	"github.com/cohesion-dev/GNX/backend/internal/models"
)

type ComicRepository struct {
	db *gorm.DB
}

func NewComicRepository(db *gorm.DB) *ComicRepository {
	return &ComicRepository{
		db: db,
	}
}

func (r *ComicRepository) Create(comic *models.Comic) error {
	return r.db.Create(comic).Error
}

func (r *ComicRepository) GetList(limit, offset int) ([]models.Comic, error) {
	var comics []models.Comic
	err := r.db.Limit(limit).Offset(offset).Find(&comics).Error
	return comics, err
}

func (r *ComicRepository) GetByID(id uint) (*models.Comic, error) {
	var comic models.Comic
	err := r.db.Preload("Roles").Preload("Sections").First(&comic, id).Error
	return &comic, err
}

func (r *ComicRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.Comic{}).Where("id = ?", id).Update("status", status).Error
}
