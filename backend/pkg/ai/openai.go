package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cohesion-dev/GNX/backend/internal/models"
)

type OpenAIClient struct {
	apiKey  string
	baseURL string
}

type NovelAnalysis struct {
	Brief      string
	Roles      []RoleInfo
	IconPrompt string
	BgPrompt   string
}

type RoleInfo struct {
	Name        string
	Brief       string
	Voice       string
	ImagePrompt string
}

type StoryboardInfo struct {
	ImagePrompt string
	Details     []StoryboardDetailInfo
}

type StoryboardDetailInfo struct {
	Text     string
	RoleName string
	Voice    string
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
	}
}

func (c *OpenAIClient) AnalyzeNovel(content, userPrompt string) (*NovelAnalysis, error) {
	prompt := fmt.Sprintf(`请分析以下小说内容，并按照以下JSON格式返回分析结果：
{
  "brief": "小说简介",
  "roles": [
    {
      "name": "角色名称",
      "brief": "角色简介",
      "voice": "音色标识(如male_young, female_young等)",
      "image_prompt": "用于生成角色头像的英文提示词"
    }
  ],
  "icon_prompt": "用于生成漫画图标的英文提示词",
  "bg_prompt": "用于生成漫画背景的英文提示词"
}

用户提示词: %s

小说内容:
%s`, userPrompt, content)

	response, err := c.chatCompletion(prompt)
	if err != nil {
		return nil, err
	}

	var analysis NovelAnalysis
	if err := json.Unmarshal([]byte(response), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse analysis: %w", err)
	}

	return &analysis, nil
}

func (c *OpenAIClient) GenerateStoryboards(content string, roles []models.ComicRole, userPrompt string) ([]StoryboardInfo, error) {
	rolesJSON, _ := json.Marshal(roles)
	prompt := fmt.Sprintf(`请将以下章节内容转换为分镜脚本，按照以下JSON格式返回：
[
  {
    "image_prompt": "用于生成分镜图片的英文提示词",
    "details": [
      {
        "text": "对话或旁白文本",
        "role_name": "角色名称(如果是角色对话)",
        "voice": "音色标识"
      }
    ]
  }
]

已知角色信息: %s
用户提示词: %s

章节内容:
%s`, string(rolesJSON), userPrompt, content)

	response, err := c.chatCompletion(prompt)
	if err != nil {
		return nil, err
	}

	var storyboards []StoryboardInfo
	if err := json.Unmarshal([]byte(response), &storyboards); err != nil {
		return nil, fmt.Errorf("failed to parse storyboards: %w", err)
	}

	return storyboards, nil
}

func (c *OpenAIClient) GenerateImage(prompt, userPrompt string) ([]byte, error) {
	fullPrompt := fmt.Sprintf("%s. Style: %s", prompt, userPrompt)

	requestBody, _ := json.Marshal(map[string]interface{}{
		"model":  "dall-e-3",
		"prompt": fullPrompt,
		"n":      1,
		"size":   "1024x1024",
	})

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/images/generations", c.baseURL), bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var result struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no image generated")
	}

	imageResp, err := http.Get(result.Data[0].URL)
	if err != nil {
		return nil, err
	}
	defer imageResp.Body.Close()

	return ioutil.ReadAll(imageResp.Body)
}

func (c *OpenAIClient) chatCompletion(prompt string) (string, error) {
	requestBody, _ := json.Marshal(map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是一个专业的小说分析和漫画脚本创作助手。请严格按照要求的JSON格式返回结果。",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
	})

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", c.baseURL), bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return result.Choices[0].Message.Content, nil
}
