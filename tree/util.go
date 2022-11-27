package sstree

import (
	"github.com/chrwhy/open-pinyin/parser"
	"log"
	"strings"
)

func ParsePinyin(text string) [][]string {
	return parser.Parse(text)
	//return [][]string{parser.GreedyParse(text)}
}

func PreProcess(input string) string {
	return strings.Replace(strings.TrimSpace(input), "'", "", -1)
}

func PrintSuggestions(suggestions []string) {
	checker := make(map[string]string)
	for _, suggestion := range suggestions {
		if _, ok := checker[suggestion]; !ok {
			log.Println("建议:", suggestion)
			//checker[suggestion] = "1"
		}
	}
}
