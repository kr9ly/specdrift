package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_empty(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ConfigFileName), []byte("{}\n"), 0644)

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.UpdateMode != "" {
		t.Errorf("update_mode = %q, want empty", cfg.UpdateMode)
	}
}

func TestLoadConfig_interactive(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ConfigFileName), []byte(`{"update_mode": "interactive"}`+"\n"), 0644)

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.UpdateMode != "interactive" {
		t.Errorf("update_mode = %q, want %q", cfg.UpdateMode, "interactive")
	}
}

func TestLoadConfig_missing_file(t *testing.T) {
	dir := t.TempDir()

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.UpdateMode != "" {
		t.Errorf("update_mode = %q, want empty", cfg.UpdateMode)
	}
}

func TestLoadConfig_invalid_json(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ConfigFileName), []byte("not json"), 0644)

	_, err := LoadConfig(dir)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
