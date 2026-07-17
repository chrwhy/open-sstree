package dict

import (
	"os"
	"testing"
)

func TestStripLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no brackets", "hello", "hello"},
		{"with brackets", "音乐[yue]节", "音乐节"},
		{"multiple brackets", "智能[neng]电视[neng]", "智能电视"},
		{"empty brackets", "test[]test", "testtest"},
		{"brackets at start", "[abc]hello", "hello"},
		{"brackets at end", "hello[abc]", "hello"},
		{"empty string", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripLine([]rune(tt.input))
			if result != tt.expected {
				t.Errorf("StripLine(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLoadSentences(t *testing.T) {
	// Create a temporary dict file
	content := "今天\n明天\n李白醉不醒@100\n音乐[yue]节\n人工\n人工控制\n"
	tmpFile, err := os.CreateTemp("", "test*.dict")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	sentences, err := LoadSentences(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadSentences failed: %v", err)
	}

	if len(sentences) == 0 {
		t.Fatal("LoadSentences returned empty result")
	}

	// Check that sentences were loaded
	found := make(map[string]bool)
	for _, s := range sentences {
		word := ""
		for _, w := range s.Words {
			word += w
		}
		found[word] = true
	}

	expectedWords := []string{"今天", "明天", "李白醉不醒", "音乐节", "人工", "人工控制"}
	for _, w := range expectedWords {
		if !found[w] {
			t.Errorf("expected word %q not found in loaded sentences", w)
		}
	}
}

func TestLoadSentences_ScoreParsing(t *testing.T) {
	content := "李白醉不醒@100\n普通词\n"
	tmpFile, err := os.CreateTemp("", "test_score*.dict")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	sentences, err := LoadSentences(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadSentences failed: %v", err)
	}

	for _, s := range sentences {
		word := ""
		for _, w := range s.Words {
			word += w
		}
		if word == "李白醉不醒" && s.Score != 101 { // 100 + 1
			t.Errorf("expected score 101 for 李白醉不醒, got %d", s.Score)
		}
		if word == "普通词" && s.Score != 1 { // 0 + 1
			t.Errorf("expected score 1 for 普通词, got %d", s.Score)
		}
	}
}

func TestLoadSentences_FileNotFound(t *testing.T) {
	_, err := LoadSentences("/nonexistent/file.dict")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestLoadSentences_Deduplication(t *testing.T) {
	content := "今天\n今天\n今天\n"
	tmpFile, err := os.CreateTemp("", "test_dedup*.dict")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	sentences, err := LoadSentences(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadSentences failed: %v", err)
	}

	if len(sentences) != 1 {
		t.Errorf("expected 1 sentence after dedup, got %d", len(sentences))
	}
}
