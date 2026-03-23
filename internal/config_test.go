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
	if cfg.RequireReason {
		t.Error("require_reason should be false by default")
	}
}

func TestLoadConfig_require_reason(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ConfigFileName), []byte(`{"require_reason": true}`+"\n"), 0644)

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.RequireReason {
		t.Error("require_reason should be true")
	}
}

func TestLoadConfig_missing_file(t *testing.T) {
	dir := t.TempDir()

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RequireReason {
		t.Error("require_reason should be false when file is missing")
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
