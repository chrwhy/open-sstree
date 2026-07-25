package sstree

import (
	"log/slog"
	"strings"

	"github.com/chrwhy/open-pinyin/parser"
)

func ParsePinyin(text string) [][]string {
	return parser.Parse(text)
}

func PreProcess(input string) string {
	return strings.ToLower(strings.Replace(strings.TrimSpace(input), "'", "", -1))
}

// PrintSuggestions prints suggestions with scores (already sorted by score descending).
// Returns the number of suggestions printed.
func PrintSuggestions(suggestions []Suggestion) int {
	count := 0
	for _, s := range suggestions {
		if count >= 20 {
			break
		}
		count++
		slog.Info("建议", "suggestion", s.Text, "score", s.Score)
	}
	return count
}
