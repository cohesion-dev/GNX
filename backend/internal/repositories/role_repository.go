package repositories

import (
	"gorm.io/gorm"

	"github.com/cohesion-dev/GNX/backend/internal/models"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{
		db: db,
	}
}

func (r *RoleRepository) Create(role *models.ComicRole) error {
	return r.db.Create(role).Error
}

func (r *RoleRepository) GetByID(id uint) (*models.ComicRole, error) {
	var role models.ComicRole
	err := r.db.First(&role, id).Error
	return &role, err
}

func (r *RoleRepository) GetByComicID(comicID uint) ([]models.ComicRole, error) {
	var roles []models.ComicRole
	err := r.db.Where("comic_id = ?", comicID).Find(&roles).Error
	return roles, err
}

func (r *RoleRepository) UpdateImageURL(id uint, imageURL string) error {
	return r.db.Model(&models.ComicRole{}).Where("id = ?", id).Update("image_url", imageURL).Error
}

func (r *RoleRepository) GetByNameAndComicID(name string, comicID uint) (*models.ComicRole, error) {
	var role models.ComicRole
	err := r.db.Where("name = ? AND comic_id = ?", name, comicID).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) UpdateByID(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.ComicRole{}).Where("id = ?", id).Updates(updates).Error
}
