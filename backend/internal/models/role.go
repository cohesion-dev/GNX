package models

import (
	"time"
)

type ComicRole struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ComicID   uint      `json:"comic_id" gorm:"not null"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	ImageURL  string    `json:"image_url" gorm:"size:500"`
	Brief     string    `json:"brief" gorm:"type:text"`
	Voice     string    `json:"voice" gorm:"size:100"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Comic Comic `json:"comic,omitempty" gorm:"foreignKey:ComicID"`
}

func (ComicRole) TableName() string {
	return "comic_role"
}
