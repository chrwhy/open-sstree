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

// PrintSuggestions prints unique suggestions up to maxResults.
// Returns the number of unique suggestions printed.
func PrintSuggestions(suggestions []string) int {
	checker := make(map[string]bool)
	count := 0
	for _, suggestion := range suggestions {
		if !checker[suggestion] {
			if count >= 20 {
				break
			}
			checker[suggestion] = true
			count++
			slog.Info("建议", "suggestion", suggestion)
		}
	}
	return count
}
