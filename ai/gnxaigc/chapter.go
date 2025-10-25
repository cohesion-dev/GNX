package gnxaigc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	jsonrepair "github.com/RealAlexandreAI/json-repair"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

type TTSVoiceItem struct {
	VoiceName string `json:"voice_name"`
	VoiceType string `json:"voice_type"`
}

// CharacterBasicProfile captures the minimal identifying attributes for a role.
type CharacterBasicProfile struct {
	Name   string `json:"name"`
	Gender string `json:"gender"`
	Age    string `json:"age"`
}

// CharacterVisualProfile stores the small visual anchor set we currently enforce.
type CharacterVisualProfile struct {
	Hair               string `json:"hair"`
	HabitualExpression string `json:"habitual_expression"`
	SkinTone           string `json:"skin_tone"`
	FaceShape          string `json:"face_shape"`
}

// CharacterTTSProfile contains the limited TTS knobs available in the stack.
type CharacterTTSProfile struct {
	VoiceName  string  `json:"voice_name"`
	VoiceType  string  `json:"voice_type"`
	SpeedRatio float64 `json:"speed_ratio"`
}

// CharacterFeature aggregates the structured persona data plus free-form remarks.
type CharacterFeature struct {
	Basic   CharacterBasicProfile  `json:"basic"`
	Visual  CharacterVisualProfile `json:"visual"`
	TTS     CharacterTTSProfile    `json:"tts"`
	Comment string                 `json:"comment,omitempty"`
}

type SummaryChapterInput struct {
	// Novel Title 小说标题
	NovelTitle string
	// Chapter Title 小说章节标题
	ChapterTitle string
	// 小说某章节的原文内容
	Content string
	// 待选的语音风格列表
	AvailableVoiceStyles []TTSVoiceItem
	// 已有角色人设与多模态锚点信息
	CharacterFeatures []CharacterFeature
}

type SourceTextSegment struct {
	// 分镜对应的语音文本片段
	Text string `json:"text"`
	// 该文本片段的语音风格描述，用于指导TTS合成
	VoiceName string `json:"voice_name"`
	// 该文本片段的音色类型，用于指导TTS合成
	VoiceType string `json:"voice_type"`
	// 该文本片段的语速比例，用于指导TTS合成
	SpeedRatio float64 `json:"speed_ratio"`
	// 是否为旁白文本片段
	IsNarration bool `json:"is_narration,omitempty"`
}

type StoryboardItem struct {
	// 分镜的原文文本片段，因为一段话可能有不同的语音形式故需要分割
	SourceTextSegments []SourceTextSegment `json:"source_text_segments"`
	// 用于提供给文生图大模型的提示词
	ImagePrompt string `json:"image_prompt"`
}

type SummaryChapterOutput struct {
	// 分镜列表
	StoryboardItems []StoryboardItem `json:"storyboard_items"`
	// 输出的角色画像更新
	CharacterFeatures []CharacterFeature `json:"character_features"`
}

func (g *GnxAIGC) SummaryChapter(ctx context.Context, input SummaryChapterInput) (*SummaryChapterOutput, error) {
	jsonSchema := map[string]any{
		"type":     "object",
		"required": []string{"storyboard_items"},
		"properties": map[string]any{
			"storyboard_items": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"required": []string{
						"source_text_segments",
						"image_prompt",
					},
					"properties": map[string]any{
						"source_text_segments": map[string]any{
							"type":        "array",
							"description": "该分镜对应的多个语音文本片段及其配音选择。如一句话中可能包含旁白和对话，需要分成多个文本片段分别处理。",
							"items": map[string]any{
								"type": "object",
								"required": []string{
									"text",
									"voice_name",
									"voice_type",
									"speed_ratio",
									"is_narration",
								},
								"properties": map[string]any{
									"text": map[string]any{
										"type":        "string",
										"description": "分镜对应的语音文本片段",
									},
									"voice_name": map[string]any{
										"type":        "string",
										"description": "该文本片段的语音风格描述",
									},
									"voice_type": map[string]any{
										"type":        "string",
										"description": "该文本片段的音色类型",
									},
									"speed_ratio": map[string]any{
										"type":        "number",
										"description": "语速比例",
									},
									"is_narration": map[string]any{
										"type":        "boolean",
										"description": "是否为旁白文本片段",
									},
								},
							},
						},
						"image_prompt": map[string]any{
							"type":        "string",
							"description": "用于生成该分镜图像的提示词，建议使用英文描述，以便更好地兼容主流文生图大模型。",
						},
					},
				},
			},
		},
	}

	prompt := fmt.Sprintf(`
你是一个擅长从小说生成动漫分镜和配音选择的设计师，后续用户将给你每一章的小说原文，你需要按指定的输出格式进行输出。

当前用户选择的小说标题为：《%s》，章节标题为：《%s》。如果你熟悉该小说的背景设定和角色人设，也可以结合你已有的知识进行参考。

配音选择时，你可以从以下提供的语音风格列表中选择合适的语音风格：

%s

请根据小说内容和情感，合理分割文本片段，并为每个片段选择合适的语音风格和语速比例（1.0为正常语速，>1.0为加快语速，<1.0为放慢语速）。

分镜设计时，请根据小说内容生成每个分镜的图像提示词，确保提示词能够准确描述该分镜的场景和氛围，并且相邻的分镜的场景一般不会突变，请尽可能输出详细的提示词，以保持一致性，这将用于后续提供给文生图大模型生成图像。

请严格按照以下给定的JSONSchema, 仅输出一个合法的 JSON 对象, 不要包含任何前导或后续的说明文字、代码块标记、引号等进行输出结果的编写，确保输出内容**严格符合JSONSchema的要求**且格式正确:

%s
`,
		input.NovelTitle,
		input.ChapterTitle,
		func() string {
			voiceStylesJSON, _ := json.MarshalIndent(input.AvailableVoiceStyles, "", "  ")
			return string(voiceStylesJSON)
		}(),
		func() string {
			jsonSchemaBytes, _ := json.MarshalIndent(jsonSchema, "", "  ")
			return string(jsonSchemaBytes)
		}(),
	)

	resp, err := g.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: g.LanguageModel,
		N:     openai.Int(1),
		Messages: []openai.ChatCompletionMessageParamUnion{
			{
				OfSystem: &openai.ChatCompletionSystemMessageParam{
					Content: openai.ChatCompletionSystemMessageParamContentUnion{
						OfString: openai.String(prompt),
					},
				},
			},
			{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfString: openai.String(input.Content),
					},
				},
			},
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &shared.ResponseFormatJSONObjectParam{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no chat completion choices received")
	}

	content := resp.Choices[0].Message.Content

	fmt.Printf("SummaryChapter chat completion content: %s\n", content)

	var output SummaryChapterOutput
	if err := json.Unmarshal([]byte(content), &output); err == nil {
		return &output, nil
	}

	// 如果解析失败，则尝试下修复
	contentFixed, err := jsonrepair.RepairJSON(content)
	if err != nil {
		return nil, fmt.Errorf("failed to repair JSON content: %w", err)
	}

	if err := json.Unmarshal([]byte(contentFixed), &output); err != nil {
		return nil, fmt.Errorf("failed to unmarshal repaired JSON content: %w", err)
	}

	return &output, nil
}
