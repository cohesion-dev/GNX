package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"qiniu-ai-image-generator/gnxaigc"
)

type ComicGeneratorConfig struct {
	NovelTitle string
	OutputDir  string
	ImageStyle string
}

type ComicGenerator struct {
	ctx                context.Context
	aigc               *gnxaigc.GnxAIGC
	config             ComicGeneratorConfig
	availableVoices    []gnxaigc.TTSVoiceItem
	characterRegistry  map[string]gnxaigc.CharacterFeature
	characterOrder     []string
	characterAssets    map[string]*CharacterAsset
	globalCharacterDir string
}

func NewComicGenerator(ctx context.Context, cfg ComicGeneratorConfig, aigc *gnxaigc.GnxAIGC) *ComicGenerator {
	globalDir := filepath.Join(cfg.OutputDir, "characters")
	return &ComicGenerator{
		ctx:                ctx,
		aigc:               aigc,
		config:             cfg,
		characterRegistry:  make(map[string]gnxaigc.CharacterFeature),
		characterOrder:     make([]string, 0),
		characterAssets:    make(map[string]*CharacterAsset),
		globalCharacterDir: globalDir,
	}
}

func (g *ComicGenerator) Run(inputPath string, maxChapters int) error {
	if err := os.MkdirAll(g.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}
	if err := os.MkdirAll(g.globalCharacterDir, 0755); err != nil {
		return fmt.Errorf("creating shared character directory: %w", err)
	}

	chapters, err := SplitChaptersFromFile(inputPath)
	if err != nil {
		return fmt.Errorf("reading novel file: %w", err)
	}

	fmt.Printf("Successfully split novel into %d chapters\n", len(chapters))

	if err := g.loadVoices(); err != nil {
		return err
	}

	processCount := len(chapters)
	if maxChapters > 0 && maxChapters < processCount {
		processCount = maxChapters
	}

	for idx := 0; idx < processCount; idx++ {
		if err := g.processChapter(idx, chapters[idx]); err != nil {
			fmt.Printf("Error processing chapter %d: %v\n", idx+1, err)
		}
	}

	if err := g.writeGlobalManifest(); err != nil {
		fmt.Printf("%v\n", err)
	}

	fmt.Printf("\nTracked %d unique characters across chapters\n", len(g.characterOrder))
	fmt.Printf("\n=== Comic Generation Complete ===\n")
	fmt.Printf("Processed %d chapters\n", processCount)
	fmt.Printf("Output saved to: %s\n", g.config.OutputDir)

	return nil
}

func (g *ComicGenerator) loadVoices() error {
	fmt.Println("Fetching available TTS voices...")

	voiceList, err := g.aigc.GetVoiceList(g.ctx)
	if err != nil {
		return fmt.Errorf("fetching voice list: %w", err)
	}

	voices := make([]gnxaigc.TTSVoiceItem, 0, len(voiceList))
	for _, voice := range voiceList {
		voices = append(voices, gnxaigc.TTSVoiceItem{
			VoiceName: voice.VoiceName,
			VoiceType: voice.VoiceType,
		})
	}

	g.availableVoices = voices
	fmt.Printf("Loaded %d available voices\n", len(g.availableVoices))
	return nil
}

func (g *ComicGenerator) processChapter(index int, chapter NovelChapter) error {
	fmt.Printf("\n=== Processing Chapter %d: %s ===\n", index+1, chapter.Title)

	chapterDir := filepath.Join(g.config.OutputDir, fmt.Sprintf("chapter_%03d", index+1))
	if err := os.MkdirAll(chapterDir, 0755); err != nil {
		return fmt.Errorf("creating chapter directory: %w", err)
	}

	summary, err := g.generateSummary(chapter)
	if err != nil {
		return err
	}

	g.updateCharacterRegistry(summary.CharacterFeatures)

	if err := g.syncCharacterAssets(chapterDir, summary); err != nil {
		fmt.Printf("Error syncing character assets: %v\n", err)
	}

	if err := g.writeStoryboard(chapterDir, summary); err != nil {
		fmt.Printf("%v\n", err)
	}

	if err := g.generateChapterPages(chapterDir, summary); err != nil {
		fmt.Printf("Error generating chapter %d pages: %v\n", index+1, err)
	}

	fmt.Printf("Chapter %d completed!\n", index+1)
	return nil
}

func (g *ComicGenerator) generateSummary(chapter NovelChapter) (*gnxaigc.SummaryChapterOutput, error) {
	fmt.Println("Generating storyboard...")

	existingFeatures := collectOrderedFeatures(g.characterOrder, g.characterRegistry)

	for attempt := 0; attempt < 3; attempt++ {
		summary, err := g.aigc.SummaryChapter(g.ctx, gnxaigc.SummaryChapterInput{
			NovelTitle:           g.config.NovelTitle,
			ChapterTitle:         chapter.Title,
			Content:              chapter.Content,
			AvailableVoiceStyles: g.availableVoices,
			CharacterFeatures:    existingFeatures,
		})
		if err != nil {
			fmt.Printf("  Error summarizing chapter: %v\n", err)
			fmt.Println("  Retrying...")
			continue
		}
		return summary, nil
	}

	return nil, fmt.Errorf("generating storyboard for chapter %q: failed after retries", chapter.Title)
}

func (g *ComicGenerator) updateCharacterRegistry(features []gnxaigc.CharacterFeature) {
	for _, feature := range features {
		key := normalizeCharacterKey(feature.Basic.Name)
		if key == "" {
			fmt.Println("  Warning: skip unnamed character in summary output")
			continue
		}
		if _, exists := g.characterRegistry[key]; !exists {
			g.characterOrder = append(g.characterOrder, key)
		}
		g.characterRegistry[key] = feature
	}
}

func (g *ComicGenerator) syncCharacterAssets(chapterDir string, summary *gnxaigc.SummaryChapterOutput) error {
	characterDir := filepath.Join(chapterDir, "characters")
	if err := os.MkdirAll(characterDir, 0755); err != nil {
		return fmt.Errorf("creating character directory: %w", err)
	}

	fmt.Printf("Syncing %d character concept images...\n", len(summary.CharacterFeatures))

	manifest := ChapterCharacterManifest{}
	trimmedStyle := strings.TrimSpace(g.config.ImageStyle)

	for _, feature := range summary.CharacterFeatures {
		key := normalizeCharacterKey(feature.Basic.Name)
		if key == "" {
			fmt.Println("  Skipping character with empty name in concept art stage")
			continue
		}

		globalIdx := findCharacterIndex(g.characterOrder, key)
		if globalIdx < 0 {
			globalIdx = len(g.characterOrder)
		}

		asset, exists := g.characterAssets[key]
		if !exists {
			asset = &CharacterAsset{}
		}

		fileStem := asset.FileStem
		if fileStem == "" {
			fileStem = sanitizeCharacterFileStem(feature.Basic.Name, globalIdx)
		}

		prompt := strings.TrimSpace(feature.ConceptArtPrompt)
		if prompt == "" {
			fmt.Printf("  [Character %d] Missing concept_art_prompt for %s, skip image generation.\n", globalIdx+1, feature.Basic.Name)
			asset.Feature = feature
			asset.FileStem = fileStem
			g.characterAssets[key] = asset

			manifest.Characters = append(manifest.Characters, ChapterCharacterEntry{
				Name:            feature.Basic.Name,
				ConceptArtNotes: feature.ConceptArtNotes,
			})
			continue
		}

		fullPrompt := prompt
		if trimmedStyle != "" {
			fullPrompt = fmt.Sprintf("%s %s", trimmedStyle, prompt)
		}
		fullPrompt = strings.TrimSpace(fullPrompt)

		globalImageFile := filepath.Join(g.globalCharacterDir, fmt.Sprintf("%s.png", fileStem))
		globalPromptFile := filepath.Join(g.globalCharacterDir, fmt.Sprintf("%s_prompt.txt", fileStem))

		shouldGenerate := strings.TrimSpace(asset.Prompt) == "" || strings.TrimSpace(asset.Feature.ConceptArtPrompt) != prompt || strings.TrimSpace(asset.Prompt) != fullPrompt
		if !shouldGenerate {
			if _, err := os.Stat(globalImageFile); err != nil {
				shouldGenerate = true
			}
		}

		var freshImage []byte
		if shouldGenerate {
			if asset.ImagePath != "" {
				baseData, err := os.ReadFile(asset.ImagePath)
				if err != nil {
					fmt.Printf("  [Character %d] Warning: cannot read existing concept art (%v), fallback to text generation.\n", globalIdx+1, err)
					freshImage, err = g.aigc.GenerateImageByText(g.ctx, fullPrompt)
					if err != nil {
						fmt.Printf("    Error generating concept art: %v\n", err)
						continue
					}
				} else {
					fmt.Printf("  [Character %d] Refining concept art via img2img for %s\n", globalIdx+1, feature.Basic.Name)
					freshImage, err = g.aigc.GenerateImageByImage(g.ctx, baseData, fullPrompt)
					if err != nil {
						fmt.Printf("    Error refining concept art: %v\n", err)
						fmt.Printf("    Falling back to text-to-image generation.\n")
						freshImage, err = g.aigc.GenerateImageByText(g.ctx, fullPrompt)
						if err != nil {
							fmt.Printf("    Error generating concept art: %v\n", err)
							continue
						}
					}
				}
			} else {
				fmt.Printf("  [Character %d] Generating concept art from scratch for %s\n", globalIdx+1, feature.Basic.Name)
				imageData, err := g.aigc.GenerateImageByText(g.ctx, fullPrompt)
				if err != nil {
					fmt.Printf("    Error generating concept art: %v\n", err)
					continue
				}
				freshImage = imageData
			}

			if err := os.WriteFile(globalImageFile, freshImage, 0644); err != nil {
				fmt.Printf("    Error saving concept art: %v\n", err)
				continue
			}
		} else {
			fmt.Printf("  [Character %d] Reusing existing concept art for %s\n", globalIdx+1, feature.Basic.Name)
		}

		if err := os.WriteFile(globalPromptFile, []byte(fullPrompt+"\n"), 0644); err != nil {
			fmt.Printf("    Error saving concept prompt: %v\n", err)
		}

		asset.Feature = feature
		asset.FileStem = fileStem
		asset.ImagePath = globalImageFile
		asset.Prompt = fullPrompt
		g.characterAssets[key] = asset

		chapterImageFile := filepath.Join(characterDir, fmt.Sprintf("%s.png", fileStem))
		if asset.ImagePath != "" {
			if err := copyFile(asset.ImagePath, chapterImageFile); err != nil {
				fmt.Printf("    Error copying concept art to chapter folder: %v\n", err)
			}
		}

		chapterPromptFile := filepath.Join(characterDir, fmt.Sprintf("%s_prompt.txt", fileStem))
		if err := os.WriteFile(chapterPromptFile, []byte(fullPrompt+"\n"), 0644); err != nil {
			fmt.Printf("    Error writing chapter prompt file: %v\n", err)
		}

		manifest.Characters = append(manifest.Characters, ChapterCharacterEntry{
			Name:            feature.Basic.Name,
			ImageFile:       filepath.Base(chapterImageFile),
			PromptFile:      filepath.Base(chapterPromptFile),
			GlobalImageFile: filepath.Base(globalImageFile),
			GlobalPrompt:    fullPrompt,
			ConceptArtNotes: feature.ConceptArtNotes,
		})
	}

	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling chapter character manifest: %w", err)
	}

	manifestPath := filepath.Join(characterDir, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestBytes, 0644); err != nil {
		return fmt.Errorf("writing character manifest: %w", err)
	}

	return nil
}

func (g *ComicGenerator) writeStoryboard(chapterDir string, summary *gnxaigc.SummaryChapterOutput) error {
	storyboardJSON, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling storyboard: %w", err)
	}

	storyboardFile := filepath.Join(chapterDir, "storyboard.json")
	if err := os.WriteFile(storyboardFile, storyboardJSON, 0644); err != nil {
		return fmt.Errorf("saving storyboard: %w", err)
	}

	fmt.Printf("Saved storyboard to %s\n", storyboardFile)
	return nil
}

func (g *ComicGenerator) generateChapterPages(chapterDir string, summary *gnxaigc.SummaryChapterOutput) error {
	totalPages := len(summary.StoryboardPages)
	fmt.Printf("Processing %d storyboard pages...\n", totalPages)

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		firstErr error
	)

	chapterFeatures := summary.CharacterFeatures

	for pageIndex, page := range summary.StoryboardPages {
		wg.Add(1)

		go func(pageIndex int, pageItem gnxaigc.StoryboardPage) {
			defer wg.Done()

			mu.Lock()
			fmt.Printf("  [Page %d/%d] Generating image...\n", pageIndex+1, totalPages)
			mu.Unlock()

			fullPrompt := gnxaigc.ComposePageImagePrompt(g.config.ImageStyle, pageItem)

			referenceKeys := collectPageCharacterKeys(pageItem, chapterFeatures)
			var referenceImages [][]byte
			for _, key := range referenceKeys {
				asset := g.characterAssets[key]
				if asset == nil || asset.ImagePath == "" {
					continue
				}
				data, err := os.ReadFile(asset.ImagePath)
				if err != nil {
					mu.Lock()
					fmt.Printf("    Warning: failed to read reference image for %s: %v\n", asset.Feature.Basic.Name, err)
					if firstErr == nil {
						firstErr = fmt.Errorf("reading reference image for %s: %w", asset.Feature.Basic.Name, err)
					}
					mu.Unlock()
					continue
				}
				referenceImages = append(referenceImages, data)
			}

			var (
				imageData []byte
				err       error
			)

			switch len(referenceImages) {
			case 0:
				imageData, err = g.aigc.GenerateImageByText(g.ctx, fullPrompt)
			case 1:
				mu.Lock()
				fmt.Printf("    Using single reference image for page %d\n", pageIndex+1)
				mu.Unlock()
				imageData, err = g.aigc.GenerateImageByImage(g.ctx, referenceImages[0], fullPrompt)
				if err != nil {
					mu.Lock()
					fmt.Printf("    Error generating page image via img2img: %v\n", err)
					fmt.Printf("    Falling back to text-to-image for page %d\n", pageIndex+1)
					mu.Unlock()
					imageData, err = g.aigc.GenerateImageByText(g.ctx, fullPrompt)
				}
			default:
				mu.Lock()
				fmt.Printf("    Using %d reference images for page %d\n", len(referenceImages), pageIndex+1)
				mu.Unlock()
				imageData, err = g.aigc.GenerateImageByImages(g.ctx, referenceImages, fullPrompt)
				if err != nil {
					mu.Lock()
					fmt.Printf("    Error generating page image via multi img2img: %v\n", err)
					fmt.Printf("    Falling back to text-to-image for page %d\n", pageIndex+1)
					mu.Unlock()
					imageData, err = g.aigc.GenerateImageByText(g.ctx, fullPrompt)
				}
			}

			if err != nil {
				mu.Lock()
				fmt.Printf("    Error generating image for page %d: %v\n", pageIndex+1, err)
				if firstErr == nil {
					firstErr = fmt.Errorf("generating image for page %d: %w", pageIndex+1, err)
				}
				mu.Unlock()
			} else {
				imageFile := filepath.Join(chapterDir, fmt.Sprintf("page_%03d.png", pageIndex+1))
				if err := os.WriteFile(imageFile, imageData, 0644); err != nil {
					mu.Lock()
					fmt.Printf("    Error saving image for page %d: %v\n", pageIndex+1, err)
					if firstErr == nil {
						firstErr = fmt.Errorf("saving image for page %d: %w", pageIndex+1, err)
					}
					mu.Unlock()
				} else {
					mu.Lock()
					fmt.Printf("    Saved image to %s\n", imageFile)
					mu.Unlock()
				}
			}

			var audioWg sync.WaitGroup

			for panelIdx, panel := range pageItem.Panels {
				segmentCount := len(panel.SourceTextSegments)

				for segIdx, segment := range panel.SourceTextSegments {
					audioWg.Add(1)

					go func(panelIndex, audioIndex, totalSegments int, audioSegment gnxaigc.SourceTextSegment) {
						defer audioWg.Done()

						mu.Lock()
						fmt.Printf("  [Page %d/%d] Panel %d audio %d/%d...\n", pageIndex+1, totalPages, panelIndex+1, audioIndex+1, totalSegments)
						mu.Unlock()

						audioData, err := g.aigc.TextToSpeechSimple(
							g.ctx,
							audioSegment.Text,
							audioSegment.VoiceType,
							audioSegment.SpeedRatio,
						)
						if err != nil {
							mu.Lock()
							fmt.Printf("    Error generating audio for page %d panel %d segment %d: %v\n", pageIndex+1, panelIndex+1, audioIndex+1, err)
							if firstErr == nil {
								firstErr = fmt.Errorf("generating audio for page %d panel %d segment %d: %w", pageIndex+1, panelIndex+1, audioIndex+1, err)
							}
							mu.Unlock()
							return
						}

						audioFile := filepath.Join(
							chapterDir,
							fmt.Sprintf(
								"page_%03d_panel_%02d_audio_%03d.mp3",
								pageIndex+1,
								panelIndex+1,
								audioIndex+1,
							),
						)

						if err := os.WriteFile(audioFile, audioData, 0644); err != nil {
							mu.Lock()
							fmt.Printf("    Error saving audio for page %d panel %d segment %d: %v\n", pageIndex+1, panelIndex+1, audioIndex+1, err)
							if firstErr == nil {
								firstErr = fmt.Errorf("saving audio for page %d panel %d segment %d: %w", pageIndex+1, panelIndex+1, audioIndex+1, err)
							}
							mu.Unlock()
						} else {
							mu.Lock()
							fmt.Printf("    Saved audio to %s\n", audioFile)
							mu.Unlock()
						}
					}(panelIdx, segIdx, segmentCount, segment)
				}
			}

			audioWg.Wait()
		}(pageIndex, page)
	}

	wg.Wait()

	return firstErr
}

func (g *ComicGenerator) writeGlobalManifest() error {
	globalManifest := ChapterCharacterManifest{}

	for _, key := range g.characterOrder {
		feature, ok := g.characterRegistry[key]
		if !ok {
			continue
		}

		asset := g.characterAssets[key]
		entry := ChapterCharacterEntry{
			Name:            feature.Basic.Name,
			ConceptArtNotes: feature.ConceptArtNotes,
		}

		if asset != nil {
			entry.GlobalImageFile = filepath.Base(asset.ImagePath)
			entry.GlobalPrompt = asset.Prompt
			if asset.FileStem != "" {
				entry.PromptFile = fmt.Sprintf("%s_prompt.txt", asset.FileStem)
			}
		}

		globalManifest.Characters = append(globalManifest.Characters, entry)
	}

	if len(globalManifest.Characters) == 0 {
		return nil
	}

	manifestBytes, err := json.MarshalIndent(globalManifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling global character manifest: %w", err)
	}

	manifestPath := filepath.Join(g.globalCharacterDir, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestBytes, 0644); err != nil {
		return fmt.Errorf("writing global character manifest: %w", err)
	}

	return nil
}
