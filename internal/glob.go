package internal

import (
	"os"
	"path/filepath"
	"strings"
)

// IsGlobPattern reports whether s contains glob metacharacters.
func IsGlobPattern(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

// ExpandGlob expands a glob pattern into matching file paths.
// Supports ** for recursive directory matching.
func ExpandGlob(pattern string) ([]string, error) {
	if !strings.Contains(pattern, "**") {
		return filepath.Glob(pattern)
	}

	// Split pattern at the first **
	parts := strings.SplitN(pattern, "**", 2)
	root := parts[0]
	if root == "" {
		root = "."
	} else {
		root = strings.TrimRight(root, string(filepath.Separator))
	}
	suffix := parts[1]
	if strings.HasPrefix(suffix, string(filepath.Separator)) {
		suffix = suffix[1:]
	}

	var matches []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}
		if d.IsDir() {
			return nil
		}
		if suffix == "" {
			matches = append(matches, path)
			return nil
		}
		// Match the suffix against the relative path from root
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		matched, _ := filepath.Match(suffix, filepath.Base(rel))
		if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}
