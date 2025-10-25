package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"qiniu-ai-image-generator/gnxaigc"
)

func main() {
	inputFile := flag.String("input", "", "Path to novel text file")
	outputDir := flag.String("output", "output", "Output directory for generated comic assets")
	novelTitle := flag.String("title", "未知小说", "Novel title")
	maxChapters := flag.Int("max-chapters", 0, "Maximum number of chapters to process (0 for all)")
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
		summary, err := aigc.SummaryChapter(context.Background(), gnxaigc.SummaryChapterInput{
			NovelTitle:            *novelTitle,
			ChapterTitle:          chapter.Title,
			Content:               chapter.Content,
			AvailableVoiceStyles:  availableVoices,
			CharacterFeatures:     characterFeatures,
		})
		if err != nil {
			fmt.Printf("Error generating storyboard for chapter %d: %v\n", i+1, err)
			continue
		}

		characterFeatures = summary.CharacterFeatures

		storyboardJSON, _ := json.MarshalIndent(summary, "", "  ")
		storyboardFile := filepath.Join(chapterDir, "storyboard.json")
		if err := os.WriteFile(storyboardFile, storyboardJSON, 0644); err != nil {
			fmt.Printf("Error saving storyboard: %v\n", err)
		} else {
			fmt.Printf("Saved storyboard to %s\n", storyboardFile)
		}

		fmt.Printf("Processing %d storyboard items...\n", len(summary.StoryboardItems))

		for j, item := range summary.StoryboardItems {
			fmt.Printf("  [%d/%d] Generating image...\n", j+1, len(summary.StoryboardItems))
			imageData, err := aigc.GenerateImageByText(context.Background(), item.ImagePrompt)
			if err != nil {
				fmt.Printf("    Error generating image: %v\n", err)
			} else {
				imageFile := filepath.Join(chapterDir, fmt.Sprintf("scene_%03d.png", j+1))
				if err := os.WriteFile(imageFile, imageData, 0644); err != nil {
					fmt.Printf("    Error saving image: %v\n", err)
				} else {
					fmt.Printf("    Saved image to %s\n", imageFile)
				}
			}

			for k, segment := range item.SourceTextSegments {
				fmt.Printf("  [%d/%d] Generating audio segment %d/%d...\n",
					j+1, len(summary.StoryboardItems), k+1, len(item.SourceTextSegments))

				audioData, err := aigc.TextToSpeechSimple(
					context.Background(),
					segment.Text,
					segment.VoiceType,
					segment.SpeedRatio,
				)
				if err != nil {
					fmt.Printf("    Error generating audio: %v\n", err)
					continue
				}

				audioFile := filepath.Join(chapterDir, fmt.Sprintf("scene_%03d_audio_%03d.mp3", j+1, k+1))
				if err := os.WriteFile(audioFile, audioData, 0644); err != nil {
					fmt.Printf("    Error saving audio: %v\n", err)
				} else {
					fmt.Printf("    Saved audio to %s\n", audioFile)
				}
			}
		}

		fmt.Printf("Chapter %d completed!\n", i+1)
	}

	fmt.Printf("\n=== Comic Generation Complete ===\n")
	fmt.Printf("Processed %d chapters\n", processCount)
	fmt.Printf("Output saved to: %s\n", *outputDir)
}
