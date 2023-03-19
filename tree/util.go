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
	count := 0
	for _, suggestion := range suggestions {
		if _, ok := checker[suggestion]; !ok {
			if count >= 20 {
				//break
			}
			count = count + 1
			log.Println("建议:", suggestion)
			//checker[suggestion] = "1"
		}
	}
}
