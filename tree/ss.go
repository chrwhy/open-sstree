package sstree

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chrwhy/open-sstree/dict"
)

const DEFAULT_FOREST = "default"

// SSTree is the main search suggestion tree engine.
// It manages multiple forests loaded from .dict files and provides thread-safe search.
type SSTree struct {
	forests map[string]*Forest
	mu      sync.RWMutex
	dictDir string
}

// New creates a new SSTree instance and loads all .dict files from the given directory.
func New(dictDir string) (*SSTree, error) {
	s := &SSTree{
		forests: make(map[string]*Forest),
		dictDir: dictDir,
	}
	if err := s.Load(); err != nil {
		return nil, fmt.Errorf("initial load: %w", err)
	}
	return s, nil
}

// Load scans the dict directory for .dict files and builds forests.
// This is used for initial loading.
func (s *SSTree) Load() error {
	newForests := make(map[string]*Forest)
	entries, err := os.ReadDir(s.dictDir)
	if err != nil {
		return fmt.Errorf("read dict directory %s: %w", s.dictDir, err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".dict") {
			continue
		}
		if entry.Name() == "pinyin.dict" || entry.Name() == "cn_pinyin.dict" {
			continue
		}

		filePath := s.dictDir + "/" + entry.Name()
		t0 := time.Now()
		slog.Info("Processing dict file", "file", entry.Name())
		lines, err := dict.LoadSentences(filePath)
		if err != nil {
			return fmt.Errorf("load dict %s: %w", entry.Name(), err)
		}
		tempForest := BuildForest(lines)
		cateName := strings.Replace(entry.Name(), ".dict", "", -1)
		newForests[cateName] = tempForest
		slog.Info("Dict loaded", "file", entry.Name(), "trees", len(tempForest.Trees), "cost", time.Since(t0))
	}

	s.mu.Lock()
	s.forests = newForests
	s.mu.Unlock()

	return nil
}

// Reload reloads all .dict files and replaces the forests atomically.
func (s *SSTree) Reload() error {
	return s.Load()
}

// Search performs a search on the forest for the given category and keyword.
func (s *SSTree) Search(cate, keyword string) []*TreeNode {
	keyword = PreProcess(keyword)
	if len(keyword) < 1 {
		return []*TreeNode{}
	}

	s.mu.RLock()
	forest := s.forests[cate]
	s.mu.RUnlock()

	if forest == nil {
		return []*TreeNode{}
	}

	return XSearch(forest, keyword)
}

// GetForest returns the forest for the given category (for testing/debugging).
func (s *SSTree) GetForest(cate string) *Forest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.forests[cate]
}
