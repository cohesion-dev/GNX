package services

import (
	"gorm.io/gorm"
)

type ComicService struct {
	db *gorm.DB
}

func NewComicService(db *gorm.DB) *ComicService {
	return &ComicService{db: db}
}
