package internal

import (
	"os"
	"sort"
)

// DepGraph represents a full dependency graph from spec files to all referenced sources.
// Keys are basePath-relative spec file paths, values are their referenced source paths.
type DepGraph struct {
	// Forward maps spec file -> referenced sources (both code and docs)
	Forward map[string][]string
	// Reverse maps source file -> spec files that reference it
	Reverse map[string][]string
}

// BuildFullGraph builds a complete dependency graph from spec files,
// including all source references (not just .md files).
func BuildFullGraph(files []string, basePath string) *DepGraph {
	forward := make(map[string][]string)

	for _, f := range files {
		rel := relativize(f, basePath)
		deps := extractAllDeps(f)
		forward[rel] = deps
	}

	reverse := make(map[string][]string)
	for spec, deps := range forward {
		for _, dep := range deps {
			reverse[dep] = append(reverse[dep], spec)
		}
	}

	// Sort for stable output
	for k := range forward {
		sort.Strings(forward[k])
	}
	for k := range reverse {
		sort.Strings(reverse[k])
	}

	return &DepGraph{Forward: forward, Reverse: reverse}
}

// extractAllDeps parses a file and returns all source ref paths.
func extractAllDeps(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	parsed, err := ParseAnnotations(string(content))
	if err != nil || !parsed.Declared {
		return nil
	}

	var deps []string
	seen := make(map[string]bool)
	collectAllRefs(parsed.Annotations, &deps, seen)
	return deps
}

func collectAllRefs(annotations []*Annotation, deps *[]string, seen map[string]bool) {
	for _, a := range annotations {
		for _, ref := range a.Sources {
			if ref.Path != "" && !seen[ref.Path] {
				seen[ref.Path] = true
				*deps = append(*deps, ref.Path)
			}
		}
		collectAllRefs(a.Children, deps, seen)
	}
}
