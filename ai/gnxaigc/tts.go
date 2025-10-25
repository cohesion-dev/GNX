package gnxaigc

import (
	"context"
	"encoding/base64"
	"fmt"
)

// VoiceItem 音色信息结构体
type VoiceItem struct {
	VoiceName  string `json:"voice_name"`
	VoiceType  string `json:"voice_type"`
	URL        string `json:"url"`
	Category   string `json:"category"`
	UpdateTime int64  `json:"updatetime"`
}

func (g *GnxAIGC) GetVoiceList(ctx context.Context) ([]VoiceItem, error) {
	var voiceList []VoiceItem

	err := g.client.Get(ctx, "/voice/list", nil, &voiceList)
	if err != nil {
		return nil, err
	}
	return voiceList, nil
}

// TTSAudioParams TTS音频参数
type TTSAudioParams struct {
	VoiceType  string  `json:"voice_type"`
	Encoding   string  `json:"encoding"`
	SpeedRatio float64 `json:"speed_ratio,omitempty"` // 默认1.0
}

// TTSRequestParams TTS请求参数
type TTSRequestParams struct {
	Text string `json:"text"`
}

// TTSRequest TTS请求结构体
type TTSRequest struct {
	Audio   TTSAudioParams   `json:"audio"`
	Request TTSRequestParams `json:"request"`
}

// TTSAddition TTS响应附加信息
type TTSAddition struct {
	Duration string `json:"duration"`
}

// TTSResponse TTS响应结构体
type TTSResponse struct {
	ReqID     string      `json:"reqid"`
	Operation string      `json:"operation"`
	Sequence  int         `json:"sequence"`
	Data      string      `json:"data"`
	Addition  TTSAddition `json:"addition"`
}

func (g *GnxAIGC) TextToSpeech(ctx context.Context, req TTSRequest) (*TTSResponse, error) {
	reqBody := req
	var ttsResp TTSResponse
	err := g.client.Post(ctx, "/voice/tts", reqBody, &ttsResp)
	if err != nil {
		return nil, err
	}
	return &ttsResp, nil
}

func (g *GnxAIGC) TextToSpeechSimple(ctx context.Context, text, voiceType string, ratio float64) ([]byte, error) {
	req := TTSRequest{
		Audio: TTSAudioParams{
			VoiceType:  voiceType,
			Encoding:   "mp3",
			SpeedRatio: ratio,
		},
		Request: TTSRequestParams{
			Text: text,
		},
	}
	ttsResp, err := g.TextToSpeech(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("TextToSpeech failed: %w", err)
	}
	bs, err := base64.StdEncoding.DecodeString(ttsResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode TTS data: %w", err)
	}
	return bs, nil
}
