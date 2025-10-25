package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode"

	"qiniu-ai-image-generator/gnxaigc"
)

// CharacterAsset tracks the persisted concept art for a role to keep cross-chapter consistency.
type CharacterAsset struct {
	Feature   gnxaigc.CharacterFeature
	ImagePath string
	Prompt    string
	FileStem  string
}

type ChapterCharacterEntry struct {
	Name            string `json:"name"`
	ImageFile       string `json:"image_file,omitempty"`
	PromptFile      string `json:"prompt_file,omitempty"`
	GlobalImageFile string `json:"global_image_file,omitempty"`
	GlobalPrompt    string `json:"global_prompt,omitempty"`
	ConceptArtNotes string `json:"concept_art_notes,omitempty"`
}

type ChapterCharacterManifest struct {
	Characters []ChapterCharacterEntry `json:"characters"`
}

func normalizeCharacterKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func collectOrderedFeatures(order []string, registry map[string]gnxaigc.CharacterFeature) []gnxaigc.CharacterFeature {
	features := make([]gnxaigc.CharacterFeature, 0, len(order))
	for _, key := range order {
		if feature, ok := registry[key]; ok {
			features = append(features, feature)
		}
	}
	return features
}

func findCharacterIndex(order []string, key string) int {
	for idx, existing := range order {
		if existing == key {
			return idx
		}
	}
	return -1
}

func copyFile(src, dst string) error {
	if src == "" || dst == "" {
		return fmt.Errorf("copyFile: empty path")
	}
	if src == dst {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func collectSegmentCharacterKeys(segment gnxaigc.SourceTextSegment, chapterFeatures []gnxaigc.CharacterFeature, seen map[string]struct{}) {
	appendName := func(name string) {
		key := normalizeCharacterKey(name)
		if key == "" {
			return
		}
		seen[key] = struct{}{}
	}

	for _, name := range segment.CharacterNames {
		appendName(name)
	}

	for _, idx := range segment.CharacterRefs {
		if idx < 0 || idx >= len(chapterFeatures) {
			continue
		}
		appendName(chapterFeatures[idx].Basic.Name)
	}
}

func collectPageCharacterKeys(page gnxaigc.StoryboardPage, chapterFeatures []gnxaigc.CharacterFeature) []string {
	seen := make(map[string]struct{})
	for _, panel := range page.Panels {
		for _, segment := range panel.SourceTextSegments {
			collectSegmentCharacterKeys(segment, chapterFeatures, seen)
		}
	}

	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func main() {
	inputFile := flag.String("input", "", "Path to novel text file")
	outputDir := flag.String("output", "output", "Output directory for generated comic assets")
	novelTitle := flag.String("title", "未知小说", "Novel title")
	maxChapters := flag.Int("max-chapters", 0, "Maximum number of chapters to process (0 for all)")
	imageStyle := flag.String("image-style", "卡通风格，", "Image style prompt prefix to prepend to each scene's image prompt")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Error: -input flag is required")
		flag.Usage()
		os.Exit(1)
	}

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	aigc := gnxaigc.NewGnxAIGC(gnxaigc.Config{})

	chapters, err := SplitChaptersFromFile(*inputFile)
	if err != nil {
		fmt.Printf("Error reading novel file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully split novel into %d chapters\n", len(chapters))

	fmt.Println("Fetching available TTS voices...")
	voiceList, err := aigc.GetVoiceList(context.Background())
	if err != nil {
		fmt.Printf("Error fetching voice list: %v\n", err)
		os.Exit(1)
	}

	var availableVoices []gnxaigc.TTSVoiceItem
	for _, voice := range voiceList {
		availableVoices = append(availableVoices, gnxaigc.TTSVoiceItem{
			VoiceName: voice.VoiceName,
			VoiceType: voice.VoiceType,
		})
	}
	fmt.Printf("Loaded %d available voices\n", len(availableVoices))

	characterRegistry := make(map[string]gnxaigc.CharacterFeature)
	characterOrder := make([]string, 0)
	characterAssets := make(map[string]*CharacterAsset)

	globalCharacterDir := filepath.Join(*outputDir, "characters")
	if err := os.MkdirAll(globalCharacterDir, 0755); err != nil {
		fmt.Printf("Error creating shared character directory: %v\n", err)
		os.Exit(1)
	}

	processCount := len(chapters)
	if *maxChapters > 0 && *maxChapters < processCount {
		processCount = *maxChapters
	}

	for i := 0; i < processCount; i++ {
		chapter := chapters[i]
		fmt.Printf("\n=== Processing Chapter %d: %s ===\n", i+1, chapter.Title)

		chapterDir := filepath.Join(*outputDir, fmt.Sprintf("chapter_%03d", i+1))
		if err := os.MkdirAll(chapterDir, 0755); err != nil {
			fmt.Printf("Error creating chapter directory: %v\n", err)
			continue
		}

		fmt.Println("Generating storyboard...")

		existingFeatures := collectOrderedFeatures(characterOrder, characterRegistry)

		summaryChapter := func(existing []gnxaigc.CharacterFeature) (*gnxaigc.SummaryChapterOutput, error) {
			for attempt := 0; attempt < 3; attempt++ {
				summary, err := aigc.SummaryChapter(context.Background(), gnxaigc.SummaryChapterInput{
					NovelTitle:           *novelTitle,
					ChapterTitle:         chapter.Title,
					Content:              chapter.Content,
					AvailableVoiceStyles: availableVoices,
					CharacterFeatures:    existing,
				})
				if err != nil {
					fmt.Printf("  Error summarizing chapter: %v\n", err)
					fmt.Println("  Retrying...")
					continue
				}
				return summary, nil
			}
			return nil, fmt.Errorf("failed to summarize chapter after retries")
		}
		summary, err := summaryChapter(existingFeatures)
		if err != nil {
			fmt.Printf("Error generating storyboard for chapter %d: %v\n", i+1, err)
			continue
		}

		for _, feature := range summary.CharacterFeatures {
			key := normalizeCharacterKey(feature.Basic.Name)
			if key == "" {
				fmt.Println("  Warning: skip unnamed character in summary output")
				continue
			}
			if _, exists := characterRegistry[key]; !exists {
				characterOrder = append(characterOrder, key)
			}
			characterRegistry[key] = feature
		}

		characterDir := filepath.Join(chapterDir, "characters")
		if err := os.MkdirAll(characterDir, 0755); err != nil {
			fmt.Printf("Error creating character directory: %v\n", err)
		} else {
			fmt.Printf("Syncing %d character concept images...\n", len(summary.CharacterFeatures))
			manifest := ChapterCharacterManifest{}
			for _, feature := range summary.CharacterFeatures {
				key := normalizeCharacterKey(feature.Basic.Name)
				if key == "" {
					fmt.Println("  Skipping character with empty name in concept art stage")
					continue
				}

				globalIdx := findCharacterIndex(characterOrder, key)
				if globalIdx < 0 {
					globalIdx = len(characterOrder)
				}

				asset, hasAsset := characterAssets[key]
				if !hasAsset {
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
					characterAssets[key] = asset
					manifest.Characters = append(manifest.Characters, ChapterCharacterEntry{
						Name:            feature.Basic.Name,
						ConceptArtNotes: feature.ConceptArtNotes,
					})
					continue
				}

				trimmedStyle := strings.TrimSpace(*imageStyle)
				fullPrompt := prompt
				if trimmedStyle != "" {
					fullPrompt = fmt.Sprintf("%s %s", trimmedStyle, prompt)
				}
				fullPrompt = strings.TrimSpace(fullPrompt)

				globalImageFile := filepath.Join(globalCharacterDir, fmt.Sprintf("%s.png", fileStem))
				globalPromptFile := filepath.Join(globalCharacterDir, fmt.Sprintf("%s_prompt.txt", fileStem))

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
							freshImage, err = aigc.GenerateImageByText(context.Background(), fullPrompt)
							if err != nil {
								fmt.Printf("    Error generating concept art: %v\n", err)
								continue
							}
						} else {
							fmt.Printf("  [Character %d] Refining concept art via img2img for %s\n", globalIdx+1, feature.Basic.Name)
							freshImage, err = aigc.GenerateImageByImage(context.Background(), baseData, fullPrompt)
							if err != nil {
								fmt.Printf("    Error refining concept art: %v\n", err)
								fmt.Printf("    Falling back to text-to-image generation.\n")
								freshImage, err = aigc.GenerateImageByText(context.Background(), fullPrompt)
								if err != nil {
									fmt.Printf("    Error generating concept art: %v\n", err)
									continue
								}
							}
						}
					} else {
						fmt.Printf("  [Character %d] Generating concept art from scratch for %s\n", globalIdx+1, feature.Basic.Name)
						imageData, err := aigc.GenerateImageByText(context.Background(), fullPrompt)
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
				characterAssets[key] = asset

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
				fmt.Printf("    Error marshalling chapter character manifest: %v\n", err)
			} else {
				manifestPath := filepath.Join(characterDir, "manifest.json")
				if err := os.WriteFile(manifestPath, manifestBytes, 0644); err != nil {
					fmt.Printf("    Error writing character manifest: %v\n", err)
				}
			}
		}

		storyboardJSON, _ := json.MarshalIndent(summary, "", "  ")
		storyboardFile := filepath.Join(chapterDir, "storyboard.json")
		if err := os.WriteFile(storyboardFile, storyboardJSON, 0644); err != nil {
			fmt.Printf("Error saving storyboard: %v\n", err)
		} else {
			fmt.Printf("Saved storyboard to %s\n", storyboardFile)
		}

		totalPages := len(summary.StoryboardPages)
		fmt.Printf("Processing %d storyboard pages...\n", totalPages)

		var wg sync.WaitGroup
		var mu sync.Mutex

		for j, page := range summary.StoryboardPages {
			wg.Add(1)
			go func(pageIndex int, pageItem gnxaigc.StoryboardPage, total int, chapterFeatures []gnxaigc.CharacterFeature) {
				defer wg.Done()

				mu.Lock()
				fmt.Printf("  [Page %d/%d] Generating image...\n", pageIndex+1, total)
				mu.Unlock()

				fullPrompt := gnxaigc.ComposePageImagePrompt(*imageStyle, pageItem)

				referenceKeys := collectPageCharacterKeys(pageItem, chapterFeatures)
				var referenceImages [][]byte
				for _, key := range referenceKeys {
					asset := characterAssets[key]
					if asset == nil || asset.ImagePath == "" {
						continue
					}
					data, err := os.ReadFile(asset.ImagePath)
					if err != nil {
						mu.Lock()
						fmt.Printf("    Warning: failed to read reference image for %s: %v\n", asset.Feature.Basic.Name, err)
						mu.Unlock()
						continue
					}
					referenceImages = append(referenceImages, data)
				}

				var imageData []byte
				var err error
				if len(referenceImages) > 1 {
					mu.Lock()
					fmt.Printf("    Using %d reference images for page %d\n", len(referenceImages), pageIndex+1)
					mu.Unlock()
					imageData, err = aigc.GenerateImageByImages(context.Background(), referenceImages, fullPrompt)
					if err != nil {
						mu.Lock()
						fmt.Printf("    Error generating page image via multi img2img: %v\n", err)
						fmt.Printf("    Falling back to text-to-image for page %d\n", pageIndex+1)
						mu.Unlock()
						imageData, err = aigc.GenerateImageByText(context.Background(), fullPrompt)
					}
				} else if len(referenceImages) == 1 {
					mu.Lock()
					fmt.Printf("    Using single reference image for page %d\n", pageIndex+1)
					mu.Unlock()
					imageData, err = aigc.GenerateImageByImage(context.Background(), referenceImages[0], fullPrompt)
					if err != nil {
						mu.Lock()
						fmt.Printf("    Error generating page image via img2img: %v\n", err)
						fmt.Printf("    Falling back to text-to-image for page %d\n", pageIndex+1)
						mu.Unlock()
						imageData, err = aigc.GenerateImageByText(context.Background(), fullPrompt)
					}
				} else {
					imageData, err = aigc.GenerateImageByText(context.Background(), fullPrompt)
				}
				if err != nil {
					mu.Lock()
					fmt.Printf("    Error generating image for page %d: %v\n", pageIndex+1, err)
					mu.Unlock()
				} else {
					imageFile := filepath.Join(chapterDir, fmt.Sprintf("page_%03d.png", pageIndex+1))
					if err := os.WriteFile(imageFile, imageData, 0644); err != nil {
						mu.Lock()
						fmt.Printf("    Error saving image for page %d: %v\n", pageIndex+1, err)
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
							fmt.Printf("  [Page %d/%d] Panel %d audio %d/%d...\n",
								pageIndex+1, total, panelIndex+1, audioIndex+1, totalSegments)
							mu.Unlock()

							audioData, err := aigc.TextToSpeechSimple(
								context.Background(),
								audioSegment.Text,
								audioSegment.VoiceType,
								audioSegment.SpeedRatio,
							)
							if err != nil {
								mu.Lock()
								fmt.Printf("    Error generating audio for page %d panel %d segment %d: %v\n",
									pageIndex+1, panelIndex+1, audioIndex+1, err)
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
								fmt.Printf("    Error saving audio for page %d panel %d segment %d: %v\n",
									pageIndex+1, panelIndex+1, audioIndex+1, err)
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
			}(j, page, totalPages, summary.CharacterFeatures)
		}

		wg.Wait()

		fmt.Printf("Chapter %d completed!\n", i+1)
	}

	globalManifest := ChapterCharacterManifest{}
	for _, key := range characterOrder {
		feature, ok := characterRegistry[key]
		if !ok {
			continue
		}
		asset := characterAssets[key]
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

	if len(globalManifest.Characters) > 0 {
		manifestBytes, err := json.MarshalIndent(globalManifest, "", "  ")
		if err != nil {
			fmt.Printf("Error marshalling global character manifest: %v\n", err)
		} else {
			manifestPath := filepath.Join(globalCharacterDir, "manifest.json")
			if err := os.WriteFile(manifestPath, manifestBytes, 0644); err != nil {
				fmt.Printf("Error writing global character manifest: %v\n", err)
			}
		}
	}

	fmt.Printf("\nTracked %d unique characters across chapters\n", len(characterOrder))

	fmt.Printf("\n=== Comic Generation Complete ===\n")
	fmt.Printf("Processed %d chapters\n", processCount)
	fmt.Printf("Output saved to: %s\n", *outputDir)
}

func sanitizeCharacterFileStem(name string, index int) string {
	base := strings.TrimSpace(name)
	if base == "" {
		return fmt.Sprintf("character_%02d", index+1)
	}
	var builder strings.Builder
	for _, r := range base {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		} else {
			builder.WriteRune('_')
		}
	}
	cleaned := strings.Trim(builder.String(), "_")
	if cleaned == "" {
		return fmt.Sprintf("character_%02d", index+1)
	}
	if index >= 0 {
		return fmt.Sprintf("%s_%02d", cleaned, index+1)
	}
	return cleaned
}
