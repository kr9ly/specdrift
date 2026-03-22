package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

// CheckStatus represents the overall status of a spec file check.
type CheckStatus int

const (
	CheckOK      CheckStatus = iota
	CheckDrifted             // at least one source drifted or missing
	CheckEmpty               // declared but no annotations
	CheckSkipped             // no specdrift declaration
	CheckError               // parse or I/O error
)

// CheckResult holds the check results for a single spec file.
type CheckResult struct {
	SpecFile    string
	Status      CheckStatus
	Annotations []*Annotation
	Error       error
}

// Check parses a spec file and checks all source annotations against actual file hashes.
// basePath is the directory from which source paths are resolved.
func Check(specFile string, basePath string) *CheckResult {
	content, err := os.ReadFile(specFile)
	if err != nil {
		return &CheckResult{SpecFile: specFile, Status: CheckError, Error: err}
	}

	parsed, err := ParseAnnotations(string(content))
	if err != nil {
		return &CheckResult{SpecFile: specFile, Status: CheckError, Error: err}
	}

	if !parsed.Declared {
		return &CheckResult{SpecFile: specFile, Status: CheckSkipped}
	}

	if parsed.Version > CurrentVersion {
		return &CheckResult{
			SpecFile: specFile,
			Status:   CheckError,
			Error:    fmt.Errorf("unsupported version v%d (this tool supports up to v%d)", parsed.Version, CurrentVersion),
		}
	}

	if len(parsed.Annotations) == 0 {
		return &CheckResult{SpecFile: specFile, Status: CheckEmpty}
	}

	checkAnnotations(parsed.Annotations, basePath)

	result := &CheckResult{
		SpecFile:    specFile,
		Status:      CheckOK,
		Annotations: parsed.Annotations,
	}

	_, drift, missing, todo := result.CountByStatus()
	if drift > 0 || missing > 0 || todo > 0 {
		result.Status = CheckDrifted
	}

	return result
}

func checkAnnotations(annotations []*Annotation, basePath string) {
	for _, a := range annotations {
		for i := range a.Sources {
			ref := &a.Sources[i]
			if ref.Status == StatusTodo {
				continue
			}
			fullPath := filepath.Join(basePath, ref.Path)
			hash, err := HashFile(fullPath)
			if err != nil {
				ref.Status = StatusMissing
			} else {
				ref.ActualHash = hash
				if hash == ref.ExpectedHash {
					ref.Status = StatusOK
				} else {
					ref.Status = StatusDrift
				}
			}
		}
		checkAnnotations(a.Children, basePath)
	}
}

// CountByStatus returns counts of OK, DRIFT, MISSING, and TODO source refs (flattened).
func (r *CheckResult) CountByStatus() (ok, drift, missing, todo int) {
	countAnnotations(r.Annotations, &ok, &drift, &missing, &todo)
	return
}

func countAnnotations(annotations []*Annotation, ok, drift, missing, todo *int) {
	for _, a := range annotations {
		for _, ref := range a.Sources {
			switch ref.Status {
			case StatusOK:
				*ok++
			case StatusDrift:
				*drift++
			case StatusMissing:
				*missing++
			case StatusTodo:
				*todo++
			}
		}
		countAnnotations(a.Children, ok, drift, missing, todo)
	}
}
