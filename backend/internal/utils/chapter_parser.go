package utils

import (
	"regexp"
	"strings"
)

type Chapter struct {
	Index   int
	Title   string
	Content string
}

func ParseNovelChapters(content string) []Chapter {
	chapterPattern := regexp.MustCompile(`(?m)^(第[一二三四五六七八九十百千0-9]+[章节回]|Chapter\s*\d+|[0-9]+\.)\s*(.+?)$`)
	
	matches := chapterPattern.FindAllStringIndex(content, -1)
	if len(matches) == 0 {
		return []Chapter{{
			Index:   1,
			Title:   "第一章",
			Content: strings.TrimSpace(content),
		}}
	}
	
	var chapters []Chapter
	for i, match := range matches {
		startIdx := match[0]
		var endIdx int
		if i < len(matches)-1 {
			endIdx = matches[i+1][0]
		} else {
			endIdx = len(content)
		}
		
		chapterText := content[startIdx:endIdx]
		lines := strings.SplitN(chapterText, "\n", 2)
		title := strings.TrimSpace(lines[0])
		
		var chapterContent string
		if len(lines) > 1 {
			chapterContent = strings.TrimSpace(lines[1])
		}
		
		if chapterContent != "" {
			chapters = append(chapters, Chapter{
				Index:   i + 1,
				Title:   title,
				Content: chapterContent,
			})
		}
	}
	
	return chapters
}
