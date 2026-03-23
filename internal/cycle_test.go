package internal

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestDetectCycles_NoCycle(t *testing.T) {
	graph := map[string][]string{
		"a.md": {"b.md"},
		"b.md": {"c.md"},
		"c.md": nil,
	}
	cycles := DetectCycles(graph)
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestDetectCycles_DirectCycle(t *testing.T) {
	graph := map[string][]string{
		"a.md": {"b.md"},
		"b.md": {"a.md"},
	}
	cycles := DetectCycles(graph)
	if len(cycles) == 0 {
		t.Fatal("expected cycle, got none")
	}
	assertCycleContains(t, cycles, "a.md", "b.md")
}

func TestDetectCycles_IndirectCycle(t *testing.T) {
	graph := map[string][]string{
		"a.md": {"b.md"},
		"b.md": {"c.md"},
		"c.md": {"a.md"},
	}
	cycles := DetectCycles(graph)
	if len(cycles) == 0 {
		t.Fatal("expected cycle, got none")
	}
	assertCycleContains(t, cycles, "a.md", "b.md", "c.md")
}

func TestDetectCycles_SelfReference(t *testing.T) {
	graph := map[string][]string{
		"a.md": {"a.md"},
	}
	cycles := DetectCycles(graph)
	if len(cycles) == 0 {
		t.Fatal("expected cycle, got none")
	}
	if len(cycles[0]) != 2 || cycles[0][0] != "a.md" || cycles[0][1] != "a.md" {
		t.Errorf("expected [a.md, a.md], got %v", cycles[0])
	}
}

func TestDetectCycles_MultipleIndependent(t *testing.T) {
	graph := map[string][]string{
		"a.md": {"b.md"},
		"b.md": {"a.md"},
		"c.md": {"d.md"},
		"d.md": {"c.md"},
	}
	cycles := DetectCycles(graph)
	if len(cycles) < 2 {
		t.Errorf("expected at least 2 cycles, got %d: %v", len(cycles), cycles)
	}
}

func TestDetectCycles_EmptyGraph(t *testing.T) {
	graph := map[string][]string{}
	cycles := DetectCycles(graph)
	if len(cycles) != 0 {
		t.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestBuildDocGraph_OnlyMdEdges(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "spec.md"), `<!-- specdrift -->
<!-- source: other.md@TODO -->
some text
<!-- /source -->
<!-- source: main.go@TODO -->
code ref
<!-- /source -->
`)

	writeFile(t, filepath.Join(dir, "other.md"), `<!-- specdrift -->
<!-- source: lib.go@TODO -->
lib ref
<!-- /source -->
`)

	graph := BuildDocGraph(
		[]string{filepath.Join(dir, "spec.md"), filepath.Join(dir, "other.md")},
		dir,
	)

	specDeps := graph["spec.md"]
	if len(specDeps) != 1 || specDeps[0] != "other.md" {
		t.Errorf("spec.md deps: expected [other.md], got %v", specDeps)
	}

	otherDeps := graph["other.md"]
	if len(otherDeps) != 0 {
		t.Errorf("other.md deps: expected [], got %v", otherDeps)
	}
}

func TestBuildDocGraph_TransitiveDiscovery(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "a.md"), `<!-- specdrift -->
<!-- source: b.md@TODO -->
text
<!-- /source -->
`)

	writeFile(t, filepath.Join(dir, "b.md"), `<!-- specdrift -->
<!-- source: c.md@TODO -->
text
<!-- /source -->
`)

	writeFile(t, filepath.Join(dir, "c.md"), `<!-- specdrift -->
`)

	// Only pass a.md; b.md and c.md should be discovered transitively
	graph := BuildDocGraph([]string{filepath.Join(dir, "a.md")}, dir)

	if _, exists := graph["b.md"]; !exists {
		t.Error("b.md should be discovered transitively")
	}
	if _, exists := graph["c.md"]; !exists {
		t.Error("c.md should be discovered transitively")
	}
}

func TestBuildDocGraph_CircularReference(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "a.md"), `<!-- specdrift -->
<!-- source: b.md@TODO -->
text
<!-- /source -->
`)

	writeFile(t, filepath.Join(dir, "b.md"), `<!-- specdrift -->
<!-- source: a.md@TODO -->
text
<!-- /source -->
`)

	graph := BuildDocGraph([]string{filepath.Join(dir, "a.md")}, dir)

	cycles := DetectCycles(graph)
	if len(cycles) == 0 {
		t.Fatal("expected cycle between a.md and b.md")
	}
	assertCycleContains(t, cycles, "a.md", "b.md")
}

func TestBuildDocGraph_NonexistentTarget(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "spec.md"), `<!-- specdrift -->
<!-- source: missing.md@TODO -->
text
<!-- /source -->
`)

	graph := BuildDocGraph([]string{filepath.Join(dir, "spec.md")}, dir)

	specDeps := graph["spec.md"]
	if len(specDeps) != 1 || specDeps[0] != "missing.md" {
		t.Errorf("spec.md deps: expected [missing.md], got %v", specDeps)
	}

	// missing.md should be in graph with nil deps (no file to read)
	if deps, exists := graph["missing.md"]; !exists {
		t.Error("missing.md should be in graph")
	} else if deps != nil {
		t.Errorf("missing.md deps: expected nil, got %v", deps)
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// assertCycleContains checks that at least one cycle in cycles contains all the given nodes.
func assertCycleContains(t *testing.T, cycles [][]string, nodes ...string) {
	t.Helper()
	for _, cycle := range cycles {
		cycleNodes := make(map[string]bool)
		for _, n := range cycle {
			cycleNodes[n] = true
		}
		allFound := true
		for _, n := range nodes {
			if !cycleNodes[n] {
				allFound = false
				break
			}
		}
		if allFound {
			return
		}
	}
	// Format for error message
	sort.Strings(nodes)
	t.Errorf("no cycle contains all of %v; cycles: %v", nodes, cycles)
}
