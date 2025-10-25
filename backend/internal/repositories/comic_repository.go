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

func (r *ComicRepository) GetListWithFilter(limit, offset int, status string) ([]models.Comic, int64, error) {
	var comics []models.Comic
	var total int64

	query := r.db.Model(&models.Comic{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&comics).Error
	return comics, total, err
}

func (r *ComicRepository) GetByID(id uint) (*models.Comic, error) {
	var comic models.Comic
	err := r.db.First(&comic, id).Error
	return &comic, err
}

func (r *ComicRepository) GetByIDWithRelations(id uint) (*models.Comic, error) {
	var comic models.Comic
	err := r.db.Preload("Roles").Preload("Sections").First(&comic, id).Error
	return &comic, err
}

func (r *ComicRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.Comic{}).Where("id = ?", id).Update("status", status).Error
}

func (r *ComicRepository) UpdateBrief(id uint, brief string) error {
	return r.db.Model(&models.Comic{}).Where("id = ?", id).Update("brief", brief).Error
}

func (r *ComicRepository) UpdateIcon(id uint, icon string) error {
	return r.db.Model(&models.Comic{}).Where("id = ?", id).Update("icon", icon).Error
}

func (r *ComicRepository) UpdateBg(id uint, bg string) error {
	return r.db.Model(&models.Comic{}).Where("id = ?", id).Update("bg", bg).Error
}

func (r *ComicRepository) GetByTitle(title string) (*models.Comic, error) {
	var comic models.Comic
	err := r.db.Where("title = ?", title).First(&comic).Error
	return &comic, err
}

func (r *ComicRepository) UpdateIconPrompt(id uint, prompt string) error {
	return r.db.Model(&models.Comic{}).Where("id = ?", id).Update("icon_prompt", prompt).Error
}

func (r *ComicRepository) UpdateBgPrompt(id uint, prompt string) error {
	return r.db.Model(&models.Comic{}).Where("id = ?", id).Update("bg_prompt", prompt).Error
}

func (r *ComicRepository) UpdateHasMoreContent(id uint, hasMore bool) error {
	return r.db.Model(&models.Comic{}).Where("id = ?", id).Update("has_more_content", hasMore).Error
}

func (r *ComicRepository) UpdateNovelFileURL(id uint, url string) error {
	return r.db.Model(&models.Comic{}).Where("id = ?", id).Update("novel_file_url", url).Error
}

func (r *ComicRepository) UpdateComicInfo(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Comic{}).Where("id = ?", id).Updates(updates).Error
}
