package internal

import (
	"testing"
)

func TestComputeCoverage_Basic(t *testing.T) {
	reverse := map[string][]string{
		"main.go":    {"docs/spec/commands.md"},
		"checker.go": {"docs/spec/commands.md"},
	}

	sources := []string{"main.go", "checker.go", "utils.go"}
	result := ComputeCoverage(reverse, sources)

	if result.Total != 3 {
		t.Errorf("total: expected 3, got %d", result.Total)
	}
	if len(result.Covered) != 2 {
		t.Errorf("covered: expected 2, got %d", len(result.Covered))
	}
	if len(result.NotCovered) != 1 || result.NotCovered[0] != "utils.go" {
		t.Errorf("not covered: expected [utils.go], got %v", result.NotCovered)
	}
}

func TestComputeCoverage_AllCovered(t *testing.T) {
	reverse := map[string][]string{
		"a.go": {"spec.md"},
		"b.go": {"spec.md"},
	}

	result := ComputeCoverage(reverse, []string{"a.go", "b.go"})

	if len(result.NotCovered) != 0 {
		t.Errorf("expected all covered, got not covered: %v", result.NotCovered)
	}
	if len(result.Covered) != 2 {
		t.Errorf("expected 2 covered, got %d", len(result.Covered))
	}
}

func TestComputeCoverage_NoneCovered(t *testing.T) {
	reverse := map[string][]string{}

	result := ComputeCoverage(reverse, []string{"a.go", "b.go"})

	if len(result.Covered) != 0 {
		t.Errorf("expected none covered, got %v", result.Covered)
	}
	if len(result.NotCovered) != 2 {
		t.Errorf("expected 2 not covered, got %d", len(result.NotCovered))
	}
}

func TestComputeCoverage_Empty(t *testing.T) {
	result := ComputeCoverage(map[string][]string{}, []string{})

	if result.Total != 0 {
		t.Errorf("expected 0 total, got %d", result.Total)
	}
}

func TestComputeCoverage_MultipleSpecs(t *testing.T) {
	reverse := map[string][]string{
		"config.go": {"spec/commands.md", "spec/config.md"},
	}

	result := ComputeCoverage(reverse, []string{"config.go"})

	if len(result.Covered) != 1 {
		t.Fatalf("expected 1 covered, got %d", len(result.Covered))
	}
	if len(result.Covered[0].Specs) != 2 {
		t.Errorf("expected 2 specs, got %v", result.Covered[0].Specs)
	}
}
