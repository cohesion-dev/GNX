package gnxaigc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTTSSpeech(t *testing.T) {
	g := NewGnxAIGC(Config{})
	items, err := g.GetVoiceList(context.TODO())
	require.NoError(t, err)
	require.Greater(t, len(items), 0)

	audioData, err := g.TextToSpeechSimple(context.TODO(), "你好，欢迎使用GNX智能语音合成服务。", items[0].VoiceType, 1.0)
	require.NoError(t, err)
	require.Greater(t, len(audioData), 0)
}
