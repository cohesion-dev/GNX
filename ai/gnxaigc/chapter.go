package gnxaigc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
	// ConceptArtPrompt 用于生成角色原画的英文提示词，保证跨页一致
	ConceptArtPrompt string `json:"concept_art_prompt"`
	// ConceptArtNotes 可选补充说明，记录与上一章的差异或微调方向
	ConceptArtNotes string `json:"concept_art_notes,omitempty"`
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
	// MaxPanelsPerPage 控制单页内的最大分格数量，默认四格，至少一格
	MaxPanelsPerPage int
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

type StoryboardPanel struct {
	// 分镜的原文文本片段，因为一段话可能有不同的语音形式故需要分割
	SourceTextSegments []SourceTextSegment `json:"source_text_segments"`
	// PanelSummary 用于概述此分格的关键情节（可选）
	PanelSummary string `json:"panel_summary,omitempty"`
	// VisualPrompt 详细描述该分格的主要视觉元素与构图
	VisualPrompt string `json:"visual_prompt"`
}

type StoryboardPage struct {
	// 单页内的多个分格，保持 1-4 个结构
	Panels []StoryboardPanel `json:"panels"`
	// LayoutHint 描述分格排列方式，如 "2x2 grid"、"vertical triptych"
	LayoutHint string `json:"layout_hint"`
	// 用于提供给文生图大模型的提示词（按页生成）
	ImagePrompt string `json:"image_prompt"`
	// PageSummary 概述整页节奏（可选）
	PageSummary string `json:"page_summary,omitempty"`
}

type SummaryChapterOutput struct {
	// 分页分镜列表
	StoryboardPages []StoryboardPage `json:"storyboard_pages"`
	// 输出的角色画像更新（含原画提示词，需提供给下游图生图流程）
	CharacterFeatures []CharacterFeature `json:"character_features"`
}

const (
	defaultMaxPanelsPerPage = 4
	minPanelsPerPage        = 1
)

func maxPanelsPerPageOrDefault(limit int) int {
	if limit < minPanelsPerPage {
		return defaultMaxPanelsPerPage
	}
	return limit
}

func buildVoiceStylesJSON(items []TTSVoiceItem) string {
	if len(items) == 0 {
		return "[]"
	}
	bs, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return "[]"
	}
	return string(bs)
}

func buildCharacterFeaturesJSON(items []CharacterFeature) string {
	if len(items) == 0 {
		return "[]"
	}
	bs, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return "[]"
	}
	return string(bs)
}

func buildStoryboardSchema(maxPanelsPerPage int) map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"storyboard_pages", "character_features"},
		"properties": map[string]any{
			"storyboard_pages": map[string]any{
				"type":     "array",
				"minItems": 1,
				"items": map[string]any{
					"type": "object",
					"required": []string{
						"panels",
						"layout_hint",
						"image_prompt",
					},
					"properties": map[string]any{
						"panels": map[string]any{
							"type":        "array",
							"description": "单页内的多个分格，保持 1-4 个结构。",
							"minItems":    minPanelsPerPage,
							"maxItems":    maxPanelsPerPage,
							"items": map[string]any{
								"type": "object",
								"required": []string{
									"source_text_segments",
									"visual_prompt",
								},
								"properties": map[string]any{
									"source_text_segments": map[string]any{
										"type":        "array",
										"description": "该分镜对应的多个语音文本片段及其配音选择。如一句话中可能包含旁白和对话，需要分成多个文本片段分别处理。",
										"minItems":    1,
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
									"panel_summary": map[string]any{
										"type":        "string",
										"description": "可选：概述该分格的关键情节",
									},
									"visual_prompt": map[string]any{
										"type":        "string",
										"description": "详细描述该分格需要呈现的视觉元素、构图与角色动作，建议使用英文描述。",
									},
								},
							},
						},
						"layout_hint": map[string]any{
							"type":        "string",
							"description": "描述分格排列方式，例如 '2x2 grid'、'3-panel vertical strip'。",
						},
						"image_prompt": map[string]any{
							"type":        "string",
							"description": "用于生成该页图像的提示词，建议使用英文描述，并确保整页风格统一。",
						},
						"page_summary": map[string]any{
							"type":        "string",
							"description": "可选：概述整页的节奏或视觉重点。",
						},
					},
				},
			},
			"character_features": map[string]any{
				"type":        "array",
				"description": "本章涉及的角色画像配置，用于指导原画与后续面板生成。",
				"minItems":    1,
				"items": map[string]any{
					"type": "object",
					"required": []string{
						"basic",
						"visual",
						"tts",
						"concept_art_prompt",
					},
					"properties": map[string]any{
						"basic": map[string]any{
							"type":     "object",
							"required": []string{"name", "gender", "age"},
							"properties": map[string]any{
								"name":   map[string]any{"type": "string"},
								"gender": map[string]any{"type": "string"},
								"age":    map[string]any{"type": "string"},
							},
						},
						"visual": map[string]any{
							"type":     "object",
							"required": []string{"hair", "habitual_expression", "skin_tone", "face_shape"},
							"properties": map[string]any{
								"hair":                map[string]any{"type": "string"},
								"habitual_expression": map[string]any{"type": "string"},
								"skin_tone":           map[string]any{"type": "string"},
								"face_shape":          map[string]any{"type": "string"},
							},
						},
						"tts": map[string]any{
							"type":     "object",
							"required": []string{"voice_name", "voice_type", "speed_ratio"},
							"properties": map[string]any{
								"voice_name":  map[string]any{"type": "string"},
								"voice_type":  map[string]any{"type": "string"},
								"speed_ratio": map[string]any{"type": "number"},
							},
						},
						"concept_art_prompt": map[string]any{
							"type":        "string",
							"description": "用于生成角色原画的英文提示词。",
						},
						"concept_art_notes": map[string]any{
							"type":        "string",
							"description": "可选补充说明，写明与上一章的差异或微调方向。",
						},
						"comment": map[string]any{
							"type": "string",
						},
					},
				},
			},
		},
	}
}

func buildSummaryChapterPrompt(input SummaryChapterInput, voiceStylesJSON, schemaJSON string, maxPanelsPerPage int) string {
	novelTitle := input.NovelTitle
	chapterTitle := input.ChapterTitle
	existingCharactersJSON := buildCharacterFeaturesJSON(input.CharacterFeatures)
	return fmt.Sprintf(`
你是一个擅长从小说生成动漫分镜和配音选择的设计师，后续用户将给你每一章的小说原文，你需要按指定的输出格式进行输出。

当前用户选择的小说标题为：《%s》，章节标题为：《%s》。如果你熟悉该小说的背景设定和角色人设，也可以结合你已有的知识进行参考。

配音选择时，你可以从以下提供的语音风格列表中选择合适的语音风格：

%s

以下为已知的角色画像配置（可能来自上一章的原画或既有设定，若为空表示为初次出场）：

%s

请根据小说内容和情感，将章节拆分成多页，每一页包含 1 至 %d 个分格（panel）。确保页面之间的剧情推进自然，必要时可以增加页数，避免把大量剧情挤在同一页。为每个分格拆分合适的语音文本片段，并为每个片段选择合适的语音风格和语速比例（1.0 为正常语速，>1.0 为加快语速，<1.0 为放慢语速）。

图像生成以“页”为单位，请：
1. 为每页提供 layout_hint，明确描述分格在页面上的排列方式（如 2x2 grid、三段纵向排版等）。
2. 为每个分格提供 visual_prompt，详细描述该分格的画面构图、角色姿态、表情、关键道具与背景信息。
3. 在 image_prompt 中，总结整页应呈现的整体风格、氛围与需要统一的视觉要素，并说明应绘制为多分格漫画页面，保持 panel 之间通过细边框分隔。
4. 所有 layout_hint、visual_prompt 与 image_prompt 必须使用英语描述，不得出现任何中文字符，也不要提示模型在图像中加入文字。

在输出的 character_features 中，请：
1. 覆盖本章出现的每位角色（含新角色与历史角色），并输出英文的 concept_art_prompt，确保可直接用于角色原画的文生图。
2. 若角色已经在已有配置中出现，请继承其既有视觉特征，必要时仅在 concept_art_notes 中注明微调要点，并保持 core design 一致。
3. 若角色为全新出场，请在 concept_art_notes 中注明 "new character"，并给出灵感来源或与剧情相关的设计理由。
4. concept_art_prompt 必须避免引导模型生成文字或中文字符，应聚焦于角色造型、服装、配色、光线、姿态等视觉细节。
5. 请确保 storyboard_pages 中对角色的描写与对应的 concept_art_prompt 一致，避免跨页设定冲突。

请严格按照以下给定的JSONSchema, 仅输出一个合法的 JSON 对象, 不要包含任何前导或后续的说明文字、代码块标记、引号等进行输出结果的编写，确保输出内容**严格符合JSONSchema的要求**且格式正确:

%s
`,
		novelTitle,
		chapterTitle,
		voiceStylesJSON,
		existingCharactersJSON,
		maxPanelsPerPage,
		schemaJSON,
	)
}

func (g *GnxAIGC) SummaryChapter(ctx context.Context, input SummaryChapterInput) (*SummaryChapterOutput, error) {
	maxPanelsPerPage := maxPanelsPerPageOrDefault(input.MaxPanelsPerPage)
	jsonSchema := buildStoryboardSchema(maxPanelsPerPage)
	jsonSchemaBytes, err := json.MarshalIndent(jsonSchema, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json schema: %w", err)
	}
	voiceStylesJSON := buildVoiceStylesJSON(input.AvailableVoiceStyles)
	prompt := buildSummaryChapterPrompt(input, voiceStylesJSON, string(jsonSchemaBytes), maxPanelsPerPage)

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

// ComposePageImagePrompt 将页面级别的图像提示词与分格视觉描述整合，强化多分格漫画的布局指令。
func ComposePageImagePrompt(stylePrefix string, page StoryboardPage) string {
	var builder strings.Builder

	appendWithSpace := func(text string) {
		if text == "" {
			return
		}
		if builder.Len() > 0 && !strings.HasSuffix(builder.String(), " ") {
			builder.WriteString(" ")
		}
		builder.WriteString(text)
	}

	appendWithSpace(strings.TrimSpace(stylePrefix))
	appendWithSpace(strings.TrimSpace(page.ImagePrompt))

	layout := strings.TrimSpace(page.LayoutHint)
	if layout == "" {
		layout = fmt.Sprintf("%d panels comic layout", len(page.Panels))
	}

	appendWithSpace(fmt.Sprintf("Comic page layout: %s with clear gutters and panel borders.", layout))

	for idx, panel := range page.Panels {
		visual := strings.TrimSpace(panel.VisualPrompt)
		if visual == "" {
			visual = strings.TrimSpace(panel.PanelSummary)
		}
		if visual == "" {
			continue
		}
		appendWithSpace(fmt.Sprintf("Panel %d: %s", idx+1, visual))
	}

	appendWithSpace("Use English-only descriptive language. No Chinese characters or typography. Avoid rendering any on-screen text.")

	return strings.TrimSpace(builder.String())
}
