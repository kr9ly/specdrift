package main

import (
	"fmt"
	"os"

	"github.com/kr9ly/spec-drift/internal"
)

const usage = `spec-drift - detect drift between specs and source code

Usage:
  spec-drift check [--base <dir>] <file>...
  spec-drift update [--base <dir>] <file>...

Commands:
  check   Check spec files for drifted source annotations
  update  Update source annotation hashes to match current files

Options:
  --base <dir>  Base directory for resolving source paths (default: current directory)
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	basePath := "."
	var files []string

	for i := 0; i < len(args); i++ {
		if args[i] == "--base" {
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "error: --base requires a directory argument")
				os.Exit(1)
			}
			basePath = args[i+1]
			i++
		} else {
			files = append(files, args[i])
		}
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "error: no spec files specified")
		os.Exit(1)
	}

	switch cmd {
	case "check":
		exitCode := runCheck(files, basePath)
		os.Exit(exitCode)
	case "update":
		runUpdate(files, basePath)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command %q\n", cmd)
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}
}

func runCheck(files []string, basePath string) int {
	hasProblems := false
	for _, f := range files {
		result := internal.Check(f, basePath)
		internal.ReportText(os.Stdout, result)

		switch result.Status {
		case internal.CheckDrifted, internal.CheckEmpty, internal.CheckError:
			hasProblems = true
		}
	}
	if hasProblems {
		return 1
	}
	return 0
}

func runUpdate(files []string, basePath string) {
	for _, f := range files {
		r, err := internal.Update(f, basePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s: %v\n", f, err)
			continue
		}
		if r.Skipped {
			fmt.Printf("%s: skipped (no <!-- spec-drift --> declaration)\n", f)
		} else if r.Updated > 0 {
			fmt.Printf("%s: updated %d annotation(s)\n", f, r.Updated)
		} else {
			fmt.Printf("%s: already up to date\n", f)
		}
	}
}
