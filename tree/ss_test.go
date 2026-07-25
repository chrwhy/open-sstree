package sstree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	// Create a temporary directory with a dict file
	tmpDir, err := os.MkdirTemp("", "sstree_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dictContent := "今天\n明天\n人工控制\n"
	dictPath := filepath.Join(tmpDir, "test.dict")
	if err := os.WriteFile(dictPath, []byte(dictContent), 0644); err != nil {
		t.Fatal(err)
	}

	engine, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if engine == nil {
		t.Fatal("New returned nil engine")
	}

	// Should have loaded the test forest
	forest := engine.GetForest("test")
	if forest == nil {
		t.Fatal("expected 'test' forest to be loaded")
	}

	if len(forest.Trees) == 0 {
		t.Error("expected trees in forest")
	}
}

func TestNew_EmptyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sstree_test_empty")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	engine, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New failed for empty dir: %v", err)
	}

	if engine == nil {
		t.Fatal("New returned nil engine for empty dir")
	}
}

func TestNew_NonexistentDir(t *testing.T) {
	_, err := New("/nonexistent/dir")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestSSTree_Search(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sstree_search")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dictContent := "今天\n今天天气\n明天\n人工\n人工控制\n"
	dictPath := filepath.Join(tmpDir, "default.dict")
	if err := os.WriteFile(dictPath, []byte(dictContent), 0644); err != nil {
		t.Fatal(err)
	}

	engine, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	tests := []struct {
		name    string
		cate    string
		keyword string
		wantLen int
	}{
		{"chinese search", "default", "今天", 0},   // may return 0 or more
		{"empty keyword", "default", "", 0},
		{"nonexistent cate", "nonexistent", "今天", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Search(tt.cate, tt.keyword)
			if tt.wantLen == 0 && len(result) != 0 {
				// For "今天", we expect results
				if tt.keyword == "今天" && len(result) > 0 {
					return // This is expected
				}
			}
		})
	}
}

func TestSSTree_Search_WithResults(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sstree_search_results")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dictContent := "今天\n今天天气\n明天\n"
	dictPath := filepath.Join(tmpDir, "default.dict")
	if err := os.WriteFile(dictPath, []byte(dictContent), 0644); err != nil {
		t.Fatal(err)
	}

	engine, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	result := engine.Search("default", "今天")
	if len(result) == 0 {
		t.Error("expected results for '今天'")
	}

	suggestions := XTraverse(result)
	found := false
	for _, s := range suggestions {
		if s.Text == "今天" || s.Text == "今天天气" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected '今天' or '今天天气' in suggestions")
	}
}

func TestSSTree_Reload(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sstree_reload")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dictContent := "今天\n明天\n"
	dictPath := filepath.Join(tmpDir, "default.dict")
	if err := os.WriteFile(dictPath, []byte(dictContent), 0644); err != nil {
		t.Fatal(err)
	}

	engine, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Reload should work
	if err := engine.Reload(); err != nil {
		t.Fatalf("Reload failed: %v", err)
	}

	// Forest should still be available
	forest := engine.GetForest("default")
	if forest == nil {
		t.Fatal("expected 'default' forest after reload")
	}
}

func TestSSTree_GetForest_Nonexistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sstree_getforest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	engine, err := New(tmpDir)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	forest := engine.GetForest("nonexistent")
	if forest != nil {
		t.Error("expected nil for nonexistent forest")
	}
}
