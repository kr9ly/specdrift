package internal

import (
	"path/filepath"
	"testing"
)

func TestIsIgnored(t *testing.T) {
	patterns := []string{"*_test.go", "vendor/*"}

	tests := []struct {
		path    string
		ignored bool
	}{
		{"main.go", false},
		{"checker_test.go", true},
		{"internal/checker_test.go", true}, // base name match
		{"vendor/lib.go", true},
		{"internal/checker.go", false},
	}

	for _, tt := range tests {
		got := IsIgnored(tt.path, patterns)
		if got != tt.ignored {
			t.Errorf("IsIgnored(%q) = %v, want %v", tt.path, got, tt.ignored)
		}
	}
}

func TestFilterIgnored(t *testing.T) {
	patterns := []string{"*_test.go"}
	input := []string{"main.go", "main_test.go", "checker.go", "checker_test.go"}

	result := FilterIgnored(input, patterns)

	if len(result) != 2 {
		t.Fatalf("expected 2, got %d: %v", len(result), result)
	}
	if result[0] != "main.go" || result[1] != "checker.go" {
		t.Errorf("expected [main.go, checker.go], got %v", result)
	}
}

func TestFilterIgnored_NilPatterns(t *testing.T) {
	input := []string{"a.go", "b.go"}
	result := FilterIgnored(input, nil)

	if len(result) != 2 {
		t.Errorf("nil patterns should return all, got %v", result)
	}
}

func TestLoadIgnorePatterns(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, ".specdriftignore"), `# Comment
*_test.go
vendor/*

# Another comment
generated/*.go
`)

	patterns, err := LoadIgnorePatterns(dir)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"*_test.go", "vendor/*", "generated/*.go"}
	if len(patterns) != len(expected) {
		t.Fatalf("expected %d patterns, got %d: %v", len(expected), len(patterns), patterns)
	}
	for i, p := range patterns {
		if p != expected[i] {
			t.Errorf("pattern %d: expected %q, got %q", i, expected[i], p)
		}
	}
}

func TestLoadIgnorePatterns_NoFile(t *testing.T) {
	dir := t.TempDir()
	patterns, err := LoadIgnorePatterns(dir)
	if err != nil {
		t.Fatal(err)
	}
	if patterns != nil {
		t.Errorf("expected nil, got %v", patterns)
	}
}
