package internal

import (
	"path/filepath"
	"sort"
	"testing"
)

func TestBuildFullGraph_ForwardAndReverse(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "spec.md"), `<!-- specdrift -->
<!-- source: main.go@TODO, lib.go@TODO -->
text
<!-- /source -->
<!-- source: other.md@TODO -->
doc ref
<!-- /source -->
`)

	writeFile(t, filepath.Join(dir, "other.md"), `<!-- specdrift -->
<!-- source: lib.go@TODO -->
text
<!-- /source -->
`)

	g := BuildFullGraph(
		[]string{filepath.Join(dir, "spec.md"), filepath.Join(dir, "other.md")},
		dir,
	)

	// Forward
	specDeps := g.Forward["spec.md"]
	sort.Strings(specDeps)
	expected := []string{"lib.go", "main.go", "other.md"}
	if !sliceEqual(specDeps, expected) {
		t.Errorf("spec.md forward: expected %v, got %v", expected, specDeps)
	}

	otherDeps := g.Forward["other.md"]
	if len(otherDeps) != 1 || otherDeps[0] != "lib.go" {
		t.Errorf("other.md forward: expected [lib.go], got %v", otherDeps)
	}

	// Reverse
	libSpecs := g.Reverse["lib.go"]
	sort.Strings(libSpecs)
	expectedReverse := []string{"other.md", "spec.md"}
	if !sliceEqual(libSpecs, expectedReverse) {
		t.Errorf("lib.go reverse: expected %v, got %v", expectedReverse, libSpecs)
	}

	mainSpecs := g.Reverse["main.go"]
	if len(mainSpecs) != 1 || mainSpecs[0] != "spec.md" {
		t.Errorf("main.go reverse: expected [spec.md], got %v", mainSpecs)
	}

	otherSpecs := g.Reverse["other.md"]
	if len(otherSpecs) != 1 || otherSpecs[0] != "spec.md" {
		t.Errorf("other.md reverse: expected [spec.md], got %v", otherSpecs)
	}
}

func TestBuildFullGraph_NoDuplicates(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "spec.md"), `<!-- specdrift -->
<!-- source: main.go@TODO -->
outer
<!-- source: main.go@TODO -->
inner (same file referenced twice)
<!-- /source -->
<!-- /source -->
`)

	g := BuildFullGraph([]string{filepath.Join(dir, "spec.md")}, dir)

	deps := g.Forward["spec.md"]
	if len(deps) != 1 {
		t.Errorf("expected 1 unique dep, got %v", deps)
	}
}

func TestBuildFullGraph_NoDeclaration(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, filepath.Join(dir, "plain.md"), `# Just a markdown file
No specdrift declaration here.
`)

	g := BuildFullGraph([]string{filepath.Join(dir, "plain.md")}, dir)

	deps := g.Forward["plain.md"]
	if len(deps) != 0 {
		t.Errorf("expected no deps for undeclared file, got %v", deps)
	}
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
