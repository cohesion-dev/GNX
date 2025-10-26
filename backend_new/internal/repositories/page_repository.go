package repositories

import (
	"github.com/cohesion-dev/GNX/backend_new/internal/models"
	"gorm.io/gorm"
)

type PageRepository struct {
	db *gorm.DB
}

func NewPageRepository(db *gorm.DB) *PageRepository {
	return &PageRepository{db: db}
}

func (r *PageRepository) Create(page *models.ComicPage) error {
	return r.db.Create(page).Error
}

func (r *PageRepository) FindByID(id uint) (*models.ComicPage, error) {
	var page models.ComicPage
	err := r.db.Preload("Details").First(&page, id).Error
	return &page, err
}

func (r *PageRepository) FindBySectionID(sectionID uint) ([]models.ComicPage, error) {
	var pages []models.ComicPage
	err := r.db.Where("section_id = ?", sectionID).Order("index ASC").Preload("Details").Find(&pages).Error
	return pages, err
}

func (r *PageRepository) CreateDetail(detail *models.ComicPageDetail) error {
	return r.db.Create(detail).Error
}

func (r *PageRepository) FindDetailByID(id uint) (*models.ComicPageDetail, error) {
	var detail models.ComicPageDetail
	err := r.db.First(&detail, id).Error
	return &detail, err
}

func (r *PageRepository) Update(page *models.ComicPage) error {
	return r.db.Save(page).Error
}

func (r *PageRepository) UpdateDetail(detail *models.ComicPageDetail) error {
	return r.db.Save(detail).Error
}
