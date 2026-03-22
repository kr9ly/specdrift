package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	declRe        = regexp.MustCompile(`<!--\s*specdrift(?:\s+v(\d+))?\s*-->`)
	sourceOpenRe  = regexp.MustCompile(`<!--\s*source:\s*(.*?)\s*-->`)
	sourceCloseRe = regexp.MustCompile(`<!--\s*/source\s*-->`)
	sourceRefRe   = regexp.MustCompile(`(\S+)@([a-f0-9]{8}|TODO)`)
	todoBareRe    = regexp.MustCompile(`^\s*TODO\s*$`)
	fenceRe       = regexp.MustCompile("^\\s*(`{3,}|~{3,})")
)

// Status represents the drift check result for a source reference.
type Status int

const (
	StatusUnchecked Status = iota
	StatusOK
	StatusDrift
	StatusMissing
	StatusTodo
)

func (s Status) String() string {
	switch s {
	case StatusOK:
		return "OK"
	case StatusDrift:
		return "DRIFT"
	case StatusMissing:
		return "MISSING"
	case StatusTodo:
		return "TODO"
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

// CurrentVersion is the latest supported specdrift format version.
const CurrentVersion = 1

// ParseResult holds the result of parsing a spec file's annotations.
type ParseResult struct {
	Declared    bool
	Version     int // 0 if not declared, otherwise 1+
	Annotations []*Annotation
}

// ParseAnnotations extracts source annotations from markdown content,
// building a tree from nested open/close tags using a stack.
func ParseAnnotations(content string) (*ParseResult, error) {
	lines := strings.Split(content, "\n")
	var roots []*Annotation
	var stack []*Annotation
	declared := false
	version := 0
	inFence := false
	fenceMarker := ""

	for i, line := range lines {
		lineNum := i + 1

		// Skip fenced code blocks (``` or ~~~)
		if fm := fenceRe.FindStringSubmatch(line); fm != nil {
			if !inFence {
				inFence = true
				fenceMarker = fm[1]
			} else if fm[1][0] == fenceMarker[0] && len(fm[1]) >= len(fenceMarker) {
				inFence = false
				fenceMarker = ""
			}
			continue
		}
		if inFence {
			continue
		}

		// Strip inline code spans before matching annotations
		stripped := stripInlineCode(line)

		if !declared {
			if m := declRe.FindStringSubmatch(stripped); m != nil {
				declared = true
				if m[1] != "" {
					v, _ := strconv.Atoi(m[1])
					version = v
				} else {
					version = 1
				}
				continue
			}
		}

		if m := sourceOpenRe.FindStringSubmatch(stripped); m != nil {
			a := &Annotation{
				Line: lineNum,
			}
			if todoBareRe.MatchString(m[1]) {
				a.Sources = append(a.Sources, SourceRef{
					Status: StatusTodo,
				})
			} else {
				refs := sourceRefRe.FindAllStringSubmatch(m[1], -1)
				if len(refs) == 0 {
					return nil, fmt.Errorf("line %d: source tag contains no valid path@hash references", lineNum)
				}
				for _, ref := range refs {
					sr := SourceRef{
						Path:         ref[1],
						ExpectedHash: ref[2],
					}
					if ref[2] == "TODO" {
						sr.ExpectedHash = ""
						sr.Status = StatusTodo
					}
					a.Sources = append(a.Sources, sr)
				}
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

		if sourceCloseRe.MatchString(stripped) {
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
		Version:     version,
		Annotations: roots,
	}, nil
}

// stripInlineCode removes inline code spans (backtick-delimited) from a line
// so that annotations inside them are not matched.
func stripInlineCode(line string) string {
	result := []byte(line)
	i := 0
	for i < len(result) {
		if result[i] == '`' {
			// Count opening backticks
			start := i
			ticks := 0
			for i < len(result) && result[i] == '`' {
				ticks++
				i++
			}
			// Find matching closing backticks
			closer := strings.Repeat("`", ticks)
			end := strings.Index(string(result[i:]), closer)
			if end >= 0 {
				// Replace the entire span (including delimiters) with spaces
				for j := start; j < i+end+ticks; j++ {
					result[j] = ' '
				}
				i = i + end + ticks
			}
			// If no closing backticks found, leave as-is
		} else {
			i++
		}
	}
	return string(result)
}
