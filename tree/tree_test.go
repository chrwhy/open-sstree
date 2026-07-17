package sstree

import (
	"testing"

	"github.com/chrwhy/open-sstree/dict"
)

func makeSentence(words string, score int) dict.Sentence {
	characters := make([]string, 0, len([]rune(words)))
	for _, r := range words {
		characters = append(characters, string(r))
	}
	return dict.Sentence{Words: characters, Score: score}
}

func TestBuildForest(t *testing.T) {
	input := []dict.Sentence{
		makeSentence("今天", 10),
		makeSentence("今天天气", 20),
		makeSentence("明天", 5),
		makeSentence("人工", 3),
		makeSentence("人工控制", 8),
	}

	forest := BuildForest(input)

	if forest == nil {
		t.Fatal("BuildForest returned nil")
	}

	// Should have 3 root trees: 今, 明, 人
	if len(forest.Trees) != 3 {
		t.Errorf("expected 3 trees, got %d", len(forest.Trees))
	}

	// Check CnSlot
	if _, ok := forest.CnSlot["今"]; !ok {
		t.Error("expected '今' in forest.CnSlot")
	}
	if _, ok := forest.CnSlot["明"]; !ok {
		t.Error("expected '明' in forest.CnSlot")
	}
	if _, ok := forest.CnSlot["人"]; !ok {
		t.Error("expected '人' in forest.CnSlot")
	}
}

func TestBuildForest_EmptyInput(t *testing.T) {
	forest := BuildForest([]dict.Sentence{})
	if forest == nil {
		t.Fatal("BuildForest returned nil for empty input")
	}
	if len(forest.Trees) != 0 {
		t.Errorf("expected 0 trees, got %d", len(forest.Trees))
	}
}

func TestBuildForest_ScorePropagation(t *testing.T) {
	input := []dict.Sentence{
		makeSentence("今天", 10),
		makeSentence("今天天气", 20),
	}

	forest := BuildForest(input)
	root := forest.Trees[0]

	// Root should have max score of children
	if root.Score != 20 {
		t.Errorf("expected root score 20, got %d", root.Score)
	}
}

func TestCnSearch(t *testing.T) {
	input := []dict.Sentence{
		makeSentence("今天", 10),
		makeSentence("今天天气", 20),
		makeSentence("明天", 5),
	}
	forest := BuildForest(input)

	// Search for "今"
	root := GetRootNodeFromForest(forest, "今")
	if root == nil {
		t.Fatal("GetRootNodeFromForest('今') returned nil")
	}

	// Search for "天" in the tree rooted at "今"
	node, leftover := CnSearch(root, []rune("天"))
	if node == nil {
		t.Fatal("CnSearch returned nil node")
	}
	if node.Data != "天" {
		t.Errorf("expected node data '天', got %q", node.Data)
	}
	if len(leftover) != 0 {
		t.Errorf("expected empty leftover, got %q", string(leftover))
	}

	// Search for "天气" in the tree rooted at "今"
	node, leftover = CnSearch(root, []rune("天"))
	if node == nil {
		t.Fatal("CnSearch returned nil node")
	}
	if len(leftover) != 0 {
		t.Errorf("expected empty leftover, got %q", string(leftover))
	}
}

func TestCnSearch_NotFound(t *testing.T) {
	input := []dict.Sentence{
		makeSentence("今天", 10),
	}
	forest := BuildForest(input)

	root := GetRootNodeFromForest(forest, "今")
	if root == nil {
		t.Fatal("GetRootNodeFromForest('今') returned nil")
	}

	// Search for "明" which doesn't exist under "今"
	node, leftover := CnSearch(root, []rune("明"))
	if node != root {
		t.Errorf("expected node to be root when not found")
	}
	if string(leftover) != "明" {
		t.Errorf("expected leftover '明', got %q", string(leftover))
	}
}

func TestCnSearch_NilNode(t *testing.T) {
	node, leftover := CnSearch(nil, []rune("test"))
	if node != nil {
		t.Error("expected nil node for nil input")
	}
	if string(leftover) != "test" {
		t.Errorf("expected leftover 'test', got %q", string(leftover))
	}
}

func TestTraverse(t *testing.T) {
	input := []dict.Sentence{
		makeSentence("今天", 10),
		makeSentence("今天天气", 20),
	}
	forest := BuildForest(input)

	root := forest.Trees[0]
	result := Traverse(root, root.Data)

	if len(result) == 0 {
		t.Fatal("Traverse returned empty result")
	}

	found := make(map[string]bool)
	for _, r := range result {
		found[r] = true
	}

	if !found["今天"] {
		t.Error("expected '今天' in traverse results")
	}
	if !found["今天天气"] {
		t.Error("expected '今天天气' in traverse results")
	}
}

func TestReverseTraverse(t *testing.T) {
	input := []dict.Sentence{
		makeSentence("今天天气", 10),
	}
	forest := BuildForest(input)

	// Navigate to the leaf node "气"
	root := forest.Trees[0] // "今"
	leaf1 := root.LeafNodes[root.CnSlot["天"]] // "天"
	leaf2 := leaf1.LeafNodes[leaf1.CnSlot["天"]] // "天" (second)
	leaf3 := leaf2.LeafNodes[leaf2.CnSlot["气"]] // "气"

	result := ReverseTraverse(leaf3)
	expected := []string{"今", "天", "天", "气"}

	if len(result) != len(expected) {
		t.Fatalf("ReverseTraverse length = %d, want %d", len(result), len(expected))
	}
	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("ReverseTraverse[%d] = %q, want %q", i, result[i], expected[i])
		}
	}
}

func TestReverseTraverse_Nil(t *testing.T) {
	result := ReverseTraverse(nil)
	if len(result) != 0 {
		t.Errorf("expected empty result for nil, got %v", result)
	}
}

func TestGetRootNodeFromForest(t *testing.T) {
	input := []dict.Sentence{
		makeSentence("今天", 10),
		makeSentence("明天", 5),
	}
	forest := BuildForest(input)

	node := GetRootNodeFromForest(forest, "今")
	if node == nil {
		t.Fatal("GetRootNodeFromForest('今') returned nil")
	}
	if node.Data != "今" {
		t.Errorf("expected data '今', got %q", node.Data)
	}

	node = GetRootNodeFromForest(forest, "不")
	if node != nil {
		t.Error("expected nil for nonexistent root")
	}
}

func TestPinyinSearch(t *testing.T) {
	input := []dict.Sentence{
		makeSentence("今天", 10),
		makeSentence("明天", 5),
	}
	forest := BuildForest(input)

	root := forest.Trees[0] // "今"
	result := PinyinSearch(root, []string{"tian"}, false)
	if len(result) == 0 {
		t.Fatal("PinyinSearch returned empty result")
	}

	if result[0].Node.Data != "天" {
		t.Errorf("expected node data '天', got %q", result[0].Node.Data)
	}
}

func TestPinyinSearch_Initial(t *testing.T) {
	input := []dict.Sentence{
		makeSentence("今天", 10),
		makeSentence("明天", 5),
	}
	forest := BuildForest(input)

	root := forest.Trees[0] // "今"
	result := PinyinSearch(root, []string{"t"}, true)
	if len(result) == 0 {
		t.Fatal("PinyinSearch(initial) returned empty result")
	}
}

func TestPinyinSearch_EmptyInput(t *testing.T) {
	input := []dict.Sentence{
		makeSentence("今天", 10),
	}
	forest := BuildForest(input)

	root := forest.Trees[0]
	result := PinyinSearch(root, []string{}, false)
	if len(result) != 0 {
		t.Errorf("expected empty result for empty input, got %d", len(result))
	}
}

func TestForestSlots(t *testing.T) {
	input := []dict.Sentence{
		makeSentence("今天", 10),
	}
	forest := BuildForest(input)

	// CnSlot should contain "今"
	if _, ok := forest.CnSlot["今"]; !ok {
		t.Error("expected '今' in CnSlot")
	}

	// PinyinSlot should have entries for "今" pinyin
	if len(forest.PinyinSlot) == 0 {
		t.Error("expected PinyinSlot to have entries")
	}

	// PinyinInitialSlot should have entries
	if len(forest.PinyinInitialSlot) == 0 {
		t.Error("expected PinyinInitialSlot to have entries")
	}
}

func TestTreeNodeSlots(t *testing.T) {
	input := []dict.Sentence{
		makeSentence("今天", 10),
	}
	forest := BuildForest(input)

	root := forest.Trees[0] // "今"

	// CnSlot should contain "天"
	if _, ok := root.CnSlot["天"]; !ok {
		t.Error("expected '天' in root.CnSlot")
	}

	// LeafNodes should have one child
	if len(root.LeafNodes) != 1 {
		t.Errorf("expected 1 leaf node, got %d", len(root.LeafNodes))
	}
}
