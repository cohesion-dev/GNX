package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"

	_ "image/gif"
	_ "image/jpeg"
)

// mergeImagesSideBySide composes multiple images horizontally into a single PNG frame.
func mergeImagesSideBySide(imageData [][]byte) ([]byte, error) {
	if len(imageData) == 0 {
		return nil, errors.New("mergeImagesSideBySide: no image data provided")
	}

	type decoded struct {
		img    image.Image
		bounds image.Rectangle
	}

	decodedImages := make([]decoded, 0, len(imageData))
	totalWidth := 0
	maxHeight := 0

	for idx, data := range imageData {
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("mergeImagesSideBySide: decoding reference image %d: %w", idx+1, err)
		}

		bounds := img.Bounds()
		decodedImages = append(decodedImages, decoded{img: img, bounds: bounds})
		totalWidth += bounds.Dx()
		if h := bounds.Dy(); h > maxHeight {
			maxHeight = h
		}
	}

	canvas := image.NewRGBA(image.Rect(0, 0, totalWidth, maxHeight))
	offsetX := 0
	for _, item := range decodedImages {
		draw.Draw(
			canvas,
			image.Rect(offsetX, 0, offsetX+item.bounds.Dx(), item.bounds.Dy()),
			item.img,
			item.bounds.Min,
			draw.Src,
		)
		offsetX += item.bounds.Dx()
	}

	buf := &bytes.Buffer{}
	if err := png.Encode(buf, canvas); err != nil {
		return nil, fmt.Errorf("mergeImagesSideBySide: encoding merged image: %w", err)
	}

	return buf.Bytes(), nil
}
