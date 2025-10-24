package models

import (
	"time"
)

type Section struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ComicID     uint      `gorm:"not null;index" json:"comic_id"`
	Index       int       `gorm:"not null" json:"index"`
	Detail      string    `gorm:"type:text" json:"detail"`
	Status      string    `gorm:"default:pending" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	Storyboards []Storyboard `gorm:"foreignKey:SectionID" json:"storyboards,omitempty"`
}
