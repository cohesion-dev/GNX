package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitChaptersFromFile(t *testing.T) {
	path := filepath.Join("frxxz.txt")
	chapters, err := SplitChaptersFromFile(path)
	require.NoError(t, err)
	require.NotEmpty(t, chapters)

	require.GreaterOrEqual(t, len(chapters), 3)
	require.Equal(t, "第一章 山边小村", chapters[0].Title)
	require.Equal(t, "第二章 青牛镇", chapters[1].Title)
	require.NotEmpty(t, chapters[0].Content)

	// third chapter verifies we keep scanning the file correctly.
	require.Equal(t, "第三章 七玄门", chapters[2].Title)
}

func TestSplitChaptersFromText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []NovelChapter
	}{
		{
			name:  "basic-two-chapters",
			input: "第一章 开端\n内容一\n第二章 续篇\n内容二",
			expected: []NovelChapter{
				{Title: "第一章 开端", Content: "内容一"},
				{Title: "第二章 续篇", Content: "内容二"},
			},
		},
		{
			name:  "ignores-leading-blank-lines",
			input: "\n\n第一章 标题\n内容\n\n第二章 标题二\n这是第二章的内容",
			expected: []NovelChapter{
				{Title: "第一章 标题", Content: "内容"},
				{Title: "第二章 标题二", Content: "这是第二章的内容"},
			},
		},
		{
			name:  "preface-without-heading",
			input: "这是前言\n仍旧前言\n第一章 正文\n段落一",
			expected: []NovelChapter{
				{Title: "", Content: "这是前言\n仍旧前言"},
				{Title: "第一章 正文", Content: "段落一"},
			},
		},
		{
			name:  "empty-content-between-headings",
			input: "第一章 无内容\n第二章 有内容\n这里是段落",
			expected: []NovelChapter{
				{Title: "第一章 无内容", Content: ""},
				{Title: "第二章 有内容", Content: "这里是段落"},
			},
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			chapters := SplitChaptersFromText(tc.input)
			require.Equal(t, len(tc.expected), len(chapters))

			for i := range tc.expected {
				require.Equal(t, tc.expected[i].Title, chapters[i].Title)
				require.Equal(t, tc.expected[i].Content, chapters[i].Content)
			}
		})
	}
}
