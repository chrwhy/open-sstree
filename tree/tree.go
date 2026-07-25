package sstree

import (
	"container/heap"
	"log/slog"
	"strings"
	"time"

	pydict "github.com/chrwhy/open-pinyin/dict"

	"github.com/chrwhy/open-sstree/dict"
)

type TreeNode struct {
	Data              string
	PinyinData        []string
	LeafNodes         []*TreeNode
	Parent            *TreeNode
	CnSlot            map[string]int
	PinyinSlot        map[string][]int
	PinyinInitialSlot map[string][]int
	IsBlack           bool
	Score             int
}

type Forest struct {
	Trees             []*TreeNode
	PinyinSlot        map[string][]int
	PinyinInitialSlot map[string][]int
	CnSlot            map[string]int
}

// Suggestion represents a search suggestion with its score.
type Suggestion struct {
	Text  string
	Score int
}

// suggestionHeap implements a min-heap of Suggestion, ordered by Score.
type suggestionHeap []Suggestion

func (h suggestionHeap) Len() int            { return len(h) }
func (h suggestionHeap) Less(i, j int) bool  { return h[i].Score < h[j].Score }
func (h suggestionHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *suggestionHeap) Push(x interface{}) { *h = append(*h, x.(Suggestion)) }
func (h *suggestionHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

const MaxSuggestions = 20

func BuildForest(input []dict.Sentence) *Forest {
	t0 := time.Now()
	myForest := Forest{}
	myForest.PinyinSlot = make(map[string][]int)
	myForest.CnSlot = make(map[string]int)
	myForest.PinyinInitialSlot = make(map[string][]int)

	for _, word := range input {
		if len(word.Words) < 1 {
			continue
		}
		treeRootName := word.Words[0]
		if treeRootName == "" {
			continue
		}

		var myTreeRoot *TreeNode
		if idx, ok := myForest.CnSlot[treeRootName]; ok {
			myTreeRoot = myForest.Trees[idx]
		}

		if myTreeRoot != nil {
			if word.Score > myTreeRoot.Score {
				myTreeRoot.Score = word.Score
			}
			internalBuildTree(myTreeRoot, word.Words[1:], word.Score)
		} else {
			myTreeRoot = &TreeNode{}
			myTreeRoot.CnSlot = make(map[string]int)
			myTreeRoot.PinyinInitialSlot = make(map[string][]int)
			myTreeRoot.PinyinSlot = make(map[string][]int)
			myTreeRoot.Data = treeRootName
			myTreeRoot.IsBlack = false
			myTreeRoot.PinyinData = pydict.GetCnPinyin(treeRootName)
			treeIdx := len(myForest.Trees)
			myForest.CnSlot[treeRootName] = treeIdx
			for _, pinyin := range myTreeRoot.PinyinData {
				if len(pinyin) == 0 {
					continue
				}
				myForest.PinyinSlot[pinyin] = append(myForest.PinyinSlot[pinyin], treeIdx)
				myForest.PinyinInitialSlot[pinyin[0:1]] = append(myForest.PinyinInitialSlot[pinyin[0:1]], treeIdx)
			}
			myForest.Trees = append(myForest.Trees, myTreeRoot)
			myTreeRoot.Score = word.Score
			internalBuildTree(myTreeRoot, word.Words[1:], word.Score)
		}
	}
	slog.Info("Build forest", "cost", time.Since(t0), "trees", len(myForest.Trees))
	return &myForest
}

func internalBuildTree(current *TreeNode, input []string, score int) {
	if len(input) < 1 {
		current.Score = score
		current.IsBlack = true
		return
	}

	if len(current.LeafNodes) < 1 {
		current.LeafNodes = make([]*TreeNode, 0)
		newNode := &TreeNode{Data: input[0]}
		newNode.PinyinData = pydict.GetCnPinyin(newNode.Data)
		newNode.Parent = current
		newNode.IsBlack = false
		newNode.CnSlot = make(map[string]int)
		newNode.PinyinSlot = make(map[string][]int)
		newNode.PinyinInitialSlot = make(map[string][]int)
		newNode.Score = score
		current.CnSlot[input[0]] = 0

		for _, pinyin := range newNode.PinyinData {
			if len(pinyin) == 0 {
				continue
			}
			current.PinyinSlot[pinyin] = []int{0}
			current.PinyinInitialSlot[pinyin[0:1]] = append(current.PinyinInitialSlot[pinyin[0:1]], 0)
		}
		current.LeafNodes = append(current.LeafNodes, newNode)
		internalBuildTree(newNode, input[1:], score)
	} else {
		var found *TreeNode
		if idx, ok := current.CnSlot[input[0]]; ok {
			found = current.LeafNodes[idx]
		}

		if found == nil {
			newNode := &TreeNode{Data: input[0]}
			newNode.PinyinData = pydict.GetCnPinyin(newNode.Data)
			newNode.Parent = current
			newNode.IsBlack = false
			newNode.PinyinSlot = make(map[string][]int)
			newNode.PinyinInitialSlot = make(map[string][]int)
			newNode.CnSlot = make(map[string]int)
			newNode.Score = score

			childIdx := len(current.LeafNodes)
			current.CnSlot[input[0]] = childIdx
			for _, pinyin := range newNode.PinyinData {
				if len(pinyin) == 0 {
					continue
				}
				current.PinyinSlot[pinyin] = append(current.PinyinSlot[pinyin], childIdx)
				current.PinyinInitialSlot[pinyin[0:1]] = append(current.PinyinInitialSlot[pinyin[0:1]], childIdx)
			}
			current.LeafNodes = append(current.LeafNodes, newNode)
			internalBuildTree(newNode, input[1:], score)
		} else {
			if score > found.Score {
				found.Score = score
			}
			internalBuildTree(found, input[1:], score)
		}
	}
}

func CnSearch(node *TreeNode, input []rune) (*TreeNode, []rune) {
	if node == nil {
		return nil, input
	}

	if len(input) == 0 {
		return node, input
	}

	head := string(input[0:1])
	if idx, ok := node.CnSlot[head]; ok {
		leaf := node.LeafNodes[idx]
		return CnSearch(leaf, input[1:])
	}

	return node, input
}

func GetRootNodeFromForest(farm *Forest, input string) *TreeNode {
	if farm == nil {
		return nil
	}
	if idx, ok := farm.CnSlot[input]; ok {
		return farm.Trees[idx]
	}
	return nil
}

func GetPinyinRootNodeFromForest(farm *Forest, firstPinyin string) []*TreeNode {
	foundNodes := make([]*TreeNode, 0)
	slots := farm.PinyinSlot[firstPinyin]

	for _, slot := range slots {
		slotNode := farm.Trees[slot]
		for _, pinyin := range slotNode.PinyinData {
			if pinyin == firstPinyin {
				foundNodes = append(foundNodes, slotNode)
				break
			}
		}
	}
	return foundNodes
}

func GetPinyinPrefixNodeFromNode(node *TreeNode, firstPinyin string) []*TreeNode {
	slog.Debug("GetPinyinPrefixNodeFromNode", "firstPinyin", firstPinyin)
	candidates := node.LeafNodes
	result := make([]*TreeNode, 0)
	for _, candidate := range candidates {
		for _, pinyin := range candidate.PinyinData {
			if strings.HasPrefix(pinyin, firstPinyin) {
				result = append(result, candidate)
			}
		}
	}

	return result
}

func GetPinyinPrefixRootNodeFromForest(farm *Forest, firstPinyin string) []*TreeNode {
	slog.Debug("GetPinyinPrefixRootNodeFromForest", "firstPinyin", firstPinyin)
	candidates := GetPinyinInitialRootNodeFromForest(farm, firstPinyin)
	result := make([]*TreeNode, 0)
	for _, candidate := range candidates {
		for _, pinyin := range candidate.PinyinData {
			if strings.HasPrefix(pinyin, firstPinyin) {
				result = append(result, candidate)
			}
		}
	}

	return result
}

func GetPinyinInitialRootNodeFromForest(farm *Forest, firstPinyin string) []*TreeNode {
	foundNodes := make([]*TreeNode, 0)
	if len(firstPinyin) == 0 {
		return foundNodes
	}
	initial := firstPinyin[0:1]
	slots := farm.PinyinInitialSlot[initial]
	for _, slot := range slots {
		slotNode := farm.Trees[slot]
		for _, pinyin := range slotNode.PinyinData {
			if len(pinyin) > 0 && pinyin[0:1] == initial {
				foundNodes = append(foundNodes, slotNode)
				break
			}
		}
	}

	return foundNodes
}

func Traverse(node *TreeNode, prefix string) []string {
	result := make([]string, 0)
	if node.IsBlack && len(node.LeafNodes) > 0 {
		result = append(result, prefix)
	}

	if len(node.LeafNodes) < 1 {
		result = append(result, prefix)
	}

	for _, leaf := range node.LeafNodes {
		result = append(result, Traverse(leaf, prefix+leaf.Data)...)
	}

	return result
}

// TraverseTopK traverses the tree and collects top-K suggestions using a min-heap with pruning.
// node.Score is the maximum score in the subtree (guaranteed by internalBuildTree's max logic).
// When the heap is full and node.Score <= heap minimum, the entire subtree is skipped.
func TraverseTopK(node *TreeNode, prefix string, h *suggestionHeap) {
	// Pruning: heap is full and subtree max score <= heap min → skip entire subtree
	if h.Len() >= MaxSuggestions && node.Score <= (*h)[0].Score {
		return
	}

	if node.IsBlack {
		if h.Len() < MaxSuggestions {
			heap.Push(h, Suggestion{Text: prefix, Score: node.Score})
		} else if node.Score > (*h)[0].Score {
			heap.Pop(h)
			heap.Push(h, Suggestion{Text: prefix, Score: node.Score})
		}
	}

	for _, leaf := range node.LeafNodes {
		TraverseTopK(leaf, prefix+leaf.Data, h)
	}
}

func ReverseTraverse(node *TreeNode) []string {
	if node == nil {
		return []string{}
	}
	result := make([]string, 0)
	for {
		if node.Parent != nil {
			result = append(result, node.Data)
			node = node.Parent
		} else {
			result = append(result, node.Data)
			break
		}
	}

	ReverseStrings(&result)
	return result
}

// ReverseStrings reverses a string slice in place.
func ReverseStrings(arr *[]string) {
	var temp string
	length := len(*arr)
	for i := 0; i < length/2; i++ {
		temp = (*arr)[i]
		(*arr)[i] = (*arr)[length-1-i]
		(*arr)[length-1-i] = temp
	}
}

type PinyinSearchResult struct {
	Node     *TreeNode
	Leftover []string
}

func PinyinSearch(found *TreeNode, input []string, initial bool) []*PinyinSearchResult {
	result := make([]*PinyinSearchResult, 0)
	if len(input) < 1 {
		return []*PinyinSearchResult{}
	}
	if len(found.LeafNodes) < 1 && len(input) < 1 {
		return []*PinyinSearchResult{{found, []string{}}}
	} else {
		head := input[0]
		var slots []int
		if initial || len(input) == 1 {
			if len(head) == 0 {
				return result
			}
			slots = found.PinyinInitialSlot[head[0:1]]
		} else {
			slots = found.PinyinSlot[head]
		}
		checker := make(map[int]bool)
		for _, slot := range slots {
			if checker[slot] {
				//multiple pinyin case
				continue
			}
			checker[slot] = true
			slotNode := found.LeafNodes[slot]
			for _, pinyin := range slotNode.PinyinData {
				compareTo := head
				if initial {
					if len(compareTo) == 0 || len(pinyin) == 0 {
						continue
					}
					compareTo = compareTo[0:1]
					pinyin = pinyin[0:1]
				}

				if len(input) == 1 {
					if strings.HasPrefix(pinyin, compareTo) {
						result = append(result, &PinyinSearchResult{slotNode, []string{}})
						break
					}
				} else {
					if pinyin == compareTo {
						temp := PinyinSearch(slotNode, input[1:], initial)
						result = append(result, temp...)
						break
					}
				}
			}
		}

		if len(result) < 1 {
			return []*PinyinSearchResult{{found, input}}
		} else {
			return result
		}
	}
}
