package sstree

import (
	"container/heap"
	"log/slog"
	"strings"
	"time"
	"unicode"

	"github.com/chrwhy/open-pinyin/parser"

	"github.com/chrwhy/open-sstree/util"
)

func XCnPinyinSearch(forest *Forest, root *TreeNode, input []rune) []*PinyinSearchResult {
	result := make([]*PinyinSearchResult, 0)
	slog.Debug("XCnPinyinSearch", "root", root, "input", string(input))
	leftover := input
	root, leftover = XCnSearch(forest, root, input)

	if len(leftover) == 0 {
		result = append(result, &PinyinSearchResult{root, nil})
		return result
	}

	if root == nil {
		pinyinGroups := ParsePinyin(string(leftover))
		for _, pinyinGroup := range pinyinGroups {
			candidates := make([]*TreeNode, 0)
			if len(pinyinGroup) == 1 {
				candidates = GetPinyinPrefixRootNodeFromForest(forest, pinyinGroup[0])
				for _, candidate := range candidates {
					result = append(result, &PinyinSearchResult{candidate, nil})
				}
			} else {
				candidates = GetPinyinRootNodeFromForest(forest, pinyinGroup[0])
				for _, candidate := range candidates {
					temp := XPinyinSearch(forest, candidate, "", pinyinGroup[1:])
					result = append(result, temp...)
				}
			}
		}
	} else {
		searchResult := XPinyinSearch(forest, root, string(leftover), nil)
		result = append(result, searchResult...)
	}

	return result
}

func XPinyinSearch(forest *Forest, root *TreeNode, leftover string, parsedPinyinGroup []string) []*PinyinSearchResult {
	result := make([]*PinyinSearchResult, 0)
	slog.Debug("XPinyinSearch", "stopNode", root.Data, "leftover", leftover)

	if len(leftover) < 1 && len(parsedPinyinGroup) == 0 {
		result = append(result, &PinyinSearchResult{root, nil})
		return result
	}

	pinyinGroups := make([][]string, 0)
	if len(leftover) == 0 && len(parsedPinyinGroup) > 0 {
		pinyinGroups = [][]string{parsedPinyinGroup}
	} else {
		pinyinGroups = ParsePinyin(leftover)
	}

	slog.Debug("XPinyinSearch", "pinyinGroups", pinyinGroups)
	checker := make(map[string]bool)
	for _, pinyinGroup := range pinyinGroups {
		if checker[pinyinGroup[0]] {
			continue
		}
		slog.Debug("XPinyinSearch", "pinyinGroup", pinyinGroup)
		v3PinyinSearchCandidates := PinyinSearch(root, pinyinGroup, false)
		slog.Debug("PinyinSearch", "resultLen", len(v3PinyinSearchCandidates))
		if len(v3PinyinSearchCandidates) < 1 {
			checker[pinyinGroup[0]] = true
		}

		for _, v3PinyinSearchCandidate := range v3PinyinSearchCandidates {
			if len(v3PinyinSearchCandidate.Leftover) == len(pinyinGroup) {
				checker[pinyinGroup[0]] = true
			}
			if len(v3PinyinSearchCandidate.Leftover) == 0 {
				result = append(result, v3PinyinSearchCandidate)
			} else {
				slog.Debug("pinyin searched candidate", "data", v3PinyinSearchCandidate.Node.Data)
				firstChar := v3PinyinSearchCandidate.Leftover[0][0:1]
				if idx, ok := v3PinyinSearchCandidate.Node.CnSlot[firstChar]; ok {
					node := v3PinyinSearchCandidate.Node.LeafNodes[idx]
					tempResult := XCnPinyinSearch(forest, node, []rune(util.Concat(v3PinyinSearchCandidate.Leftover, "")[1:]))
					slog.Debug("XCnPinyinSearch recursive", "data", node.Data, "leftover", util.Concat(v3PinyinSearchCandidate.Leftover, ""))
					result = append(result, tempResult...)
				}
			}
		}
	}

	return result
}

func XSearch(forest *Forest, input string) []*TreeNode {
	tokens := util.Tokenize([]rune(strings.ToLower(input)))
	if len(tokens) == 0 {
		return nil
	}
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

	return candidates
}

func internalXSearch(forest *Forest, root *TreeNode, input []rune) []*TreeNode {
	if len(input) == 0 {
		return nil
	}
	tokens := util.Tokenize(input)
	if unicode.Is(unicode.Han, input[0]) {
		root, leftover := XCnSearch(forest, root, []rune(tokens[0]))
		slog.Debug("XCnSearch", "leftover", string(leftover))
		if len(leftover) == 0 {
			return []*TreeNode{root}
		} else {
			return nil
		}
	} else {
		slog.Debug("internalXSearch", "input", string(input))
		internalXSearchResult := XCnPinyinSearch(forest, root, input)

		finalResult := make([]*TreeNode, 0)
		finalResultChecker := make(map[*TreeNode]bool)
		if len(internalXSearchResult) > 0 {
			for _, internalXSearchCandidate := range internalXSearchResult {
				slog.Debug("internalXSearch stop at", "data", internalXSearchCandidate.Node.Data)
				if !finalResultChecker[internalXSearchCandidate.Node] {
					finalResult = append(finalResult, internalXSearchCandidate.Node)
					finalResultChecker[internalXSearchCandidate.Node] = true
				}
			}
		}

		pinyinGroups := ParsePinyin(string(input))
		if len(pinyinGroups) > 0 {
			slog.Debug("Going to try pure pinyin search", "input", string(input))
			tempCache := make(map[string][]*TreeNode)
			for _, pinyinGroup := range pinyinGroups {
				candidates, ok := tempCache[pinyinGroup[0]]
				if !ok {
					if root == nil {
						candidates = GetPinyinPrefixRootNodeFromForest(forest, pinyinGroup[0])
					} else {
						candidates = GetPinyinPrefixNodeFromNode(root, pinyinGroup[0])
					}
					tempCache[pinyinGroup[0]] = candidates
				}
				for _, candidate := range candidates {
					purePinyinSearchCandidates := XPinyinSearch(forest, candidate, "", pinyinGroup[1:])
					for _, purePinyinSearchCandidate := range purePinyinSearchCandidates {
						if !finalResultChecker[purePinyinSearchCandidate.Node] {
							finalResult = append(finalResult, purePinyinSearchCandidate.Node)
							finalResultChecker[purePinyinSearchCandidate.Node] = true
						}
					}
				}
			}
		}

		initials := parser.ParseInitial(string(input))
		if len(initials) > 0 {
			slog.Debug("Going to try initial pinyin search", "input", string(input))
			leftInitials := initials
			initialRoots := make([]*TreeNode, 0)
			if root == nil {
				initialRoots = GetPinyinPrefixRootNodeFromForest(forest, initials[0])
				if len(initialRoots) == 0 {
					return finalResult
				}
				leftInitials = initials[1:]
			} else {
				initialRoots = append(initialRoots, root)
			}

			for _, initialCandidate := range initialRoots {
				initialPinyinSearchResult := PinyinSearch(initialCandidate, leftInitials, true)
				for _, initialPinyinSearchCandidate := range initialPinyinSearchResult {
					if len(initialPinyinSearchCandidate.Leftover) == 0 {
						if !finalResultChecker[initialPinyinSearchCandidate.Node] {
							finalResult = append(finalResult, initialPinyinSearchCandidate.Node)
							finalResultChecker[initialPinyinSearchCandidate.Node] = true
						}
						slog.Debug("Initial search stop", "data", initialPinyinSearchCandidate.Node.Data, "leftover", initialPinyinSearchCandidate.Leftover)
					}
				}
			}
		}

		return finalResult
	}
}

func XTraverse(candidates []*TreeNode) []Suggestion {
	h := &suggestionHeap{}
	heap.Init(h)
	candidateChecker := make(map[*TreeNode]bool)
	t0 := time.Now()
	slog.Info("XTraverse", "candidateLen", len(candidates))

	for _, candidate := range candidates {
		if candidateChecker[candidate] {
			continue
		}
		candidateChecker[candidate] = true

		var parentPath string
		if candidate.Parent != nil {
			parentPath = util.Concat(ReverseTraverse(candidate), "")
		} else {
			parentPath = candidate.Data
		}
		TraverseTopK(candidate, parentPath, h)
	}

	// Extract from heap in descending score order
	results := make([]Suggestion, h.Len())
	for i := h.Len() - 1; i >= 0; i-- {
		results[i] = heap.Pop(h).(Suggestion)
	}

	slog.Info("XTraverse", "cost", time.Since(t0), "topK", len(results))
	return results
}

func XCnSearch(forest *Forest, root *TreeNode, input []rune) (*TreeNode, []rune) {
	slog.Debug("XCnSearch", "root", root, "input", string(input))
	leftover := input

	if forest == nil {
		return nil, input
	}

	if root == nil {
		root = GetRootNodeFromForest(forest, string(input[0:1]))
		if root != nil {
			slog.Debug("XCnSearch found root", "root", root, "leftover", string(leftover))
			leftover = input[1:]
		}
	}

	if root != nil {
		root, leftover = CnSearch(root, leftover)
		slog.Debug("CnSearch", "root", root, "leftover", string(leftover))
	}

	return root, leftover
}
