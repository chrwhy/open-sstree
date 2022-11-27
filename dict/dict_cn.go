package dict

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sstree/util"
	"strings"
	"time"
)

const WORDS_MIN_LEN = 2
const MAX_LINE = 5000000

type Sentence struct {
	Words []string
	Score int
}

func StripLine(word []rune) string {
	strippedLine := ""
	for i := 0; i < len(word); i++ {
		if string(word[i]) == "[" {
			for j := i; j < len(word); j++ {
				if string(word[j]) == "]" {
					i = j
					break
				} else {
					continue
				}
			}
		} else {
			strippedLine = strippedLine + string(word[i])
		}
	}

	return strippedLine
}

func LoadSentences(fileName string) []Sentence {
	Sentences := make([]Sentence, 0)
	t0 := time.Now()
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	sentencesDuplicateChecker := make(map[string]string)

	counter := 0
	for scanner.Scan() {
		sentence := scanner.Text()
		if len([]rune(sentence)) > 30 || len([]rune(sentence)) < WORDS_MIN_LEN {
			continue
		}
		counter++
		if _, ok := sentencesDuplicateChecker[strings.TrimSpace(sentence)]; ok {
			continue
		}

		sentence = StripLine([]rune(sentence))
		score := 0
		if strings.Contains(sentence, "@") {
			score = util.Str2Int(strings.Split(sentence, "@")[1])
			sentence = strings.Split(sentence, "@")[0]
		}
		score++
		sentencesDuplicateChecker[sentence] = "1"

		singleCnWordArray := []rune(sentence)
		cnLineArray := make([]string, 0)
		for _, tp := range singleCnWordArray {
			cnLineArray = append(cnLineArray, string(tp))
		}

		Sentences = append(Sentences, Sentence{Words: cnLineArray, Score: score})

		if len(Sentences) > MAX_LINE {
			break
		}
	}

	fmt.Println("共加载了", counter, "组中文词条")
	fmt.Println("有效词条", len(sentencesDuplicateChecker))
	fmt.Println("其中", counter-len(sentencesDuplicateChecker), "组重复的词条")
	fmt.Println("Load cost: ", time.Since(t0))
	return Sentences
}
