package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BuildDocGraph builds a dependency graph from spec files,
// including only edges where the target is a .md file.
// Keys and values are basePath-relative paths.
// It recursively follows .md references to discover transitive dependencies.
func BuildDocGraph(files []string, basePath string) map[string][]string {
	graph := make(map[string][]string)
	var queue []string

	for _, f := range files {
		rel := relativize(f, basePath)
		if _, exists := graph[rel]; !exists {
			graph[rel] = nil // mark as visited
			queue = append(queue, f)
		}
	}

	for len(queue) > 0 {
		f := queue[0]
		queue = queue[1:]
		rel := relativize(f, basePath)

		deps := extractDocDeps(f, basePath)
		graph[rel] = deps

		for _, dep := range deps {
			if _, exists := graph[dep]; !exists {
				graph[dep] = nil
				depPath := filepath.Join(basePath, dep)
				queue = append(queue, depPath)
			}
		}
	}

	return graph
}

// extractDocDeps parses a file and returns basePath-relative paths of .md dependencies.
func extractDocDeps(filePath string, basePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	parsed, err := ParseAnnotations(string(content))
	if err != nil || !parsed.Declared {
		return nil
	}

	var deps []string
	collectDocRefs(parsed.Annotations, &deps)
	return deps
}

func collectDocRefs(annotations []*Annotation, deps *[]string) {
	for _, a := range annotations {
		for _, ref := range a.Sources {
			if ref.Path != "" && strings.HasSuffix(ref.Path, ".md") {
				*deps = append(*deps, ref.Path)
			}
		}
		collectDocRefs(a.Children, deps)
	}
}

// DetectCycles finds all cycles in a directed graph using DFS with coloring.
// Returns a list of cycles, where each cycle is a path like [A, B, C, A].
func DetectCycles(graph map[string][]string) [][]string {
	const (
		white = 0
		gray  = 1
		black = 2
	)

	color := make(map[string]int)
	var path []string
	var cycles [][]string

	var dfs func(node string)
	dfs = func(node string) {
		color[node] = gray
		path = append(path, node)

		for _, next := range graph[node] {
			if color[next] == gray {
				start := -1
				for i, p := range path {
					if p == next {
						start = i
						break
					}
				}
				if start >= 0 {
					cycle := make([]string, len(path)-start+1)
					copy(cycle, path[start:])
					cycle[len(cycle)-1] = next
					cycles = append(cycles, cycle)
				}
			} else if color[next] == white {
				dfs(next)
			}
		}

		path = path[:len(path)-1]
		color[node] = black
	}

	for node := range graph {
		if color[node] == white {
			dfs(node)
		}
	}

	return cycles
}

// CycleErrorsByFile detects cycles in the graph and returns an error for each file
// that participates in a cycle. The map key is a basePath-relative path.
func CycleErrorsByFile(graph map[string][]string) map[string]error {
	cycles := DetectCycles(graph)
	if len(cycles) == 0 {
		return nil
	}

	errs := make(map[string]error)
	for _, cycle := range cycles {
		msg := fmt.Sprintf("circular reference: %s", strings.Join(cycle, " -> "))
		// Mark all nodes in the cycle (except the duplicated last element)
		for _, node := range cycle[:len(cycle)-1] {
			if _, exists := errs[node]; !exists {
				errs[node] = fmt.Errorf("%s", msg)
			}
		}
	}
	return errs
}

// Relativize returns a basePath-relative path.
func Relativize(path string, basePath string) string {
	return relativize(path, basePath)
}

func relativize(path string, basePath string) string {
	rel, err := filepath.Rel(basePath, path)
	if err != nil {
		return filepath.Clean(path)
	}
	return rel
}
