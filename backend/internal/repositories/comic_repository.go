package repositories

import (
	"gorm.io/gorm"
)

type ComicRepository struct {
	db *gorm.DB
}

func NewComicRepository(db *gorm.DB) *ComicRepository {
	return &ComicRepository{db: db}
}
