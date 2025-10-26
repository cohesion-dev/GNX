package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/cohesion-dev/GNX/ai/gnxaigc"
	"github.com/cohesion-dev/GNX/backend_new/internal/models"
	"github.com/cohesion-dev/GNX/backend_new/internal/repositories"
	"github.com/cohesion-dev/GNX/backend_new/pkg/storage"
)

type CharacterAsset struct {
	Feature   gnxaigc.CharacterFeature
	ImageData []byte
	Prompt    string
}

type CharacterService struct {
	roleRepo *repositories.RoleRepository
	storage  *storage.Storage
	aigc     *gnxaigc.GnxAIGC
}

func NewCharacterService(
	roleRepo *repositories.RoleRepository,
	storage *storage.Storage,
	aigc *gnxaigc.GnxAIGC,
) *CharacterService {
	return &CharacterService{
		roleRepo: roleRepo,
		storage:  storage,
		aigc:     aigc,
	}
}

// SyncCharacterAssets generates or updates concept art for characters
func (s *CharacterService) SyncCharacterAssets(
	ctx context.Context,
	comicID uint,
	imageStyle string,
	features []gnxaigc.CharacterFeature,
) (map[string]*CharacterAsset, error) {
	roles, err := s.roleRepo.FindByComicID(comicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	roleMap := make(map[string]*models.ComicRole)
	for i := range roles {
		roleMap[roles[i].Name] = &roles[i]
	}

	assets := make(map[string]*CharacterAsset)
	trimmedStyle := strings.TrimSpace(imageStyle)

	for _, feature := range features {
		name := strings.TrimSpace(feature.Basic.Name)
		if name == "" {
			fmt.Println("  Skipping character with empty name")
			continue
		}

		role, exists := roleMap[name]
		if !exists {
			fmt.Printf("  Character %s not found in roles\n", name)
			continue
		}

		prompt := strings.TrimSpace(feature.ConceptArtPrompt)
		if prompt == "" {
			fmt.Printf("  [Character %s] Missing concept_art_prompt, skip image generation\n", name)
			assets[name] = &CharacterAsset{
				Feature: feature,
			}
			continue
		}

		fullPrompt := prompt
		if trimmedStyle != "" {
			fullPrompt = fmt.Sprintf("%s %s", trimmedStyle, prompt)
		}
		fullPrompt = strings.TrimSpace(fullPrompt)

		var imageData []byte
		shouldGenerate := role.ImageID == ""

		if !shouldGenerate {
			existingImageData, err := s.storage.DownloadBytes(role.ImageID)
			if err != nil || len(existingImageData) == 0 {
				shouldGenerate = true
			} else {
				imageData = existingImageData
			}
		}

		if shouldGenerate {
			if role.ImageID != "" {
				baseData, err := s.storage.DownloadBytes(role.ImageID)
				if err != nil {
					fmt.Printf("  [Character %s] Warning: cannot read existing concept art (%v), fallback to text generation\n", name, err)
					imageData, err = s.aigc.GenerateImageByText(ctx, fullPrompt)
					if err != nil {
						fmt.Printf("    Error generating concept art: %v\n", err)
						continue
					}
				} else {
					fmt.Printf("  [Character %s] Refining concept art via img2img\n", name)
					imageData, err = s.aigc.GenerateImageByImage(ctx, baseData, fullPrompt)
					if err != nil {
						fmt.Printf("    Error refining concept art: %v\n", err)
						fmt.Printf("    Falling back to text-to-image generation\n")
						imageData, err = s.aigc.GenerateImageByText(ctx, fullPrompt)
						if err != nil {
							fmt.Printf("    Error generating concept art: %v\n", err)
							continue
						}
					}
				}
			} else {
				fmt.Printf("  [Character %s] Generating concept art from scratch\n", name)
				imageData, err = s.aigc.GenerateImageByText(ctx, fullPrompt)
				if err != nil {
					fmt.Printf("    Error generating concept art: %v\n", err)
					continue
				}
			}

			imageID := fmt.Sprintf("character_%d_%s", role.ID, name)
			if err := s.storage.UploadBytes(imageData, imageID); err != nil {
				fmt.Printf("    Error uploading concept art: %v\n", err)
				continue
			}

			role.ImageID = imageID
			if err := s.roleRepo.Update(role); err != nil {
				fmt.Printf("    Error updating role with image ID: %v\n", err)
			}
		} else {
			fmt.Printf("  [Character %s] Reusing existing concept art\n", name)
		}

		assets[name] = &CharacterAsset{
			Feature:   feature,
			ImageData: imageData,
			Prompt:    fullPrompt,
		}
	}

	return assets, nil
}

// LoadCharacterAssets loads existing character concept art from storage
func (s *CharacterService) LoadCharacterAssets(
	ctx context.Context,
	comicID uint,
) (map[string]*CharacterAsset, error) {
	roles, err := s.roleRepo.FindByComicID(comicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	assets := make(map[string]*CharacterAsset)
	for i := range roles {
		role := &roles[i]
		if role.ImageID == "" {
			continue
		}

		imageData, err := s.storage.DownloadBytes(role.ImageID)
		if err != nil {
			fmt.Printf("  Warning: failed to download image for character %s: %v\n", role.Name, err)
			continue
		}

		assets[role.Name] = &CharacterAsset{
			ImageData: imageData,
		}
	}

	return assets, nil
}
