package internal

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	declRe        = regexp.MustCompile(`<!--\s*spec-drift\s*-->`)
	sourceOpenRe  = regexp.MustCompile(`<!--\s*source:\s*(.*?)\s*-->`)
	sourceCloseRe = regexp.MustCompile(`<!--\s*/source\s*-->`)
	sourceRefRe   = regexp.MustCompile(`(\S+)@([a-f0-9]{8})`)
)

// Status represents the drift check result for a source reference.
type Status int

const (
	StatusUnchecked Status = iota
	StatusOK
	StatusDrift
	StatusMissing
)

func (s Status) String() string {
	switch s {
	case StatusOK:
		return "OK"
	case StatusDrift:
		return "DRIFT"
	case StatusMissing:
		return "MISSING"
	default:
		return "UNCHECKED"
	}
}

// SourceRef represents a single file reference within an annotation.
type SourceRef struct {
	Path         string
	ExpectedHash string
	ActualHash   string
	Status       Status
}

// Annotation represents a source annotation scope with one or more file references and optional children.
type Annotation struct {
	Sources  []SourceRef
	Line     int
	Children []*Annotation
}

// ParseResult holds the result of parsing a spec file's annotations.
type ParseResult struct {
	Declared    bool
	Annotations []*Annotation
}

// ParseAnnotations extracts source annotations from markdown content,
// building a tree from nested open/close tags using a stack.
func ParseAnnotations(content string) (*ParseResult, error) {
	lines := strings.Split(content, "\n")
	var roots []*Annotation
	var stack []*Annotation
	declared := false

	for i, line := range lines {
		lineNum := i + 1

		if !declared && declRe.MatchString(line) {
			declared = true
			continue
		}

		if m := sourceOpenRe.FindStringSubmatch(line); m != nil {
			refs := sourceRefRe.FindAllStringSubmatch(m[1], -1)
			if len(refs) == 0 {
				return nil, fmt.Errorf("line %d: source tag contains no valid path@hash references", lineNum)
			}
			a := &Annotation{
				Line: lineNum,
			}
			for _, ref := range refs {
				a.Sources = append(a.Sources, SourceRef{
					Path:         ref[1],
					ExpectedHash: ref[2],
				})
			}
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, a)
			} else {
				roots = append(roots, a)
			}
			stack = append(stack, a)
			continue
		}

		if sourceCloseRe.MatchString(line) {
			if len(stack) == 0 {
				return nil, fmt.Errorf("line %d: unexpected closing tag <!-- /source --> without matching open tag", lineNum)
			}
			stack = stack[:len(stack)-1]
			continue
		}
	}

	if len(stack) > 0 {
		unclosed := stack[len(stack)-1]
		return nil, fmt.Errorf("line %d: unclosed source tag for %s", unclosed.Line, unclosed.Sources[0].Path)
	}

	return &ParseResult{
		Declared:    declared,
		Annotations: roots,
	}, nil
}
