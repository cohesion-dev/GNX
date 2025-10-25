package main

import (
	"qiniu-ai-image-generator/gnxaigc"
)

// CharacterAsset tracks the persisted concept art for a role to keep cross-chapter consistency.
type CharacterAsset struct {
	Feature   gnxaigc.CharacterFeature
	ImagePath string
	Prompt    string
	FileStem  string
}

type ChapterCharacterEntry struct {
	Name            string `json:"name"`
	ImageFile       string `json:"image_file,omitempty"`
	PromptFile      string `json:"prompt_file,omitempty"`
	GlobalImageFile string `json:"global_image_file,omitempty"`
	GlobalPrompt    string `json:"global_prompt,omitempty"`
	ConceptArtNotes string `json:"concept_art_notes,omitempty"`
}

type ChapterCharacterManifest struct {
	Characters []ChapterCharacterEntry `json:"characters"`
}
