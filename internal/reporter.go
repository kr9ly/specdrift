package internal

import (
	"fmt"
	"io"
	"strings"
)

// ReportText writes a human-readable drift report to w.
func ReportText(w io.Writer, result *CheckResult) {
	switch result.Status {
	case CheckSkipped:
		fmt.Fprintf(w, "%s: skipped (no <!-- spec-drift --> declaration)\n", result.SpecFile)
		return
	case CheckEmpty:
		fmt.Fprintf(w, "%s: warning: declared <!-- spec-drift --> but no source annotations found\n", result.SpecFile)
		return
	case CheckError:
		fmt.Fprintf(w, "%s: error: %v\n", result.SpecFile, result.Error)
		return
	}

	fmt.Fprintf(w, "%s:\n", result.SpecFile)
	reportAnnotations(w, result.Annotations, 1)

	ok, drift, missing := result.CountByStatus()
	total := ok + drift + missing
	fmt.Fprintf(w, "\n%d drifted, %d ok, %d missing (total: %d)\n", drift, ok, missing, total)
}

func reportAnnotations(w io.Writer, annotations []*Annotation, depth int) {
	indent := strings.Repeat("  ", depth)
	for i, a := range annotations {
		aPrefix := "├─"
		if i == len(annotations)-1 && len(a.Children) == 0 {
			aPrefix = "└─"
		}

		for j, ref := range a.Sources {
			prefix := aPrefix
			if j > 0 {
				prefix = "│ "
			}

			switch ref.Status {
			case StatusOK:
				fmt.Fprintf(w, "%s%s [OK] %s line %d\n", indent, prefix, ref.Path, a.Line)
			case StatusDrift:
				fmt.Fprintf(w, "%s%s [DRIFT] %s (expected: %s, actual: %s) line %d\n",
					indent, prefix, ref.Path, ref.ExpectedHash, ref.ActualHash, a.Line)
			case StatusMissing:
				fmt.Fprintf(w, "%s%s [MISSING] %s line %d\n", indent, prefix, ref.Path, a.Line)
			}
		}

		if len(a.Children) > 0 {
			reportAnnotations(w, a.Children, depth+1)
		}
	}
}
