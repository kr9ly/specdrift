package internal

import (
	"bufio"
	"fmt"
	"io"
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

// InteractiveUpdate checks a spec file and prompts the user for each drifted annotation.
// Returns the update result. The prompter reads from r and writes prompts to w.
func InteractiveUpdate(specFile string, basePath string, r io.Reader, w io.Writer) (*UpdateResult, error) {
	content, err := os.ReadFile(specFile)
	if err != nil {
		return nil, err
	}

	text := string(content)
	lines := strings.Split(text, "\n")

	if !declRe.MatchString(text) {
		return &UpdateResult{Skipped: true}, nil
	}

	parsed, err := ParseAnnotations(text)
	if err != nil {
		return nil, err
	}

	if len(parsed.Annotations) == 0 {
		return &UpdateResult{Updated: 0}, nil
	}

	checkAnnotations(parsed.Annotations, basePath)

	// Collect drifted annotations (flattened)
	var drifted []*Annotation
	collectDrifted(parsed.Annotations, &drifted)

	if len(drifted) == 0 {
		return &UpdateResult{Updated: 0}, nil
	}

	scanner := bufio.NewScanner(r)
	updated := 0
	// Track which source tag lines to update (1-based line number -> true)
	acceptedLines := make(map[int]bool)

	for i, a := range drifted {
		fmt.Fprintf(w, "\n--- [%d/%d] %s line %d ", i+1, len(drifted), specFile, a.Line)
		printDriftSummary(w, a)
		fmt.Fprintln(w)

		// Show enclosed content
		printAnnotationContext(w, lines, a)
		fmt.Fprintln(w)

		// Show hash changes
		printHashChanges(w, a)
		fmt.Fprintln(w)

		fmt.Fprintf(w, "Update this annotation? [y/n/q] ")

		if !scanner.Scan() {
			break
		}
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))

		switch answer {
		case "y", "yes":
			acceptedLines[a.Line] = true
			updated++
		case "q", "quit":
			goto done
		default:
			// skip
		}
	}

done:
	if updated == 0 {
		return &UpdateResult{Updated: 0}, nil
	}

	// Apply accepted updates by replacing source tags at accepted lines
	result := applySelectiveUpdate(lines, acceptedLines, basePath)

	if err := os.WriteFile(specFile, []byte(result), 0644); err != nil {
		return nil, err
	}

	return &UpdateResult{Updated: updated}, nil
}

// collectDrifted flattens the annotation tree and collects annotations that have at least one drifted/missing/todo source.
func collectDrifted(annotations []*Annotation, out *[]*Annotation) {
	for _, a := range annotations {
		if hasDrift(a) {
			*out = append(*out, a)
		}
		collectDrifted(a.Children, out)
	}
}

func hasDrift(a *Annotation) bool {
	for _, ref := range a.Sources {
		if ref.Status == StatusDrift || ref.Status == StatusMissing || ref.Status == StatusTodo {
			return true
		}
	}
	return false
}

func printDriftSummary(w io.Writer, a *Annotation) {
	var statuses []string
	for _, ref := range a.Sources {
		if ref.Status != StatusOK {
			statuses = append(statuses, fmt.Sprintf("%s[%s]", ref.Path, ref.Status))
		}
	}
	fmt.Fprint(w, strings.Join(statuses, ", "))
}

func printAnnotationContext(w io.Writer, lines []string, a *Annotation) {
	start := a.Line // 1-based, the source open tag line
	end := a.EndLine

	if end == 0 {
		// No closing tag (shouldn't happen with valid parse, but handle gracefully)
		end = start
	}

	// Show lines between open and close tags (exclusive of tags themselves)
	if end-start <= 1 {
		fmt.Fprintf(w, "  (no enclosed content)\n")
		return
	}

	for i := start; i < end-1; i++ {
		if i < len(lines) {
			fmt.Fprintf(w, "  %s\n", lines[i])
		}
	}
}

func printHashChanges(w io.Writer, a *Annotation) {
	for _, ref := range a.Sources {
		switch ref.Status {
		case StatusDrift:
			fmt.Fprintf(w, "  %s: %s -> %s\n", ref.Path, ref.ExpectedHash, ref.ActualHash)
		case StatusMissing:
			fmt.Fprintf(w, "  %s: MISSING (file not found)\n", ref.Path)
		case StatusTodo:
			if ref.Path != "" {
				fmt.Fprintf(w, "  %s: TODO\n", ref.Path)
			} else {
				fmt.Fprintf(w, "  (bare TODO)\n")
			}
		}
	}
}

// applySelectiveUpdate rewrites source tags only on accepted lines.
func applySelectiveUpdate(lines []string, acceptedLines map[int]bool, basePath string) string {
	var result []string
	for i, line := range lines {
		lineNum := i + 1
		if acceptedLines[lineNum] && sourceOpenRe.MatchString(line) {
			result = append(result, updateSourceTag(line, basePath))
		} else {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

// updateSourceTag rewrites a single source open tag line with current hashes.
func updateSourceTag(line string, basePath string) string {
	return sourceOpenRe.ReplaceAllStringFunc(line, func(match string) string {
		m := sourceOpenRe.FindStringSubmatch(match)
		if m == nil {
			return match
		}
		refsStr := m[1]

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

			fullPath := filepath.Join(basePath, path)
			newHash, err := HashFile(fullPath)
			if err != nil {
				if oldHash == "TODO" {
					parts = append(parts, fmt.Sprintf("%s@TODO", path))
				} else {
					parts = append(parts, fmt.Sprintf("%s@%s", path, oldHash))
				}
				continue
			}
			parts = append(parts, fmt.Sprintf("%s@%s", path, newHash))
		}

		return fmt.Sprintf("<!-- source: %s -->", strings.Join(parts, ", "))
	})
}
