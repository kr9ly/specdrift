package internal

import "sort"

// CoverageResult holds the result of a documentation coverage analysis.
type CoverageResult struct {
	Covered    []CoveredFile
	NotCovered []string
	Total      int
}

// CoveredFile represents a source file that is referenced by at least one spec.
type CoveredFile struct {
	Source string
	Specs  []string
}

// ComputeCoverage compares source files against the reverse graph to find coverage.
func ComputeCoverage(reverse map[string][]string, sourceFiles []string) *CoverageResult {
	result := &CoverageResult{Total: len(sourceFiles)}

	sort.Strings(sourceFiles)

	for _, src := range sourceFiles {
		if specs, ok := reverse[src]; ok && len(specs) > 0 {
			result.Covered = append(result.Covered, CoveredFile{
				Source: src,
				Specs:  specs,
			})
		} else {
			result.NotCovered = append(result.NotCovered, src)
		}
	}

	return result
}
