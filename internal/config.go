package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const ConfigFileName = ".specdrift"

// Config represents the specdrift project configuration.
type Config struct {
	RequireReason bool `json:"require_reason,omitempty"` // require --reason for update
}

// FindProjectRoot walks up from startDir looking for a .specdrift file.
// Returns the directory containing it, or empty string if not found.
func FindProjectRoot(startDir string) string {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ConfigFileName)); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// LoadConfig reads the .specdrift config file from basePath.
// Returns a zero Config if the file doesn't exist or is empty.
func LoadConfig(basePath string) (*Config, error) {
	path := filepath.Join(basePath, ConfigFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Init creates a .specdrift file in the specified directory.
// Returns an error if the file already exists.
func Init(dir string) error {
	path := filepath.Join(dir, ConfigFileName)
	if _, err := os.Stat(path); err == nil {
		return os.ErrExist
	}
	return os.WriteFile(path, []byte("{}\n"), 0644)
}
