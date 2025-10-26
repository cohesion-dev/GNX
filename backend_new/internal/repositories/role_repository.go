package repositories

import (
	"github.com/cohesion-dev/GNX/backend_new/internal/models"
	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) Create(role *models.ComicRole) error {
	return r.db.Create(role).Error
}

func (r *RoleRepository) FindByComicID(comicID uint) ([]models.ComicRole, error) {
	var roles []models.ComicRole
	err := r.db.Where("comic_id = ?", comicID).Find(&roles).Error
	return roles, err
}

func (r *RoleRepository) FindByID(id uint) (*models.ComicRole, error) {
	var role models.ComicRole
	err := r.db.First(&role, id).Error
	return &role, err
}
