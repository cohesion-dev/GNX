package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cohesion-dev/GNX/ai/gnxaigc"
	"github.com/cohesion-dev/GNX/backend/internal/repositories"
	"github.com/cohesion-dev/GNX/backend/pkg/storage"
)

type TTSService struct {
	storyboardRepo *repositories.StoryboardRepository
	aigcService    *gnxaigc.GnxAIGC
	storageService *storage.QiniuClient
}

func NewTTSService(
	storyboardRepo *repositories.StoryboardRepository,
	aigcService *gnxaigc.GnxAIGC,
	storageService *storage.QiniuClient,
) *TTSService {
	return &TTSService{
		storyboardRepo: storyboardRepo,
		aigcService:    aigcService,
		storageService: storageService,
	}
}

func (s *TTSService) GetTTSAudio(segmentID uint) ([]byte, error) {
	segment, err := s.storyboardRepo.GetSegmentByID(segmentID)
	if err != nil {
		return nil, fmt.Errorf("TTS segment not found: %w", err)
	}

	if segment.TTSUrl != "" {
		resp, err := http.Get(detail.TTSUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch cached audio: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch cached audio: status %d", resp.StatusCode)
		}

		audioData, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read cached audio: %w", err)
		}

		return audioData, nil
	}

	if segment.Role == nil {
		return nil, fmt.Errorf("segment has no associated role for TTS")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	audioData, err := s.aigcService.TextToSpeechSimple(ctx, segment.Text, segment.Role.VoiceType, segment.Role.SpeedRatio)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TTS audio: %w", err)
	}

	go s.uploadAndSaveURL(segmentID, audioData)

	return audioData, nil
}

func (s *TTSService) uploadAndSaveURL(segmentID uint, audioData []byte) {
	key := fmt.Sprintf("tts/%d_%d.mp3", segmentID, time.Now().Unix())

	url, err := s.storageService.UploadAudio(key, audioData)
	if err != nil {
		return
	}

	_ = s.storyboardRepo.UpdateSegmentTTSURL(segmentID, url)
}
