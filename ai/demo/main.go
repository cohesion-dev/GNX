package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"

	"qiniu-ai-image-generator/gnxaigc"
)

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

	var characterFeatures []gnxaigc.CharacterFeature

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

		summaryChapter := func() (*gnxaigc.SummaryChapterOutput, error) {
			for i := 0; i < 3; i++ {
				summary, err := aigc.SummaryChapter(context.Background(), gnxaigc.SummaryChapterInput{
					NovelTitle:           *novelTitle,
					ChapterTitle:         chapter.Title,
					Content:              chapter.Content,
					AvailableVoiceStyles: availableVoices,
					CharacterFeatures:    characterFeatures,
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
		summary, err := summaryChapter()
		if err != nil {
			fmt.Printf("Error generating storyboard for chapter %d: %v\n", i+1, err)
			continue
		}

		characterFeatures = summary.CharacterFeatures

		characterDir := filepath.Join(chapterDir, "characters")
		if err := os.MkdirAll(characterDir, 0755); err != nil {
			fmt.Printf("Error creating character directory: %v\n", err)
		} else {
			fmt.Printf("Generating %d character concept images...\n", len(summary.CharacterFeatures))
			for idx, feature := range summary.CharacterFeatures {
				prompt := strings.TrimSpace(feature.ConceptArtPrompt)
				if prompt == "" {
					fmt.Printf("  [Character %d] Missing concept_art_prompt for %s, skip.\n", idx+1, feature.Basic.Name)
					continue
				}

				trimmedStyle := strings.TrimSpace(*imageStyle)
				fullPrompt := prompt
				if trimmedStyle != "" {
					fullPrompt = fmt.Sprintf("%s %s", trimmedStyle, prompt)
				}

				fmt.Printf("  [Character %d] Generating concept art for %s\n", idx+1, feature.Basic.Name)
				imageData, err := aigc.GenerateImageByText(context.Background(), fullPrompt)
				if err != nil {
					fmt.Printf("    Error generating concept art: %v\n", err)
					continue
				}

				fileStem := sanitizeCharacterFileStem(feature.Basic.Name, idx)
				imageFile := filepath.Join(characterDir, fmt.Sprintf("%s.png", fileStem))
				if err := os.WriteFile(imageFile, imageData, 0644); err != nil {
					fmt.Printf("    Error saving concept art: %v\n", err)
				} else {
					fmt.Printf("    Saved concept art to %s\n", imageFile)
				}

				promptFile := filepath.Join(characterDir, fmt.Sprintf("%s_prompt.txt", fileStem))
				if err := os.WriteFile(promptFile, []byte(fullPrompt+"\n"), 0644); err != nil {
					fmt.Printf("    Error saving concept prompt: %v\n", err)
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
			go func(pageIndex int, pageItem gnxaigc.StoryboardPage, total int) {
				defer wg.Done()

				mu.Lock()
				fmt.Printf("  [Page %d/%d] Generating image...\n", pageIndex+1, total)
				mu.Unlock()

				fullPrompt := gnxaigc.ComposePageImagePrompt(*imageStyle, pageItem)
				imageData, err := aigc.GenerateImageByText(context.Background(), fullPrompt)
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
			}(j, page, totalPages)
		}

		wg.Wait()

		fmt.Printf("Chapter %d completed!\n", i+1)
	}

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
	return cleaned
}
