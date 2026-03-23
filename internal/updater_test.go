package internal

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInteractiveUpdate_accept(t *testing.T) {
	dir := t.TempDir()

	// Create a source file
	srcPath := filepath.Join(dir, "src", "main.go")
	os.MkdirAll(filepath.Dir(srcPath), 0755)
	os.WriteFile(srcPath, []byte("package main\n"), 0644)

	hash, _ := HashFile(srcPath)

	// Create a spec file with a wrong hash
	specContent := "<!-- specdrift -->\n\n<!-- source: src/main.go@00000000 -->\n\nThis describes main.go.\n\n<!-- /source -->\n"
	specPath := filepath.Join(dir, "spec.md")
	os.WriteFile(specPath, []byte(specContent), 0644)

	// Simulate user typing "y"
	input := strings.NewReader("y\n")
	var output bytes.Buffer

	result, err := InteractiveUpdate(specPath, dir, input, &output)
	if err != nil {
		t.Fatal(err)
	}
	if result.Updated != 1 {
		t.Errorf("updated = %d, want 1", result.Updated)
	}

	// Verify the file was updated
	updated, _ := os.ReadFile(specPath)
	if !strings.Contains(string(updated), hash) {
		t.Errorf("expected hash %s in updated file, got:\n%s", hash, string(updated))
	}
}

func TestInteractiveUpdate_skip(t *testing.T) {
	dir := t.TempDir()

	srcPath := filepath.Join(dir, "src", "main.go")
	os.MkdirAll(filepath.Dir(srcPath), 0755)
	os.WriteFile(srcPath, []byte("package main\n"), 0644)

	specContent := "<!-- specdrift -->\n\n<!-- source: src/main.go@00000000 -->\n\nContent here.\n\n<!-- /source -->\n"
	specPath := filepath.Join(dir, "spec.md")
	os.WriteFile(specPath, []byte(specContent), 0644)

	// Simulate user typing "n"
	input := strings.NewReader("n\n")
	var output bytes.Buffer

	result, err := InteractiveUpdate(specPath, dir, input, &output)
	if err != nil {
		t.Fatal(err)
	}
	if result.Updated != 0 {
		t.Errorf("updated = %d, want 0", result.Updated)
	}

	// Verify the file was NOT updated
	content, _ := os.ReadFile(specPath)
	if strings.Contains(string(content), "package main") {
		t.Error("file should not have been updated")
	}
	if !strings.Contains(string(content), "00000000") {
		t.Error("original hash should be preserved")
	}
}

func TestInteractiveUpdate_quit(t *testing.T) {
	dir := t.TempDir()

	srcPath := filepath.Join(dir, "src", "a.go")
	os.MkdirAll(filepath.Dir(srcPath), 0755)
	os.WriteFile(srcPath, []byte("package a\n"), 0644)

	srcPath2 := filepath.Join(dir, "src", "b.go")
	os.WriteFile(srcPath2, []byte("package b\n"), 0644)

	specContent := "<!-- specdrift -->\n\n<!-- source: src/a.go@00000000 -->\n\nA docs.\n\n<!-- /source -->\n\n<!-- source: src/b.go@00000000 -->\n\nB docs.\n\n<!-- /source -->\n"
	specPath := filepath.Join(dir, "spec.md")
	os.WriteFile(specPath, []byte(specContent), 0644)

	// Simulate: accept first, quit on second
	input := strings.NewReader("y\nq\n")
	var output bytes.Buffer

	result, err := InteractiveUpdate(specPath, dir, input, &output)
	if err != nil {
		t.Fatal(err)
	}
	if result.Updated != 1 {
		t.Errorf("updated = %d, want 1", result.Updated)
	}

	// First annotation should be updated, second should keep old hash
	content, _ := os.ReadFile(specPath)
	hashA, _ := HashFile(srcPath)
	if !strings.Contains(string(content), "src/a.go@"+hashA) {
		t.Error("first annotation should be updated")
	}
	if !strings.Contains(string(content), "src/b.go@00000000") {
		t.Error("second annotation should keep old hash")
	}
}

func TestInteractiveUpdate_shows_context(t *testing.T) {
	dir := t.TempDir()

	srcPath := filepath.Join(dir, "src", "main.go")
	os.MkdirAll(filepath.Dir(srcPath), 0755)
	os.WriteFile(srcPath, []byte("package main\n"), 0644)

	specContent := "<!-- specdrift -->\n\n<!-- source: src/main.go@00000000 -->\n\nThis is the **important** description.\n\n<!-- /source -->\n"
	specPath := filepath.Join(dir, "spec.md")
	os.WriteFile(specPath, []byte(specContent), 0644)

	input := strings.NewReader("n\n")
	var output bytes.Buffer

	InteractiveUpdate(specPath, dir, input, &output)

	out := output.String()
	if !strings.Contains(out, "important") {
		t.Errorf("output should show enclosed content, got:\n%s", out)
	}
	if !strings.Contains(out, "00000000") {
		t.Errorf("output should show old hash, got:\n%s", out)
	}
}

func TestInteractiveUpdate_no_drift(t *testing.T) {
	dir := t.TempDir()

	srcPath := filepath.Join(dir, "src", "main.go")
	os.MkdirAll(filepath.Dir(srcPath), 0755)
	os.WriteFile(srcPath, []byte("package main\n"), 0644)

	hash, _ := HashFile(srcPath)

	specContent := "<!-- specdrift -->\n\n<!-- source: src/main.go@" + hash + " -->\n\nUp to date.\n\n<!-- /source -->\n"
	specPath := filepath.Join(dir, "spec.md")
	os.WriteFile(specPath, []byte(specContent), 0644)

	input := strings.NewReader("")
	var output bytes.Buffer

	result, err := InteractiveUpdate(specPath, dir, input, &output)
	if err != nil {
		t.Fatal(err)
	}
	if result.Updated != 0 {
		t.Errorf("updated = %d, want 0", result.Updated)
	}
}

func TestInteractiveUpdate_skipped_no_declaration(t *testing.T) {
	dir := t.TempDir()

	specContent := "# Just a regular markdown file\n"
	specPath := filepath.Join(dir, "spec.md")
	os.WriteFile(specPath, []byte(specContent), 0644)

	input := strings.NewReader("")
	var output bytes.Buffer

	result, err := InteractiveUpdate(specPath, dir, input, &output)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Skipped {
		t.Error("expected Skipped=true")
	}
}
