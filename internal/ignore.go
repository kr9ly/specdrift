package internal

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

const IgnoreFileName = ".specdriftignore"

// LoadIgnorePatterns reads a .specdriftignore file from the base directory.
// Returns nil if the file doesn't exist. Each non-empty, non-comment line is a pattern.
func LoadIgnorePatterns(basePath string) ([]string, error) {
	path := filepath.Join(basePath, IgnoreFileName)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns, scanner.Err()
}

// IsIgnored checks if a path matches any of the ignore patterns.
func IsIgnored(path string, patterns []string) bool {
	for _, pat := range patterns {
		// Match against the full path
		if matched, _ := filepath.Match(pat, path); matched {
			return true
		}
		// Match against the base name
		if matched, _ := filepath.Match(pat, filepath.Base(path)); matched {
			return true
		}
	}
	return false
}

// FilterIgnored removes ignored paths from a slice.
func FilterIgnored(paths []string, patterns []string) []string {
	if len(patterns) == 0 {
		return paths
	}
	var result []string
	for _, p := range paths {
		if !IsIgnored(p, patterns) {
			result = append(result, p)
		}
	}
	return result
}
