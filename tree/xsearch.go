package sstree

import (
	"github.com/chrwhy/open-pinyin/parser"
	"log"
	"sstree/util"
	"unicode"
)

func XCnPinyinSearch(forest *Forest, root *TreeNode, input []rune) []*PinyinSearchV3Result {
	result := make([]*PinyinSearchV3Result, 0)
	log.Println("internalXSearch root: ", root)
	log.Println("internalXSearch input: ", string(input))
	leftover := input
	root, leftover = XCnSearch(forest, root, input)

	if len(leftover) == 0 {
		result = append(result, &PinyinSearchV3Result{root, nil})
		return result
	}

	if root == nil {
		pinyinGroups := ParsePinyin(string(leftover))
		for _, pinyinGroup := range pinyinGroups {
			candidates := make([]*TreeNode, 0)
			if len(pinyinGroup) == 1 {
				candidates = GetPinyinPrefixRootNodeFromForest(forest, pinyinGroup[0])
				for _, candidate := range candidates {
					result = append(result, &PinyinSearchV3Result{candidate, nil})
				}
			} else {
				candidates = GetPinyinRootNodeFromForest(forest, pinyinGroup[0])
				for _, candidate := range candidates {
					temp := XPinyinSearchV2(forest, candidate, "", pinyinGroup[1:])
					result = append(result, temp...)
				}
			}
		}
	} else {
		searchResult := XPinyinSearchV2(forest, root, string(leftover), nil)
		result = append(result, searchResult...)
	}

	return result
}

func XPinyinSearchV2(forest *Forest, root *TreeNode, leftover string, parsedPinyinGroup []string) []*PinyinSearchV3Result {
	result := make([]*PinyinSearchV3Result, 0)
	log.Println("stop node: ", root.Data)
	log.Println("leftover: ", leftover)

	if len(leftover) < 1 && len(parsedPinyinGroup) == 0 {
		result = append(result, &PinyinSearchV3Result{root, nil})
		return result
	}

	pinyinGroups := make([][]string, 0)
	if len(leftover) == 0 && len(parsedPinyinGroup) > 0 {
		pinyinGroups = [][]string{parsedPinyinGroup}
	} else {
		pinyinGroups = ParsePinyin(leftover)
	}

	log.Println("pinyin groups:", pinyinGroups)
	checker := make(map[string]string)
	for _, pinyinGroup := range pinyinGroups {
		if _, ok := checker[pinyinGroup[0]]; ok {
			//continue
		}
		log.Println("===============")
		log.Println("pinyin group:", pinyinGroup)
		v3PinyinSearchCandidates := PinyinSearchV2(root, pinyinGroup, false)
		log.Println("PinyinSearchV2 result len: ", len(v3PinyinSearchCandidates))
		if len(v3PinyinSearchCandidates) < 1 {
			checker[pinyinGroup[0]] = pinyinGroup[0]
		}

		for _, v3PinyinSearchCandidate := range v3PinyinSearchCandidates {
			if len(v3PinyinSearchCandidate.Leftover) == len(pinyinGroup) {
				checker[pinyinGroup[0]] = pinyinGroup[0]
			}
			if len(v3PinyinSearchCandidate.Leftover) == 0 {
				result = append(result, v3PinyinSearchCandidate)
			} else {
				log.Println("pinyin searched candidates: ", v3PinyinSearchCandidate.Node.Data)
				firstChar := v3PinyinSearchCandidate.Leftover[0][0:1]
				slot := v3PinyinSearchCandidate.Node.CnSlot[firstChar]
				if len(slot) > 0 {
					node := v3PinyinSearchCandidate.Node.LeaveNodes[util.Str2Int(slot)]
					tempResult := XCnPinyinSearch(forest, node, []rune(util.Concat(v3PinyinSearchCandidate.Leftover, "")[1:]))
					log.Println(node.Data + "=====" + util.Concat(v3PinyinSearchCandidate.Leftover, ""))
					result = append(result, tempResult...)
				}
			}
		}
	}

	return result
}

func XSearch(forest *Forest, input string) []string {
	finalResult := make([]string, 0)
	tokens := util.Tokenize([]rune(input))
	candidates := internalXSearch(forest, nil, []rune(tokens[0]))
	if len(candidates) == 0 {
		return nil
	}

	for _, token := range tokens[1:] {
		tempCandidates := make([]*TreeNode, 0)
		for _, candidate := range candidates {
			temp := internalXSearch(forest, candidate, []rune(token))
			if len(temp) > 0 {
				tempCandidates = append(tempCandidates, temp...)
			}
		}

		if len(tempCandidates) == 0 {
			candidates = nil
			break
		} else {
			candidates = tempCandidates
		}
	}
	duplicateChecker := make(map[*TreeNode]*TreeNode)
	if len(candidates) > 0 {
		for _, candidate := range candidates {
			if _, ok := duplicateChecker[candidate]; ok {
				continue
			}
			duplicateChecker[candidate] = candidate
			suggestions := Traverse(candidate, util.Concat(ReverseTraverse(candidate), ""))
			finalResult = append(finalResult, suggestions...)
		}
	}

	return finalResult
}

func internalXSearch(forest *Forest, root *TreeNode, input []rune) []*TreeNode {
	tokens := util.Tokenize(input)
	if unicode.Is(unicode.Han, input[0]) {
		root, leftover := XCnSearch(forest, root, []rune(tokens[0]))
		log.Println("XCnSearch leftover: ", leftover)
		if len(leftover) == 0 {
			return []*TreeNode{root}
		} else {
			return nil
		}
	} else {
		log.Println("internalXSearch:", string(input))
		internalXSearchResult := XCnPinyinSearch(forest, root, input)

		finalResult := make([]*TreeNode, 0)
		if len(internalXSearchResult) > 0 {
			for _, internalXSearchCandidate := range internalXSearchResult {
				log.Println("internalXSearch stop at: ", internalXSearchCandidate.Node.Data)
				finalResult = append(finalResult, internalXSearchCandidate.Node)
			}
			//return finalResult
		}

		pinyinGroups := ParsePinyin(string(input))
		if len(pinyinGroups) > 0 {
			log.Println("Going to try pure pinyin search: ", input)
			tempCache := make(map[string][]*TreeNode)
			for _, pinyinGroup := range pinyinGroups {
				candidates, ok := tempCache[pinyinGroup[0]]
				if !ok {
					candidates = GetPinyinPrefixRootNodeFromForest(forest, pinyinGroup[0])
					tempCache[pinyinGroup[0]] = candidates
				}
				for _, candidate := range candidates {
					purePinyinSearchCandidates := XPinyinSearchV2(forest, candidate, "", pinyinGroup[1:])
					for _, purePinyinSearchCandidate := range purePinyinSearchCandidates {
						finalResult = append(finalResult, purePinyinSearchCandidate.Node)
					}
				}
			}
		}

		initials := parser.ParseInitial(string(input))
		if len(initials) > 0 {
			log.Println("Going to try initial pinyin search: ", input)
			leftInitials := initials
			initialRoots := make([]*TreeNode, 0)
			if root == nil {
				initialRoots = GetPinyinPrefixRootNodeFromForest(forest, initials[0])
				if len(initialRoots) == 0 {
					return nil
				}
				leftInitials = initials[1:]
			} else {
				initialRoots = append(initialRoots, root)
			}

			for _, initialCandidate := range initialRoots {
				initialPinyinSearchResult := PinyinSearchV2(initialCandidate, leftInitials, true)
				for _, initialPinyinSearchCandidate := range initialPinyinSearchResult {
					if len(initialPinyinSearchCandidate.Leftover) == 0 {
						finalResult = append(finalResult, initialPinyinSearchCandidate.Node)
						log.Println("Initial search stop node: ", initialPinyinSearchCandidate.Node.Data)
						log.Println("Initial search leftover: ", initialPinyinSearchCandidate.Leftover)
					}
				}
			}
		}

		return finalResult
	}
}

func XCnSearch(forest *Forest, root *TreeNode, input []rune) (*TreeNode, []rune) {
	log.Println("internalXSearch start: ")
	log.Println("root: ", root)
	log.Println("input: ", string(input))
	leftover := input

	if root == nil {
		root = GetRootNodeFromForest(forest, string(input[0:1]))
		if root != nil {
			log.Println("root : ", root)
			log.Println("leftover: ", leftover)
			leftover = input[1:]
		}
	}

	if root != nil {
		root, leftover = CnSearchV2(root, leftover)
		log.Println("CnSearchV2 root : ", root)
		log.Println("CnSearchV2 leftover: ", string(leftover))
	}

	return root, leftover
}
