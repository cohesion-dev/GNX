package models

import (
	"time"
)

type ComicStoryboardPanel struct {
	ID        uint `json:"id" gorm:"primaryKey"`
	SectionID uint `json:"section_id" gorm:"not null;index:idx_section_panel"`
	PageID    uint `json:"page_id" gorm:"not null;index:idx_page_panel"`
	Index     int  `json:"index" gorm:"not null"`
	
	VisualPrompt string `json:"visual_prompt" gorm:"type:text;not null"`
	PanelSummary string `json:"panel_summary,omitempty" gorm:"type:text"`
	
	ImageURL  string    `json:"image_url" gorm:"size:500"`
	Status    string    `json:"status" gorm:"size:20;default:'pending'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Section              ComicSection        `json:"section,omitempty" gorm:"foreignKey:SectionID"`
	Page                 ComicStoryboardPage `json:"page,omitempty" gorm:"foreignKey:PageID"`
	SourceTextSegments   []ComicStoryboardSegment `json:"source_text_segments,omitempty" gorm:"foreignKey:PanelID"`
}

func (ComicStoryboardPanel) TableName() string {
	return "comic_storyboard_panel"
}

type ComicStoryboardSegment struct {
	ID      uint `json:"id" gorm:"primaryKey"`
	PanelID uint `json:"panel_id" gorm:"not null;index:idx_panel_segment"`
	Index   int  `json:"index" gorm:"not null"`
	
	Text          string   `json:"text" gorm:"type:text;not null"`
	CharacterRefs []string `json:"character_refs,omitempty" gorm:"type:text[]"`
	
	TTSUrl    string    `json:"tts_url,omitempty" gorm:"size:500"`
	RoleID    *uint     `json:"role_id,omitempty" gorm:"index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Panel ComicStoryboardPanel `json:"panel,omitempty" gorm:"foreignKey:PanelID"`
	Role  *ComicRole           `json:"role,omitempty" gorm:"foreignKey:RoleID"`
}

func (ComicStoryboardSegment) TableName() string {
	return "comic_storyboard_segment"
}
