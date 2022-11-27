package sstree

import (
	"log"
	"os"
	"sstree/dict"
	"strings"
	"time"
)

var MyForests map[string]*Forest

// const DICT = "sample.dict"
const DEFAULT_FOREST = "default"

func init() {
	MultiLoad()
}

func MultiLoad() {
	MyForests = make(map[string]*Forest)
	entries, _ := os.ReadDir("./")
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".dict") {
			if entry.Name() == "pinyin.dict" || entry.Name() == "cn_pinyin.dict" {
				continue
			}
			t0 := time.Now()
			log.Println("Processing ", entry.Name())
			lines := dict.LoadSentences(entry.Name())
			tempForest := BuildForest(lines)
			MyForests[strings.Replace(entry.Name(), ".dict", "", -1)] = tempForest
			log.Printf("There are %d trees in my cn forest\n", len(tempForest.Trees))
			log.Println("Build trees cost: ", time.Since(t0))
		}
	}
}

func Search(cate, keyword string) []string {
	keyword = PreProcess(keyword)
	if len(keyword) < 1 {
		return []string{}
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	//log.SetOutput(io.Discard)
	finalResult := make([]string, 0)
	//finalResult = V1Search(keyword)
	finalResult = XSearch(MyForests[cate], keyword)
	return finalResult
}
