package models

import (
	"time"
)

type ComicStoryboardPage struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	SectionID uint   `json:"section_id" gorm:"not null;index:idx_section_page"`
	Index     int    `json:"index" gorm:"not null"`
	
	ImagePrompt string `json:"image_prompt" gorm:"type:text;not null"`
	LayoutHint  string `json:"layout_hint" gorm:"type:text;not null"`
	PageSummary string `json:"page_summary,omitempty" gorm:"type:text"`
	
	ImageURL  string    `json:"image_url" gorm:"size:500"`
	Status    string    `json:"status" gorm:"size:20;default:'pending'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Section ComicSection           `json:"section,omitempty" gorm:"foreignKey:SectionID"`
	Panels  []ComicStoryboardPanel `json:"panels,omitempty" gorm:"foreignKey:PageID"`
}

func (ComicStoryboardPage) TableName() string {
	return "comic_storyboard_page"
}
