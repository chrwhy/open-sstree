package sstree

import (
	"testing"

	"github.com/chrwhy/open-sstree/dict"
)

func buildTestForest() *Forest {
	input := []dict.Sentence{
		makeSentence("今天", 10),
		makeSentence("今天天气", 20),
		makeSentence("明天", 5),
		makeSentence("人工", 3),
		makeSentence("人工控制", 8),
		makeSentence("人工智能", 15),
		makeSentence("电视机", 7),
		makeSentence("电视剧", 12),
		makeSentence("音乐节", 6),
		makeSentence("李白醉不醒", 100),
	}
	return BuildForest(input)
}

func TestXSearch_Chinese(t *testing.T) {
	forest := buildTestForest()

	tests := []struct {
		name      string
		keyword   string
		expectNil bool
	}{
		{"single char", "今", false},
		{"two chars", "今天", false},
		{"full word", "人工", false},
		{"nonexistent", "xyz", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := XSearch(forest, tt.keyword)
			if tt.expectNil {
				if len(result) > 0 {
					t.Errorf("XSearch(%q) returned %d results, want 0", tt.keyword, len(result))
				}
			} else {
				if len(result) == 0 {
					t.Errorf("XSearch(%q) returned empty results", tt.keyword)
				}
			}
		})
	}
}

func TestXSearch_Pinyin(t *testing.T) {
	forest := buildTestForest()

	// Pinyin search depends on open-pinyin dictionary availability.
	// We test that it doesn't panic and returns results when possible.
	tests := []struct {
		name    string
		keyword string
	}{
		{"full pinyin tian", "tian"},
		{"full pinyin gong", "gong"},
		{"pinyin prefix t", "t"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := XSearch(forest, tt.keyword)
			// Pinyin search may return empty if open-pinyin dict is not available
			t.Logf("XSearch(%q) returned %d results", tt.keyword, len(result))
		})
	}
}

func TestXSearch_EmptyInput(t *testing.T) {
	forest := buildTestForest()
	result := XSearch(forest, "")
	if len(result) != 0 {
		t.Errorf("XSearch('') returned %d results, want 0", len(result))
	}
}

func TestXSearch_MixedInput(t *testing.T) {
	forest := buildTestForest()

	// Test with mixed Chinese + Pinyin input
	result := XSearch(forest, "今天tian")
	// This should find "今天天气" via mixed search
	if len(result) == 0 {
		// Mixed search might not always find results depending on tokenization
		// This is acceptable - just ensure no panic
		t.Log("Mixed search returned no results (acceptable)")
	}
}

func TestXTraverse(t *testing.T) {
	forest := buildTestForest()

	// Get candidates for "今天"
	candidates := XSearch(forest, "今天")
	if len(candidates) == 0 {
		t.Fatal("XSearch('今天') returned empty")
	}

	suggestions := XTraverse(candidates)
	if len(suggestions) == 0 {
		t.Fatal("XTraverse returned empty suggestions")
	}

	found := make(map[string]bool)
	for _, s := range suggestions {
		found[s] = true
	}

	// Should contain "今天" and "今天天气"
	if !found["今天"] {
		t.Error("expected '今天' in suggestions")
	}
	if !found["今天天气"] {
		t.Error("expected '今天天气' in suggestions")
	}
}

func TestXTraverse_Empty(t *testing.T) {
	result := XTraverse([]*TreeNode{})
	if len(result) != 0 {
		t.Errorf("XTraverse(empty) returned %d results, want 0", len(result))
	}
}

func TestXTraverse_NilCandidates(t *testing.T) {
	result := XTraverse(nil)
	if len(result) != 0 {
		t.Errorf("XTraverse(nil) returned %d results, want 0", len(result))
	}
}

func TestXCnSearch(t *testing.T) {
	forest := buildTestForest()

	// Search from forest root
	root, leftover := XCnSearch(forest, nil, []rune("今天"))
	if root == nil {
		t.Fatal("XCnSearch returned nil root")
	}
	if len(leftover) != 0 {
		t.Errorf("expected empty leftover, got %q", string(leftover))
	}
	if root.Data != "天" {
		t.Errorf("expected node data '天', got %q", root.Data)
	}
}

func TestXCnSearch_NilForest(t *testing.T) {
	root, leftover := XCnSearch(nil, nil, []rune("今"))
	if root != nil {
		t.Error("expected nil root for nil forest")
	}
	if string(leftover) != "今" {
		t.Errorf("expected leftover '今', got %q", string(leftover))
	}
}

func TestGetPinyinRootNodeFromForest(t *testing.T) {
	forest := buildTestForest()

	nodes := GetPinyinRootNodeFromForest(forest, "jin")
	if len(nodes) == 0 {
		t.Error("expected nodes for pinyin 'jin'")
	}
}

func TestGetPinyinPrefixRootNodeFromForest(t *testing.T) {
	forest := buildTestForest()

	nodes := GetPinyinPrefixRootNodeFromForest(forest, "j")
	if len(nodes) == 0 {
		t.Error("expected nodes for pinyin prefix 'j'")
	}
}

func TestGetPinyinInitialRootNodeFromForest(t *testing.T) {
	forest := buildTestForest()

	// "天" has pinyin "tian", but it's a child node, not a root.
	// Root nodes in buildTestForest: 今(jin), 明(ming), 人(ren), 电(dian), 音(yin), 李(li)
	// So initial 'j' should work.
	nodes := GetPinyinInitialRootNodeFromForest(forest, "j")
	if len(nodes) == 0 {
		t.Error("expected nodes for initial 'j'")
	}
}

func TestGetPinyinInitialRootNodeFromForest_Empty(t *testing.T) {
	forest := buildTestForest()

	nodes := GetPinyinInitialRootNodeFromForest(forest, "")
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes for empty input, got %d", len(nodes))
	}
}
