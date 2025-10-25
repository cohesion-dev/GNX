package main

import (
	"context"
	"flag"
	"fmt"
	"os"

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

	generator := NewComicGenerator(
		context.Background(),
		ComicGeneratorConfig{
			NovelTitle: *novelTitle,
			OutputDir:  *outputDir,
			ImageStyle: *imageStyle,
		},
		gnxaigc.NewGnxAIGC(gnxaigc.Config{}),
	)

	if err := generator.Run(*inputFile, *maxChapters); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
