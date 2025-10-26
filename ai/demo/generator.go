package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"qiniu-ai-image-generator/gnxaigc"
)

type ComicGeneratorConfig struct {
	NovelTitle string
	OutputDir  string
	ImageStyle string
}

type SlideshowAudio struct {
	File           string   `json:"file"`
	Text           string   `json:"text"`
	CharacterNames []string `json:"characterNames,omitempty"`
	PanelIndex     int      `json:"panelIndex"`
	SegmentIndex   int      `json:"segmentIndex"`
	IsNarration    bool     `json:"isNarration,omitempty"`
}

type SlideshowPage struct {
	PageIndex   int              `json:"pageIndex"`
	Image       string           `json:"image"`
	LayoutHint  string           `json:"layoutHint,omitempty"`
	ImagePrompt string           `json:"imagePrompt,omitempty"`
	Panels      int              `json:"panels"`
	Audio       []SlideshowAudio `json:"audio"`
}

type SlideshowChapter struct {
	Title  string          `json:"title"`
	Folder string          `json:"folder,omitempty"`
	Slides []SlideshowPage `json:"slides"`
}

type SlideshowDocument struct {
	NovelTitle string             `json:"novelTitle"`
	Chapters   []SlideshowChapter `json:"chapters"`
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
	slideshows         []SlideshowChapter
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
		slideshows:         make([]SlideshowChapter, 0),
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

	if err := g.writeRootSlideshow(); err != nil {
		fmt.Printf("Error writing slideshow index: %v\n", err)
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

	slidesForChapter := g.buildSlideshowPages(summary, chapterDir, "")
	if err := g.writeChapterSlideshow(chapterDir, chapter.Title, slidesForChapter); err != nil {
		fmt.Printf("Error writing slideshow for chapter %d: %v\n", index+1, err)
	}

	chapterFolder := filepath.Base(chapterDir)
	slidesForRoot := g.buildSlideshowPages(summary, chapterDir, chapterFolder)
	g.slideshows = append(g.slideshows, SlideshowChapter{
		Title:  chapter.Title,
		Folder: chapterFolder,
		Slides: slidesForRoot,
	})

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
				composite, mergeErr := mergeImagesSideBySide(referenceImages)
				if mergeErr != nil {
					mu.Lock()
					fmt.Printf("    Warning: failed to merge %d reference images for page %d: %v\n", len(referenceImages), pageIndex+1, mergeErr)
					mu.Unlock()
					imageData, err = g.aigc.GenerateImageByText(g.ctx, fullPrompt)
					break
				}

				mu.Lock()
				fmt.Printf("    Merged %d reference images for page %d\n", len(referenceImages), pageIndex+1)
				mu.Unlock()
				imageData, err = g.aigc.GenerateImageByImage(g.ctx, composite, fullPrompt)
				if err != nil {
					mu.Lock()
					fmt.Printf("    Error generating page image via merged img2img: %v\n", err)
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

						const maxTTSAudioRetries = 3
						var (
							audioData []byte
							err       error
						)

						for attempt := 1; attempt <= maxTTSAudioRetries; attempt++ {
							audioData, err = g.aigc.TextToSpeechSimple(
								g.ctx,
								audioSegment.Text,
								audioSegment.VoiceType,
								audioSegment.SpeedRatio,
							)
							if err == nil {
								break
							}

							if attempt == maxTTSAudioRetries || !isRateLimitError(err) {
								break
							}

							sleep := time.Second * time.Duration(attempt)
							mu.Lock()
							fmt.Printf("    Rate limited generating audio for page %d panel %d segment %d, retrying in %s\n", pageIndex+1, panelIndex+1, audioIndex+1, sleep)
							mu.Unlock()
							time.Sleep(sleep)
						}

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

func (g *ComicGenerator) buildSlideshowPages(summary *gnxaigc.SummaryChapterOutput, chapterDir, pathPrefix string) []SlideshowPage {
	slides := make([]SlideshowPage, 0, len(summary.StoryboardPages))
	cleanPrefix := strings.Trim(pathPrefix, "/")

	for pageIndex, page := range summary.StoryboardPages {
		imageName := fmt.Sprintf("page_%03d.png", pageIndex+1)
		imagePath := filepath.Join(chapterDir, imageName)

		relativeImage := ""
		if _, err := os.Stat(imagePath); err == nil {
			relativeImage = joinForHTML(cleanPrefix, imageName)
		}

		slide := SlideshowPage{
			PageIndex:   pageIndex,
			Image:       relativeImage,
			LayoutHint:  page.LayoutHint,
			ImagePrompt: page.ImagePrompt,
			Panels:      len(page.Panels),
			Audio:       make([]SlideshowAudio, 0),
		}

		for panelIndex, panel := range page.Panels {
			for segmentIndex, segment := range panel.SourceTextSegments {
				audioName := fmt.Sprintf(
					"page_%03d_panel_%02d_audio_%03d.mp3",
					pageIndex+1,
					panelIndex+1,
					segmentIndex+1,
				)
				audioPath := filepath.Join(chapterDir, audioName)
				if _, err := os.Stat(audioPath); err != nil {
					continue
				}

				audio := SlideshowAudio{
					File:         joinForHTML(cleanPrefix, audioName),
					Text:         segment.Text,
					PanelIndex:   panelIndex,
					SegmentIndex: segmentIndex,
					IsNarration:  segment.IsNarration,
				}

				if len(segment.CharacterNames) > 0 {
					audio.CharacterNames = append([]string(nil), segment.CharacterNames...)
				}

				slide.Audio = append(slide.Audio, audio)
			}
		}

		slides = append(slides, slide)
	}

	return slides
}

func joinForHTML(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return path.Join(prefix, name)
}

func (g *ComicGenerator) writeChapterSlideshow(chapterDir, chapterTitle string, slides []SlideshowPage) error {
	doc := SlideshowDocument{
		NovelTitle: g.config.NovelTitle,
		Chapters: []SlideshowChapter{
			{
				Title:  chapterTitle,
				Slides: slides,
			},
		},
	}

	target := filepath.Join(chapterDir, "slideshow.html")
	pageTitle := fmt.Sprintf("%s - %s 漫画预览", g.config.NovelTitle, chapterTitle)
	return g.writeSlideshowHTML(target, doc, pageTitle, 0)
}

func (g *ComicGenerator) writeRootSlideshow() error {
	if len(g.slideshows) == 0 {
		return nil
	}

	doc := SlideshowDocument{
		NovelTitle: g.config.NovelTitle,
		Chapters:   make([]SlideshowChapter, 0, len(g.slideshows)),
	}

	doc.Chapters = append(doc.Chapters, g.slideshows...)

	target := filepath.Join(g.config.OutputDir, "index.html")
	pageTitle := fmt.Sprintf("%s - 漫画预览", g.config.NovelTitle)
	return g.writeSlideshowHTML(target, doc, pageTitle, 0)
}

func (g *ComicGenerator) writeSlideshowHTML(targetPath string, doc SlideshowDocument, pageTitle string, initialChapterIndex int) error {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("ensuring directory for slideshow html: %w", err)
	}

	dataBytes, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling slideshow data: %w", err)
	}
	dataBytes = bytes.ReplaceAll(dataBytes, []byte("</script>"), []byte("<\\/script>"))

	config := map[string]any{
		"initialChapterIndex": initialChapterIndex,
		"autoStart":           true,
		"pageAdvanceDelayMs":  5000,
	}
	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshalling slideshow config: %w", err)
	}
	configBytes = bytes.ReplaceAll(configBytes, []byte("</script>"), []byte("<\\/script>"))

	var builder strings.Builder
	builder.WriteString("<!DOCTYPE html>\n")
	builder.WriteString("<html lang=\"zh-CN\">\n")
	builder.WriteString("<head>\n")
	builder.WriteString("  <meta charset=\"UTF-8\">\n")
	builder.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	builder.WriteString("  <title>")
	builder.WriteString(template.HTMLEscapeString(pageTitle))
	builder.WriteString("</title>\n")
	builder.WriteString("  <style>\n")
	builder.WriteString(`:root {
	color-scheme: light dark;
	font-family: "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
	background: #0f172a;
	color: #e2e8f0;
	line-height: 1.6;
}

body {
	margin: 0;
	min-height: 100vh;
	background: radial-gradient(circle at top left, rgba(56, 189, 248, 0.18), transparent 55%),
							radial-gradient(circle at bottom right, rgba(249, 168, 212, 0.18), transparent 50%),
							#0f172a;
}

.app {
	display: flex;
	flex-direction: column;
	min-height: 100vh;
}

.app__header {
	display: flex;
	flex-wrap: wrap;
	justify-content: space-between;
	gap: 1rem;
	padding: 1.5rem 2rem;
	border-bottom: 1px solid rgba(148, 163, 184, 0.2);
	backdrop-filter: blur(12px);
	background: rgba(15, 23, 42, 0.6);
}

.app__info h1 {
	margin: 0;
	font-size: 1.6rem;
}

.app__chapter {
	margin: 0.35rem 0 0;
	font-size: 1.05rem;
	color: rgba(226, 232, 240, 0.75);
}

.status-bar {
	padding: 0.8rem 1rem;
	border-radius: 0.8rem;
	background: rgba(148, 163, 184, 0.15);
	color: rgba(226, 232, 240, 0.85);
	min-width: 220px;
}

.app__main {
	flex: 1;
	padding: 2rem;
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 1.5rem;
	text-align: center;
}

.slide-frame {
	position: relative;
	display: flex;
	justify-content: center;
	align-items: center;
	width: min(92vw, 1080px);
	max-height: 75vh;
	background: rgba(15, 23, 42, 0.65);
	border-radius: 1.2rem;
	border: 1px solid rgba(148, 163, 184, 0.25);
	overflow: hidden;
	box-shadow: 0 35px 120px rgba(15, 23, 42, 0.45);
}

.slide-image {
	width: 100%;
	height: auto;
	object-fit: contain;
}

.slide-image--missing {
	opacity: 0.32;
}

.slide-counter {
	position: absolute;
	bottom: 1rem;
	right: 1.2rem;
	padding: 0.45rem 0.85rem;
	border-radius: 999px;
	background: rgba(15, 23, 42, 0.7);
	border: 1px solid rgba(148, 163, 184, 0.35);
	font-size: 0.9rem;
}

.subtitle {
	font-size: 1.2rem;
	max-width: 80ch;
	padding: 0.9rem 1.2rem;
	border-radius: 0.9rem;
	background: rgba(30, 41, 59, 0.7);
	border: 1px solid rgba(148, 163, 184, 0.2);
	min-height: 3.2rem;
}

.meta {
	font-size: 0.95rem;
	color: rgba(226, 232, 240, 0.7);
	max-width: 90ch;
}

.app__footer {
	padding: 1.4rem 2rem 2.2rem;
	display: flex;
	flex-wrap: wrap;
	gap: 1rem;
	justify-content: space-between;
	align-items: center;
}

.control-select {
	padding: 0.65rem 1rem;
	border-radius: 0.75rem;
	border: 1px solid rgba(148, 163, 184, 0.35);
	background: rgba(15, 23, 42, 0.65);
	color: #e2e8f0;
}

.control-buttons {
	display: flex;
	gap: 0.75rem;
}

.control-btn {
	padding: 0.75rem 1.2rem;
	border-radius: 0.85rem;
	border: 1px solid rgba(59, 130, 246, 0.4);
	background: linear-gradient(145deg, rgba(59, 130, 246, 0.9), rgba(99, 102, 241, 0.9));
	color: #fff;
	font-weight: 600;
	cursor: pointer;
	transition: transform 0.15s ease, box-shadow 0.2s ease;
}

.control-btn:hover {
	transform: translateY(-2px);
	box-shadow: 0 18px 45px rgba(59, 130, 246, 0.35);
}

.control-btn:active {
	transform: translateY(0);
}

.hidden {
	display: none !important;
}

@media (max-width: 900px) {
	.app__main {
		padding: 1.5rem;
	}

	.slide-frame {
		width: min(96vw, 720px);
	}

	.subtitle {
		font-size: 1.05rem;
	}
}

@media (max-width: 600px) {
	.app__header, .app__footer {
		flex-direction: column;
		align-items: stretch;
	}

	.control-buttons {
		width: 100%;
		justify-content: space-between;
	}

	.control-btn {
		flex: 1;
	}
}
`)
	builder.WriteString("  </style>\n")
	builder.WriteString("</head>\n")
	builder.WriteString("<body>\n")
	builder.WriteString("  <div class=\"app\">\n")
	builder.WriteString("    <header class=\"app__header\">\n")
	builder.WriteString("      <div class=\"app__info\">\n")
	builder.WriteString("        <h1 id=\"novel-title\">")
	builder.WriteString(template.HTMLEscapeString(doc.NovelTitle))
	builder.WriteString("</h1>\n")
	builder.WriteString("        <p id=\"chapter-title\" class=\"app__chapter\"></p>\n")
	builder.WriteString("      </div>\n")
	builder.WriteString("      <div id=\"status\" class=\"status-bar\">准备播放</div>\n")
	builder.WriteString("    </header>\n")
	builder.WriteString("    <main class=\"app__main\">\n")
	builder.WriteString("      <div class=\"slide-frame\">\n")
	builder.WriteString("        <img id=\"slide-image\" class=\"slide-image\" alt=\"漫画页面预览\">\n")
	builder.WriteString("        <div id=\"slide-counter\" class=\"slide-counter\"></div>\n")
	builder.WriteString("      </div>\n")
	builder.WriteString("      <div id=\"subtitle\" class=\"subtitle\"></div>\n")
	builder.WriteString("      <div id=\"meta\" class=\"meta\"></div>\n")
	builder.WriteString("    </main>\n")
	builder.WriteString("    <footer class=\"app__footer\">\n")
	builder.WriteString("      <select id=\"chapter-select\" class=\"control-select hidden\"></select>\n")
	builder.WriteString("      <div class=\"control-buttons\">\n")
	builder.WriteString("        <button id=\"prev-btn\" class=\"control-btn\">上一页</button>\n")
	builder.WriteString("        <button id=\"restart-btn\" class=\"control-btn\">重新播放</button>\n")
	builder.WriteString("        <button id=\"next-btn\" class=\"control-btn\">下一页</button>\n")
	builder.WriteString("      </div>\n")
	builder.WriteString("    </footer>\n")
	builder.WriteString("  </div>\n")
	builder.WriteString("  <audio id=\"slide-audio\" preload=\"auto\"></audio>\n")
	builder.WriteString("  <script type=\"application/json\" id=\"slideshow-data\">")
	builder.Write(dataBytes)
	builder.WriteString("</script>\n")
	builder.WriteString("  <script type=\"application/json\" id=\"slideshow-config\">")
	builder.Write(configBytes)
	builder.WriteString("</script>\n")
	builder.WriteString(`  <script>
(function () {
	const dataEl = document.getElementById("slideshow-data");
	if (!dataEl) {
		return;
	}

	let data;
	try {
		data = JSON.parse(dataEl.textContent || "{}");
	} catch (error) {
		console.error("Failed to parse slideshow data", error);
		return;
	}

	const configEl = document.getElementById("slideshow-config");
	const config = {
		initialChapterIndex: 0,
		autoStart: true,
		pageAdvanceDelayMs: 5000
	};

	if (configEl) {
		try {
			const parsed = JSON.parse(configEl.textContent || "{}");
			Object.assign(config, parsed);
		} catch (error) {
			console.warn("Failed to parse slideshow config", error);
		}
	}

	if (!Array.isArray(data.chapters) || !data.chapters.length) {
		const statusEl = document.getElementById("status");
		if (statusEl) {
			statusEl.textContent = "暂无章节可播放";
		}
		return;
	}

	const novelTitleEl = document.getElementById("novel-title");
	const chapterTitleEl = document.getElementById("chapter-title");
	const slideCounterEl = document.getElementById("slide-counter");
	const imageEl = document.getElementById("slide-image");
	const subtitleEl = document.getElementById("subtitle");
	const metaEl = document.getElementById("meta");
	const statusEl = document.getElementById("status");
	const chapterSelectEl = document.getElementById("chapter-select");
	const prevBtn = document.getElementById("prev-btn");
	const nextBtn = document.getElementById("next-btn");
	const restartBtn = document.getElementById("restart-btn");
	const audioEl = document.getElementById("slide-audio");

	let chapterIndex = Number(config.initialChapterIndex) || 0;
	if (chapterIndex < 0) chapterIndex = 0;
	if (chapterIndex >= data.chapters.length) chapterIndex = data.chapters.length - 1;

	let slideIndex = 0;
	let audioIndex = 0;
	let autoTimer = null;
	let awaitingGesture = false;
	const autoDelay = Math.max(Number(config.pageAdvanceDelayMs) || 5000, 1000);

	function setStatus(message) {
		if (statusEl) {
			statusEl.textContent = message;
		}
	}

	function currentChapter() {
		return data.chapters[chapterIndex] || null;
	}

	function currentSlide() {
		const chapter = currentChapter();
		if (!chapter || !Array.isArray(chapter.slides)) {
			return null;
		}
		return chapter.slides[slideIndex] || null;
	}

	function renderChapterOptions() {
		if (!chapterSelectEl) {
			return;
		}
		chapterSelectEl.innerHTML = "";
		if (!Array.isArray(data.chapters) || data.chapters.length <= 1) {
			chapterSelectEl.classList.add("hidden");
			return;
		}
		data.chapters.forEach(function (chap, idx) {
			const option = document.createElement("option");
			option.value = String(idx);
			option.textContent = chap && chap.title ? chap.title : "章节 " + (idx + 1);
			chapterSelectEl.appendChild(option);
		});
		chapterSelectEl.value = String(chapterIndex);
		chapterSelectEl.classList.remove("hidden");
	}

	function updateChapterHeading() {
		const chapter = currentChapter();
		if (chapterTitleEl) {
			chapterTitleEl.textContent = chapter && chapter.title ? chapter.title : "章节 " + (chapterIndex + 1);
		}
		if (chapterSelectEl) {
			chapterSelectEl.value = String(chapterIndex);
		}
	}

	function clearAutoTimer() {
		if (autoTimer) {
			window.clearTimeout(autoTimer);
			autoTimer = null;
		}
	}

	function handleAutoplayBlocked() {
		if (awaitingGesture) {
			return;
		}
		awaitingGesture = true;
		setStatus("浏览器阻止自动播放，请点击页面继续。");
		const resume = function () {
			awaitingGesture = false;
			document.body.removeEventListener("click", resume);
			playCurrentAudio(true);
		};
		document.body.addEventListener("click", resume, { once: true });
	}

	function buildMeta(slide) {
		if (!slide) {
			return "";
		}
		const parts = [];
		if (slide.layoutHint) {
			parts.push("排版提示: " + slide.layoutHint);
		}
		if (slide.panels) {
			parts.push("分镜数量: " + slide.panels);
		}
		if (slide.imagePrompt) {
			parts.push("图像提示: " + slide.imagePrompt);
		}
		return parts.join(" · ");
	}

	function updateMeta(slide) {
		if (metaEl) {
			metaEl.textContent = buildMeta(slide);
		}
	}

	function updateSubtitle(clip) {
		if (!subtitleEl) {
			return;
		}
		if (!clip) {
			subtitleEl.textContent = "";
			return;
		}
		const names = Array.isArray(clip.characterNames) ? clip.characterNames.filter(Boolean) : [];
		const prefix = names.length ? names.join(" / ") + ": " : "";
		subtitleEl.textContent = prefix + (clip.text || "");
	}

	function updateStatusForAudio(slide) {
		if (!slide) {
			return;
		}
		const total = Array.isArray(slide.audio) ? slide.audio.length : 0;
		const base = "章节 " + (chapterIndex + 1) + " · 第 " + (slideIndex + 1) + " 页";
		if (total) {
			setStatus(base + " · 音频 " + (audioIndex + 1) + "/" + total);
		} else {
			setStatus(base);
		}
	}

	function ensureImage(slide) {
		if (!imageEl) {
			return;
		}
		if (slide && slide.image) {
			imageEl.src = slide.image;
			imageEl.alt = "章节 " + (chapterIndex + 1) + " - 第 " + (slideIndex + 1) + " 页";
			imageEl.classList.remove("slide-image--missing");
		} else {
			imageEl.removeAttribute("src");
			imageEl.alt = "未找到页面图像";
			imageEl.classList.add("slide-image--missing");
		}
	}

	function scheduleNextSlide() {
		clearAutoTimer();
		autoTimer = window.setTimeout(function () {
			advanceSlide(false);
		}, autoDelay);
	}

	function showSlide(targetIndex) {
		const chapter = currentChapter();
		if (!chapter || !Array.isArray(chapter.slides) || !chapter.slides.length) {
			ensureImage(null);
			updateSubtitle(null);
			updateMeta(null);
			if (slideCounterEl) {
				slideCounterEl.textContent = "";
			}
			setStatus("章节 " + (chapterIndex + 1) + " 暂无页面");
			return;
		}

		if (targetIndex < 0) targetIndex = 0;
		if (targetIndex >= chapter.slides.length) {
			advanceChapter(false);
			return;
		}

		clearAutoTimer();
		slideIndex = targetIndex;
		audioIndex = 0;

		if (audioEl) {
			audioEl.pause();
			audioEl.currentTime = 0;
			audioEl.removeAttribute("src");
		}

		const slide = currentSlide();
		ensureImage(slide);
		updateMeta(slide);
		updateSubtitle(null);

		if (slideCounterEl) {
			slideCounterEl.textContent = "第 " + (slideIndex + 1) + " / " + chapter.slides.length + " 页";
		}

		if (slide && Array.isArray(slide.audio) && slide.audio.length) {
			playCurrentAudio(false);
		} else {
			setStatus("章节 " + (chapterIndex + 1) + " · 第 " + (slideIndex + 1) + " 页 · 无音频，将在 " + Math.round(autoDelay / 1000) + " 秒后自动翻页");
			scheduleNextSlide();
		}
	}

	function playCurrentAudio(isResume) {
		const slide = currentSlide();
		if (!slide || !Array.isArray(slide.audio) || !slide.audio.length) {
			scheduleNextSlide();
			return;
		}

		if (audioIndex >= slide.audio.length) {
			scheduleNextSlide();
			return;
		}

		const clip = slide.audio[audioIndex];
		updateSubtitle(clip);
		updateStatusForAudio(slide);

		if (!audioEl || !clip.file) {
			audioIndex += 1;
			playCurrentAudio(false);
			return;
		}

		audioEl.src = clip.file;
		const playPromise = audioEl.play();
		if (playPromise && typeof playPromise.then === "function") {
			playPromise.catch(function () {
				if (isResume) {
					scheduleNextSlide();
				} else {
					handleAutoplayBlocked();
				}
			});
		}
	}

	if (audioEl) {
		audioEl.addEventListener("ended", function () {
			audioIndex += 1;
			playCurrentAudio(false);
		});

		audioEl.addEventListener("error", function () {
			audioIndex += 1;
			playCurrentAudio(false);
		});
	}

	function advanceSlide(manual) {
		const chapter = currentChapter();
		if (!chapter || !Array.isArray(chapter.slides) || !chapter.slides.length) {
			return;
		}

		if (slideIndex + 1 < chapter.slides.length) {
			showSlide(slideIndex + 1);
		} else {
			advanceChapter(manual);
		}
	}

	function retreatSlide() {
		const chapter = currentChapter();
		if (!chapter || !Array.isArray(chapter.slides) || !chapter.slides.length) {
			return;
		}

		if (slideIndex > 0) {
			showSlide(slideIndex - 1);
			return;
		}

		if (chapterIndex > 0) {
			chapterIndex -= 1;
			updateChapterHeading();
			const prevChapter = currentChapter();
			const lastIndex = prevChapter && Array.isArray(prevChapter.slides) ? prevChapter.slides.length - 1 : 0;
			showSlide(Math.max(lastIndex, 0));
		}
	}

	function advanceChapter(manual) {
		if (chapterIndex + 1 < data.chapters.length) {
			chapterIndex += 1;
			slideIndex = 0;
			audioIndex = 0;
			updateChapterHeading();
			renderChapterOptions();
			showSlide(0);
		} else if (manual) {
			setStatus("已经是最后一章");
			clearAutoTimer();
		} else {
			setStatus("所有章节播放完毕");
			clearAutoTimer();
		}
	}

	function restartShow() {
		chapterIndex = 0;
		slideIndex = 0;
		audioIndex = 0;
		renderChapterOptions();
		updateChapterHeading();
		showSlide(0);
	}

	renderChapterOptions();
	updateChapterHeading();
	showSlide(0);

	if (chapterSelectEl) {
		chapterSelectEl.addEventListener("change", function (event) {
			const value = parseInt(event.target.value, 10);
			if (!Number.isNaN(value)) {
				chapterIndex = Math.min(Math.max(value, 0), data.chapters.length - 1);
				slideIndex = 0;
				audioIndex = 0;
				updateChapterHeading();
				showSlide(0);
			}
		});
	}

	if (prevBtn) {
		prevBtn.addEventListener("click", function () {
			retreatSlide();
		});
	}

	if (nextBtn) {
		nextBtn.addEventListener("click", function () {
			advanceSlide(true);
		});
	}

	if (restartBtn) {
		restartBtn.addEventListener("click", function () {
			restartShow();
		});
	}

	document.addEventListener("keydown", function (event) {
		if (event.defaultPrevented) {
			return;
		}
		if (event.key === "ArrowRight" || event.key === " ") {
			event.preventDefault();
			advanceSlide(true);
		} else if (event.key === "ArrowLeft") {
			event.preventDefault();
			retreatSlide();
		}
	});
})();
	</script>
`)
	builder.WriteString("</body>\n")
	builder.WriteString("</html>\n")

	return os.WriteFile(targetPath, []byte(builder.String()), 0644)
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

func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "429") || strings.Contains(msg, "rate limit")
}
