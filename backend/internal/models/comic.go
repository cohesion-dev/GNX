package models

import "time"

type Comic struct {
	ID                uint      `gorm:"primarykey" json:"id"`
	Title             string    `gorm:"not null" json:"title"`
	UserPrompt        string    `gorm:"type:text" json:"user_prompt"`
	IconImageID       string    `gorm:"" json:"icon_image_id"`
	BackgroundImageID string    `gorm:"" json:"background_image_id"`
	Status            string    `gorm:"default:'pending'" json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	Roles    []ComicRole    `gorm:"foreignKey:ComicID" json:"roles,omitempty"`
	Sections []ComicSection `gorm:"foreignKey:ComicID;orderBy:index" json:"sections,omitempty"`
}

func (Comic) TableName() string {
	return "comics"
}

type ComicDetailResponse struct {
	ID                string         `json:"id"`
	Title             string         `json:"title"`
	IconImageID       string         `json:"icon_image_id"`
	BackgroundImageID string         `json:"background_image_id"`
	Status            string         `json:"status"`
	Roles             []ComicRole    `json:"roles,omitempty"`
	Sections          []ComicSection `json:"sections,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}
