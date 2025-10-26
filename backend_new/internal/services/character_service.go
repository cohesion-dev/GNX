package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/cohesion-dev/GNX/ai/gnxaigc"
	"github.com/cohesion-dev/GNX/backend_new/internal/models"
	"github.com/cohesion-dev/GNX/backend_new/internal/repositories"
	"github.com/cohesion-dev/GNX/backend_new/pkg/logger"
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

	logger.Info("[Character Assets] Syncing assets for %d characters", len(features))
	for _, feature := range features {
		name := strings.TrimSpace(feature.Basic.Name)
		if name == "" {
			logger.Warn("[Character Assets] Skipping character with empty name")
			continue
		}

		role, exists := roleMap[name]
		if !exists {
			logger.Warn("[Character Assets] Character %s not found in roles", name)
			continue
		}

		prompt := strings.TrimSpace(feature.ConceptArtPrompt)
		if prompt == "" {
			logger.Warn("[Character Assets] Character %s missing concept_art_prompt, skipping image generation", name)
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
					logger.Warn("[Character Assets] Character %s: cannot read existing concept art (%v), using text generation", name, err)
					imageData, err = s.aigc.GenerateImageByText(ctx, fullPrompt)
					if err != nil {
						logger.Error("[Character Assets] Character %s: failed to generate concept art: %v", name, err)
						continue
					}
				} else {
					logger.Info("[Character Assets] Character %s: Refining concept art via img2img", name)
					imageData, err = s.aigc.GenerateImageByImage(ctx, baseData, fullPrompt)
					if err != nil {
						logger.Warn("[Character Assets] Character %s: img2img refinement failed (%v), falling back to text generation", name, err)
						imageData, err = s.aigc.GenerateImageByText(ctx, fullPrompt)
						if err != nil {
							logger.Error("[Character Assets] Character %s: failed to generate concept art: %v", name, err)
							continue
						}
					}
				}
			} else {
				logger.Info("[Character Assets] Character %s: Generating concept art from scratch", name)
				imageData, err = s.aigc.GenerateImageByText(ctx, fullPrompt)
				if err != nil {
					logger.Error("[Character Assets] Character %s: failed to generate concept art: %v", name, err)
					continue
				}
			}

			imageID := fmt.Sprintf("character_%d_%s", role.ID, name)
			if err := s.storage.UploadBytes(imageData, imageID); err != nil {
				logger.Error("[Character Assets] Character %s: failed to upload concept art: %v", name, err)
				continue
			}

			role.ImageID = imageID
			if err := s.roleRepo.Update(role); err != nil {
				logger.Error("[Character Assets] Character %s: failed to update role with imageID: %v", name, err)
			} else {
				logger.Info("[Character Assets] Character %s: Concept art uploaded (imageID=%s)", name, imageID)
			}
		} else {
			logger.Info("[Character Assets] Character %s: Reusing existing concept art", name)
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
			logger.Warn("[Character Assets] Failed to download image for character %s: %v", role.Name, err)
			continue
		}

		assets[role.Name] = &CharacterAsset{
			ImageData: imageData,
		}
	}

	return assets, nil
}
