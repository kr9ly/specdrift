package internal

import (
	"os"
	"path/filepath"
)

const ConfigFileName = ".specdrift"

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

// Init creates a .specdrift file in the specified directory.
// Returns an error if the file already exists.
func Init(dir string) error {
	path := filepath.Join(dir, ConfigFileName)
	if _, err := os.Stat(path); err == nil {
		return os.ErrExist
	}
	return os.WriteFile(path, []byte("{}\n"), 0644)
}
