package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UpdateResult represents the outcome of an update operation.
type UpdateResult struct {
	Skipped bool
	Updated int
}

// Update parses a spec file and rewrites all source annotation hashes to match current files.
func Update(specFile string, basePath string) (*UpdateResult, error) {
	content, err := os.ReadFile(specFile)
	if err != nil {
		return nil, err
	}

	text := string(content)

	if !declRe.MatchString(text) {
		return &UpdateResult{Skipped: true}, nil
	}

	updated := 0

	result := sourceOpenRe.ReplaceAllStringFunc(text, func(match string) string {
		m := sourceOpenRe.FindStringSubmatch(match)
		if m == nil {
			return match
		}
		refsStr := m[1]

		// Bare TODO — leave as-is
		if todoBareRe.MatchString(refsStr) {
			return match
		}

		refs := sourceRefRe.FindAllStringSubmatch(refsStr, -1)
		if len(refs) == 0 {
			return match
		}

		var parts []string
		for _, ref := range refs {
			path := ref[1]
			oldHash := ref[2]

			if oldHash == "TODO" {
				fullPath := filepath.Join(basePath, path)
				newHash, err := HashFile(fullPath)
				if err != nil {
					// File not yet created — keep TODO
					parts = append(parts, fmt.Sprintf("%s@TODO", path))
					continue
				}
				updated++
				parts = append(parts, fmt.Sprintf("%s@%s", path, newHash))
				continue
			}

			fullPath := filepath.Join(basePath, path)
			newHash, err := HashFile(fullPath)
			if err != nil {
				// File missing — keep old hash
				parts = append(parts, fmt.Sprintf("%s@%s", path, oldHash))
				continue
			}

			if newHash != oldHash {
				updated++
			}
			parts = append(parts, fmt.Sprintf("%s@%s", path, newHash))
		}

		return fmt.Sprintf("<!-- source: %s -->", strings.Join(parts, ", "))
	})

	if updated > 0 {
		if err := os.WriteFile(specFile, []byte(result), 0644); err != nil {
			return nil, err
		}
	}

	return &UpdateResult{Updated: updated}, nil
}
