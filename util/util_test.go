package util

import (
	"testing"
)

func TestInt2Str(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{-1, "-1"},
		{100, "100"},
		{999999, "999999"},
	}
	for _, tt := range tests {
		result := Int2Str(tt.input)
		if result != tt.expected {
			t.Errorf("Int2Str(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestStr2Int(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		wantErr  bool
	}{
		{"0", 0, false},
		{"1", 1, false},
		{"-1", -1, false},
		{"100", 100, false},
		{"abc", 0, true},
		{"", 0, true},
		{"12.34", 0, true},
	}
	for _, tt := range tests {
		result, err := Str2Int(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("Str2Int(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && result != tt.expected {
			t.Errorf("Str2Int(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestMustStr2Int(t *testing.T) {
	// Valid cases
	if result := MustStr2Int("42"); result != 42 {
		t.Errorf("MustStr2Int(\"42\") = %d, want 42", result)
	}

	// Panic on invalid
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustStr2Int(\"abc\") should have panicked")
		}
	}()
	MustStr2Int("abc")
}

func TestIsEnCharacter(t *testing.T) {
	tests := []struct {
		input    rune
		expected bool
	}{
		{'a', true},
		{'Z', true},
		{'中', false},
		{'1', false},
		{' ', false},
		{'_', false},
	}
	for _, tt := range tests {
		result := IsEnCharacter(tt.input)
		if result != tt.expected {
			t.Errorf("IsEnCharacter(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty", "", []string{}},
		{"english only", "hello", []string{"hello"}},
		{"chinese only", "你好世界", []string{"你好世界"}},
		{"mixed", "hello你好", []string{"hello", "你好"}},
		{"chinese then english", "你好hello", []string{"你好", "hello"}},
		{"alternating", "a中b中", []string{"a", "中", "b", "中"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Tokenize([]rune(tt.input))
			if len(result) != len(tt.expected) {
				t.Errorf("Tokenize(%q) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("Tokenize(%q)[%d] = %q, want %q", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestRunesToStrings(t *testing.T) {
	result := RunesToStrings([]rune{'你', '好'})
	expected := []string{"你", "好"}
	if len(result) != len(expected) {
		t.Fatalf("RunesToStrings = %v, want %v", result, expected)
	}
	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("RunesToStrings[%d] = %q, want %q", i, result[i], expected[i])
		}
	}
}

func TestConcat(t *testing.T) {
	tests := []struct {
		input     []string
		separator string
		expected  string
	}{
		{[]string{"a", "b", "c"}, "", "abc"},
		{[]string{"a", "b", "c"}, "-", "a-b-c"},
		{[]string{"a"}, "", "a"},
		{[]string{}, "", ""},
	}
	for _, tt := range tests {
		result := Concat(tt.input, tt.separator)
		if result != tt.expected {
			t.Errorf("Concat(%v, %q) = %q, want %q", tt.input, tt.separator, result, tt.expected)
		}
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		{[]string{"a", "b", "c"}, []string{"c", "b", "a"}},
		{[]string{"a"}, []string{"a"}},
		{[]string{}, []string{}},
		{[]string{"a", "b"}, []string{"b", "a"}},
	}
	for _, tt := range tests {
		arr := make([]string, len(tt.input))
		copy(arr, tt.input)
		Reverse(&arr)
		if len(arr) != len(tt.expected) {
			t.Errorf("Reverse(%v) = %v, want %v", tt.input, arr, tt.expected)
			continue
		}
		for i := range arr {
			if arr[i] != tt.expected[i] {
				t.Errorf("Reverse(%v)[%d] = %q, want %q", tt.input, i, arr[i], tt.expected[i])
			}
		}
	}
}

func TestSort(t *testing.T) {
	arr := []string{"abc", "a", "ab"}
	Sort(arr)
	expected := []string{"a", "ab", "abc"}
	for i := range arr {
		if arr[i] != expected[i] {
			t.Errorf("Sort result[%d] = %q, want %q", i, arr[i], expected[i])
		}
	}
}

func TestTrimMark(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"你好，世界", "你好世界"},
		{"你好。", "你好"},
		{"你好！", "你好"},
		{"你好？", "你好"},
		{"hello", "hello"},
	}
	for _, tt := range tests {
		result := TrimMark(tt.input)
		if result != tt.expected {
			t.Errorf("TrimMark(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
