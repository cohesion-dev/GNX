package services

import (
	"context"
	"fmt"

	"github.com/cohesion-dev/GNX/ai/gnxaigc"
	"github.com/cohesion-dev/GNX/backend_new/internal/repositories"
)

type TTSService struct {
	pageRepo *repositories.PageRepository
	roleRepo *repositories.RoleRepository
	aigc     *gnxaigc.GnxAIGC
}

func NewTTSService(
	pageRepo *repositories.PageRepository,
	roleRepo *repositories.RoleRepository,
	aigc *gnxaigc.GnxAIGC,
) *TTSService {
	return &TTSService{
		pageRepo: pageRepo,
		roleRepo: roleRepo,
		aigc:     aigc,
	}
}

func (s *TTSService) GetTTSAudio(ctx context.Context, detailID uint) ([]byte, error) {
	detail, err := s.pageRepo.FindDetailByID(detailID)
	if err != nil {
		return nil, fmt.Errorf("detail not found: %w", err)
	}

	fmt.Printf("Generating TTS for detail ID %d with content: %s\n", detailID, detail.Content)

	voiceType := "qiniu_zh_male_whxkxg"
	speedRatio := 1.0

	if detail.RoleID != nil {
		role, err := s.roleRepo.FindByID(*detail.RoleID)
		if err == nil {
			voiceType = role.VoiceType
		}
	}

	audioData, err := s.aigc.TextToSpeechSimple(ctx, detail.Content, voiceType, speedRatio)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TTS: %w", err)
	}

	return audioData, nil
}
