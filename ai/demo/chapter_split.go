package main

import (
	"os"
	"regexp"
	"strings"
)

// NovelChapter represents a chapter parsed from a raw novel text.
type NovelChapter struct {
	Title   string
	Content string
}

var chapterHeadingPattern = regexp.MustCompile(`^第[零〇一二三四五六七八九十百千万0-9]+章.*$`)

// SplitChaptersFromFile reads the provided file and splits it into chapters.
func SplitChaptersFromFile(path string) ([]NovelChapter, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return SplitChaptersFromText(string(data)), nil
}

// SplitChaptersFromText splits a raw novel string into chapters.
func SplitChaptersFromText(raw string) []NovelChapter {
	lines := strings.Split(raw, "\n")
	var chapters []NovelChapter
	var currentTitle string
	var buffer []string

	flush := func() {
		if currentTitle == "" && len(buffer) == 0 {
			return
		}
		content := strings.Trim(strings.Join(buffer, "\n"), "\n")
		chapters = append(chapters, NovelChapter{Title: currentTitle, Content: content})
	}

	for _, line := range lines {
		trimmed := normalizeHeadingCandidate(line)
		if chapterHeadingPattern.MatchString(trimmed) {
			flush()
			currentTitle = trimmed
			buffer = buffer[:0]
			continue
		}
		if currentTitle == "" && len(strings.TrimSpace(line)) == 0 {
			continue
		}
		buffer = append(buffer, line)
	}

	flush()
	return chapters
}

// normalizeHeadingCandidate strips zero-width characters that often wrap headings in web-crawled novels.
func normalizeHeadingCandidate(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return trimmed
	}

	var b strings.Builder
	b.Grow(len(trimmed))
	for _, r := range trimmed {
		switch r {
		case '\u200B', '\u200C', '\u200D', '\uFEFF':
			continue
		default:
			b.WriteRune(r)
		}
	}

	return b.String()
}
