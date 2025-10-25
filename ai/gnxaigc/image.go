package gnxaigc

import (
	"bytes"
	"cmp"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Config struct {
	APIKey        string `json:"api_key,omitempty"`
	BaseURL       string `json:"base_url,omitempty"`
	ImageModel    string `json:"image_model,omitempty"`
	LanguageModel string `json:"language_model,omitempty"`
}

func (c *Config) validate() {
	c.APIKey = cmp.Or(c.APIKey, os.Getenv("OPENAI_API_KEY"))
	c.BaseURL = cmp.Or(c.BaseURL, os.Getenv("OPENAI_BASE_URL"), "https://openai.qiniu.com/v1")
	c.ImageModel = cmp.Or(c.ImageModel, "gemini-2.5-flash-image")
	c.LanguageModel = cmp.Or(c.LanguageModel, "deepseek/deepseek-v3.1-terminus")
}

type GnxAIGC struct {
	Config
	client openai.Client
}

func NewGnxAIGC(cfg Config) *GnxAIGC {
	cfg.validate()
	return &GnxAIGC{
		Config: cfg,
		client: openai.NewClient(
			option.WithAPIKey(cfg.APIKey),
			option.WithBaseURL(cfg.BaseURL),
		),
	}
}

func (g *GnxAIGC) GenerateImageByText(ctx context.Context, prompt string) ([]byte, error) {
	resp, err := g.client.Images.Generate(context.TODO(), openai.ImageGenerateParams{
		Prompt: prompt,
		Model:  g.ImageModel,
		N:      openai.Int(1),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate image: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, errors.New("no image data received")
	}

	bs, err := base64.StdEncoding.DecodeString(resp.Data[0].B64JSON)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image data: %w", err)
	}

	return bs, nil
}

func (g *GnxAIGC) GenerateImageByImage(ctx context.Context, imageData []byte, prompt string) ([]byte, error) {
	resp, err := g.client.Images.Edit(ctx, openai.ImageEditParams{
		Image: openai.ImageEditParamsImageUnion{
			OfFile: bytes.NewReader(imageData),
		},
		Prompt: prompt,
		N:      openai.Int(1),
		Model:  g.ImageModel,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to edit image: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, errors.New("no image data received")
	}

	bs, err := base64.StdEncoding.DecodeString(resp.Data[0].B64JSON)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image data: %w", err)
	}

	return bs, nil
}
