package models

import (
	"time"
)

type Comic struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Title      string    `gorm:"not null" json:"title"`
	Brief      string    `json:"brief"`
	Icon       string    `json:"icon"`
	Bg         string    `json:"bg"`
	UserPrompt string    `json:"user_prompt"`
	Status     string    `gorm:"default:pending" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	
	Roles    []Role    `gorm:"foreignKey:ComicID" json:"roles,omitempty"`
	Sections []Section `gorm:"foreignKey:ComicID" json:"sections,omitempty"`
}
