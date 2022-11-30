package sstree

import (
	pydict "github.com/chrwhy/open-pinyin/dict"
	"log"
	"sstree/dict"
	"sstree/util"
	"strings"
	"time"
)

type TreeNode struct {
	Data              string
	PinyinData        []string
	LeaveNodes        []*TreeNode
	Parent            *TreeNode
	CnSlot            map[string]string
	PinyinSlot        map[string][]string
	PinyinInitialSlot map[string][]string
	IsBlack           bool
	Score             int
}

type Forest struct {
	Trees             []*TreeNode
	PinyinSlot        map[string][]string
	PinyinInitialSlot map[string][]string
	CnSlot            map[string]string
}

func BuildForest(input []dict.Sentence) *Forest {
	t0 := time.Now()
	myForest := Forest{}
	myForest.PinyinSlot = make(map[string][]string)
	myForest.CnSlot = make(map[string]string)
	myForest.PinyinInitialSlot = make(map[string][]string)

	for _, word := range input {
		if len(word.Words) < 1 {
			continue
		}
		treeRootName := word.Words[0]
		if treeRootName == "" {
			continue
		}

		var myTreeRoot *TreeNode
		if myForest.CnSlot[treeRootName] != "" {
			myTreeRoot = myForest.Trees[util.Str2Int(myForest.CnSlot[treeRootName])]
		}

		if myTreeRoot != nil {
			if word.Score > myTreeRoot.Score {
				myTreeRoot.Score = word.Score
			}
			internalBuildTree(myTreeRoot, word.Words[1:], word.Score)
		} else {
			myTreeRoot = &TreeNode{}
			myTreeRoot.CnSlot = make(map[string]string)
			myTreeRoot.PinyinInitialSlot = make(map[string][]string)
			myTreeRoot.PinyinSlot = make(map[string][]string)
			myTreeRoot.Data = treeRootName
			myTreeRoot.IsBlack = false
			myTreeRoot.PinyinData = pydict.GetCnPinyin(treeRootName)
			myForest.CnSlot[treeRootName] = util.Int2Str(len(myForest.Trees))
			for _, pinyin := range myTreeRoot.PinyinData {
				myForest.PinyinSlot[pinyin] = append(myForest.PinyinSlot[pinyin], util.Int2Str(len(myForest.Trees)))
				myForest.PinyinInitialSlot[pinyin[0:1]] = append(myForest.PinyinInitialSlot[pinyin[0:1]], util.Int2Str(len(myForest.Trees)))
			}
			myForest.Trees = append(myForest.Trees, myTreeRoot)
			myTreeRoot.Score = word.Score
			internalBuildTree(myTreeRoot, word.Words[1:], word.Score)
		}
	}
	log.Println("Build forest cost: ", time.Now().Sub(t0))
	return &myForest
}

func internalBuildTree(current *TreeNode, input []string, score int) {
	if len(input) < 1 {
		current.Score = score
		current.IsBlack = true
		return
	}

	if len(current.LeaveNodes) < 1 {
		current.LeaveNodes = make([]*TreeNode, 0)
		newNode := &TreeNode{Data: input[0]}
		newNode.PinyinData = pydict.GetCnPinyin(newNode.Data)
		newNode.Parent = current
		newNode.IsBlack = false
		newNode.CnSlot = make(map[string]string)
		newNode.PinyinSlot = make(map[string][]string)
		newNode.PinyinInitialSlot = make(map[string][]string)
		newNode.Score = score
		current.CnSlot[input[0]] = "0"

		for _, pinyin := range newNode.PinyinData {
			current.PinyinSlot[pinyin] = []string{"0"}
			current.PinyinInitialSlot[pinyin[0:1]] = append(current.PinyinInitialSlot[pinyin[0:1]], "0")
		}
		current.LeaveNodes = append(current.LeaveNodes, newNode)
		internalBuildTree(newNode, input[1:], score)
	} else {
		var found *TreeNode
		if _, ok := current.CnSlot[input[0]]; ok {
			found = current.LeaveNodes[util.Str2Int(current.CnSlot[input[0]])]
		}

		if found == nil {
			newNode := &TreeNode{Data: input[0]}
			newNode.PinyinData = pydict.GetCnPinyin(newNode.Data)
			newNode.Parent = current
			newNode.IsBlack = false
			newNode.PinyinSlot = make(map[string][]string)
			newNode.PinyinInitialSlot = make(map[string][]string)
			newNode.CnSlot = make(map[string]string)
			newNode.Score = score

			current.CnSlot[input[0]] = util.Int2Str(len(current.LeaveNodes))
			for _, pinyin := range newNode.PinyinData {
				current.PinyinSlot[pinyin] = append(current.PinyinSlot[pinyin], util.Int2Str(len(current.LeaveNodes)))
				current.PinyinInitialSlot[pinyin[0:1]] = append(current.PinyinInitialSlot[pinyin[0:1]], util.Int2Str(len(current.LeaveNodes)))
			}
			current.LeaveNodes = append(current.LeaveNodes, newNode)
			internalBuildTree(newNode, input[1:], score)
		} else {
			if score > found.Score {
				found.Score = score
			}
			internalBuildTree(found, input[1:], score)
		}
	}
}

func CnSearch(node *TreeNode, input []string) *TreeNode {
	if node == nil {
		return nil
	}

	if len(input) == 0 {
		return node
	}

	head := input[0]
	if slot, ok := node.CnSlot[head]; ok {
		leave := node.LeaveNodes[util.Str2Int(slot)]
		return CnSearch(leave, input[1:])
	}

	return node
}

func CnSearchV2(node *TreeNode, input []rune) (*TreeNode, []rune) {
	if node == nil {
		return nil, input
	}

	if len(input) == 0 {
		return node, input
	}

	head := string(input[0:1])
	if slot, ok := node.CnSlot[head]; ok {
		leave := node.LeaveNodes[util.Str2Int(slot)]
		return CnSearchV2(leave, input[1:])
	}

	return node, input
}

func GetRootNodeFromForest(farm *Forest, input string) *TreeNode {
	if len(farm.CnSlot[input]) > 0 {
		return farm.Trees[util.Str2Int(farm.CnSlot[input])]
	}
	return nil
}

func GetPinyinRootNodeFromForest(farm *Forest, firstPinyin string) []*TreeNode {
	foundNodes := make([]*TreeNode, 0)
	slots := farm.PinyinSlot[firstPinyin]

	for _, slot := range slots {
		slotNode := farm.Trees[util.Str2Int(slot)]
		for _, pinyin := range slotNode.PinyinData {
			if pinyin == firstPinyin {
				foundNodes = append(foundNodes, slotNode)
				break
			}
		}
	}
	return foundNodes
}

func GetPinyinPrefixRootNodeFromForest(farm *Forest, firstPinyin string) []*TreeNode {
	log.Println("GetPinyinPrefixRootNodeFromForest first pinyin: ", firstPinyin)
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
	initial := firstPinyin[0:1]
	slots := farm.PinyinInitialSlot[initial]
	for _, slot := range slots {
		slotNode := farm.Trees[util.Str2Int(slot)]
		for _, pinyin := range slotNode.PinyinData {
			if pinyin[0:1] == initial {
				foundNodes = append(foundNodes, slotNode)
				break
			}
		}
	}

	return foundNodes
}

func Traverse(node *TreeNode, prefix string) []string {
	result := make([]string, 0)
	if node.IsBlack && len(node.LeaveNodes) > 0 {
		result = append(result, prefix)
	}

	if len(node.LeaveNodes) < 1 {
		result = append(result, prefix)
	}

	for _, leave := range node.LeaveNodes {
		result = append(result, Traverse(leave, prefix+leave.Data)...)
	}

	return result
}

func ReverseTraverse(node *TreeNode) []string {
	if node == nil {
		return []string{}
	}
	result := make([]string, 0)
	for {
		if node.Parent != nil {
			//log.Println(node.Data)
			result = append(result, node.Data)
			node = node.Parent
		} else {
			//log.Println(node.Data)
			result = append(result, node.Data)
			break
		}
	}

	util.Reverse(&result)
	return result
}

func PinyinSearch(found *TreeNode, input []string, initial bool) []*TreeNode {
	result := make([]*TreeNode, 0)
	if len(input) < 1 {
		return []*TreeNode{found}
	}

	if len(found.LeaveNodes) < 1 {
		log.Println("no leave nodes")
		return []*TreeNode{found}
	} else {
		head := input[0]
		var slots []string
		if initial || len(input) == 1 {
			slots = found.PinyinInitialSlot[head[0:1]]
		} else {
			slots = found.PinyinSlot[head]
		}
		checker := make(map[string]string)
		for _, slot := range slots {
			if _, ok := checker[slot]; ok {
				//multiple pinyin case
				continue
			}
			checker[slot] = slot
			slotNode := found.LeaveNodes[util.Str2Int(slot)]
			for _, pinyin := range slotNode.PinyinData {
				compareTo := head
				if initial {
					compareTo = compareTo[0:1]
					pinyin = pinyin[0:1]
				}
				if len(input) == 1 {
					if strings.HasPrefix(pinyin, compareTo) {
						result = append(result, PinyinSearch(slotNode, input[1:], initial)...)
						break
					}
				} else {
					if pinyin == compareTo {
						result = append(result, PinyinSearch(slotNode, input[1:], initial)...)
						break
					}
				}
			}
		}
		if len(result) < 1 {
			return []*TreeNode{found}
		}
		return result
	}
}

type PinyinSearchV3Result struct {
	Node     *TreeNode
	Leftover []string
}

func PinyinSearchV2(found *TreeNode, input []string, initial bool) []*PinyinSearchV3Result {
	result := make([]*PinyinSearchV3Result, 0)
	if len(input) < 1 {
		return []*PinyinSearchV3Result{}
	}
	if len(found.LeaveNodes) < 1 && len(input) < 1 {
		log.Println("no leave nodes")
		return []*PinyinSearchV3Result{{found, []string{}}}
	} else {
		head := input[0]
		var slots []string
		if initial || len(input) == 1 {
			slots = found.PinyinInitialSlot[head[0:1]]
		} else {
			slots = found.PinyinSlot[head]
		}
		checker := make(map[string]string)
		for _, slot := range slots {
			if _, ok := checker[slot]; ok {
				//multiple pinyin case
				continue
			}
			checker[slot] = slot
			slotNode := found.LeaveNodes[util.Str2Int(slot)]
			for _, pinyin := range slotNode.PinyinData {
				compareTo := head
				if initial {
					compareTo = compareTo[0:1]
					pinyin = pinyin[0:1]
				}

				if len(input) == 1 {
					if strings.HasPrefix(pinyin, compareTo) {
						result = append(result, &PinyinSearchV3Result{slotNode, []string{}})
						break
					}
				} else {
					if pinyin == compareTo {
						temp := PinyinSearchV2(slotNode, input[1:], initial)
						result = append(result, temp...)
						break
					}
				}
			}
		}

		if len(result) < 1 {
			return []*PinyinSearchV3Result{{found, input}}
		} else {
			return result
		}
	}
}
