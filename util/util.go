package util

import (
	"log"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

func Int2Str(val int) string {
	return strconv.Itoa(val)
}

func Str2Int(val string) int {
	result, _ := strconv.Atoi(val)
	return result
}

func IsEnCharacter(x rune) bool {
	return unicode.IsLetter(x) && !unicode.Is(unicode.Han, x)
}

func Tokenize(raw []rune) []string {
	if len(raw) == 0 {
		return []string{}
	}

	preCharacter := raw[0]
	preType := 0
	if !IsEnCharacter(preCharacter) {
		preType = 1
	}
	//part := make([]rune, 0)
	tokens := make([]string, 0)
	tokenStr := string(preCharacter)
	for i, character := range raw {
		//log.Println(string(character))
		if i == 0 {
			continue
		}

		currentType := 0
		if !IsEnCharacter(character) {
			currentType = 1
		}

		if currentType != preType {
			log.Println("token :", tokenStr)
			tokens = append(tokens, tokenStr)
			tokenStr = string(character)
			preType = currentType
		} else {
			tokenStr += string(character)
		}
	}

	tokens = append(tokens, tokenStr)
	log.Println("tokens:", tokens)

	return tokens
}

func RunesToStrings(inputs []rune) []string {
	input := make([]string, 0)
	for _, inputRune := range inputs {
		input = append(input, string(inputRune))
	}

	return input
}

func Sort(origin []string) {
	sort.Slice(origin, func(i, j int) bool {
		return len(origin[i]) < len(origin[j])
	})
}

func Concat(input []string, separator string) string {
	prefixStr := ""
	for _, pre := range input {
		if prefixStr == "" {
			prefixStr += pre
		} else {
			prefixStr += separator + pre
		}
	}

	return prefixStr
}

func Reverse(arr *[]string) {
	var temp string
	length := len(*arr)
	for i := 0; i < length/2; i++ {
		temp = (*arr)[i]
		(*arr)[i] = (*arr)[length-1-i]
		(*arr)[length-1-i] = temp
	}
}

func TrimMark(paragraph string) string {
	paragraph = strings.Replace(paragraph, "，", "", -1)
	paragraph = strings.Replace(paragraph, "。", "", -1)
	paragraph = strings.Replace(paragraph, "！", "", -1)
	paragraph = strings.Replace(paragraph, "？", "", -1)
	return paragraph
}
