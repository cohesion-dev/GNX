package aigc

import (
	"github.com/cohesion-dev/GNX/ai/gnxaigc"
	"github.com/cohesion-dev/GNX/backend_new/config"
)

func NewAIGC(cfg *config.AIConfig) *gnxaigc.GnxAIGC {
	return gnxaigc.NewGnxAIGC(gnxaigc.Config{
		APIKey:        cfg.APIKey,
		BaseURL:       cfg.BaseURL,
		ImageModel:    cfg.ImageModel,
		LanguageModel: cfg.LanguageModel,
	})
}
