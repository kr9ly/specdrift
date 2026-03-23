package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdate_changes_detail(t *testing.T) {
	dir := t.TempDir()

	srcPath := filepath.Join(dir, "src", "main.go")
	os.MkdirAll(filepath.Dir(srcPath), 0755)
	os.WriteFile(srcPath, []byte("package main\n"), 0644)

	hash, _ := HashFile(srcPath)

	specContent := "<!-- specdrift -->\n\n<!-- source: src/main.go@00000000 -->\n\nSpec text.\n\n<!-- /source -->\n"
	specPath := filepath.Join(dir, "spec.md")
	os.WriteFile(specPath, []byte(specContent), 0644)

	result, err := Update(specPath, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Changes) != 1 {
		t.Fatalf("changes = %d, want 1", len(result.Changes))
	}
	c := result.Changes[0]
	if c.Path != "src/main.go" {
		t.Errorf("path = %q, want %q", c.Path, "src/main.go")
	}
	if c.OldHash != "00000000" {
		t.Errorf("old hash = %q, want %q", c.OldHash, "00000000")
	}
	if c.NewHash != hash {
		t.Errorf("new hash = %q, want %q", c.NewHash, hash)
	}

	// Verify file was updated
	updated, _ := os.ReadFile(specPath)
	if !strings.Contains(string(updated), hash) {
		t.Error("file should contain new hash")
	}
}

func TestUpdate_todo_resolved(t *testing.T) {
	dir := t.TempDir()

	srcPath := filepath.Join(dir, "src", "main.go")
	os.MkdirAll(filepath.Dir(srcPath), 0755)
	os.WriteFile(srcPath, []byte("package main\n"), 0644)

	hash, _ := HashFile(srcPath)

	specContent := "<!-- specdrift -->\n\n<!-- source: src/main.go@TODO -->\n\nSpec.\n\n<!-- /source -->\n"
	specPath := filepath.Join(dir, "spec.md")
	os.WriteFile(specPath, []byte(specContent), 0644)

	result, err := Update(specPath, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Changes) != 1 {
		t.Fatalf("changes = %d, want 1", len(result.Changes))
	}
	c := result.Changes[0]
	if c.OldHash != "" {
		t.Errorf("old hash for TODO should be empty, got %q", c.OldHash)
	}
	if c.NewHash != hash {
		t.Errorf("new hash = %q, want %q", c.NewHash, hash)
	}
}

func TestUpdate_no_changes(t *testing.T) {
	dir := t.TempDir()

	srcPath := filepath.Join(dir, "src", "main.go")
	os.MkdirAll(filepath.Dir(srcPath), 0755)
	os.WriteFile(srcPath, []byte("package main\n"), 0644)

	hash, _ := HashFile(srcPath)

	specContent := "<!-- specdrift -->\n\n<!-- source: src/main.go@" + hash + " -->\n\nUp to date.\n\n<!-- /source -->\n"
	specPath := filepath.Join(dir, "spec.md")
	os.WriteFile(specPath, []byte(specContent), 0644)

	result, err := Update(specPath, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Changes) != 0 {
		t.Errorf("changes = %d, want 0", len(result.Changes))
	}
}

func TestUpdate_multiple_changes(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, "src"), 0755)
	os.WriteFile(filepath.Join(dir, "src", "a.go"), []byte("package a\n"), 0644)
	os.WriteFile(filepath.Join(dir, "src", "b.go"), []byte("package b\n"), 0644)

	specContent := "<!-- specdrift -->\n\n<!-- source: src/a.go@00000000, src/b.go@00000000 -->\n\nBoth files.\n\n<!-- /source -->\n"
	specPath := filepath.Join(dir, "spec.md")
	os.WriteFile(specPath, []byte(specContent), 0644)

	result, err := Update(specPath, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Changes) != 2 {
		t.Errorf("changes = %d, want 2", len(result.Changes))
	}
}

func TestUpdate_skipped_no_declaration(t *testing.T) {
	dir := t.TempDir()

	specContent := "# Just a regular markdown file\n"
	specPath := filepath.Join(dir, "spec.md")
	os.WriteFile(specPath, []byte(specContent), 0644)

	result, err := Update(specPath, dir)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Skipped {
		t.Error("expected Skipped=true")
	}
}

func TestUpdate_missing_file_keeps_hash(t *testing.T) {
	dir := t.TempDir()

	specContent := "<!-- specdrift -->\n\n<!-- source: src/missing.go@abcdef01 -->\n\nGone.\n\n<!-- /source -->\n"
	specPath := filepath.Join(dir, "spec.md")
	os.WriteFile(specPath, []byte(specContent), 0644)

	result, err := Update(specPath, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Changes) != 0 {
		t.Errorf("changes = %d, want 0 (missing file should not produce a change)", len(result.Changes))
	}

	content, _ := os.ReadFile(specPath)
	if !strings.Contains(string(content), "abcdef01") {
		t.Error("original hash should be preserved for missing file")
	}
}
