package models

import (
	"time"
)

type ComicSection struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ComicID   uint      `json:"comic_id" gorm:"not null"`
	Index     int       `json:"index" gorm:"not null"`
	Detail    string    `json:"detail" gorm:"type:text;not null"`
	Status    string    `json:"status" gorm:"size:20;default:'pending'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Comic       Comic             `json:"comic,omitempty" gorm:"foreignKey:ComicID"`
	Storyboards []ComicStoryboard `json:"storyboards,omitempty" gorm:"foreignKey:SectionID"`
}

func (ComicSection) TableName() string {
	return "comic_section"
}
