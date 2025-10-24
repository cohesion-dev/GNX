package models

import (
	"time"
)

type Role struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ComicID   uint      `gorm:"not null;index" json:"comic_id"`
	Name      string    `gorm:"not null" json:"name"`
	ImageURL  string    `json:"image_url"`
	Brief     string    `json:"brief"`
	Voice     string    `json:"voice"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
