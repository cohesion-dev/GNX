package utils

import (
	"regexp"
	"strings"
)

type Section struct {
	Index   int
	Title   string
	Content string
}

func ParseNovelSections(content string) []Section {
	chapterPattern := regexp.MustCompile(`(?m)^(第[一二三四五六七八九十百千0-9]+[章节回]|Chapter\s*\d+|[0-9]+\.)\s*(.+?)$`)
	
	matches := chapterPattern.FindAllStringIndex(content, -1)
	if len(matches) == 0 {
		return []Section{{
			Index:   1,
			Title:   "第一章",
			Content: strings.TrimSpace(content),
		}}
	}
	
	var sections []Section
	for i, match := range matches {
		startIdx := match[0]
		var endIdx int
		if i < len(matches)-1 {
			endIdx = matches[i+1][0]
		} else {
			endIdx = len(content)
		}
		
		sectionText := content[startIdx:endIdx]
		lines := strings.SplitN(sectionText, "\n", 2)
		title := strings.TrimSpace(lines[0])
		
		var sectionContent string
		if len(lines) > 1 {
			sectionContent = strings.TrimSpace(lines[1])
		}
		
		if sectionContent != "" {
			sections = append(sections, Section{
				Index:   i + 1,
				Title:   title,
				Content: sectionContent,
			})
		}
	}
	
	return sections
}
