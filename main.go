package main

import (
	"fmt"
	"os"

	"github.com/kr9ly/specdrift/internal"
)

const usage = `specdrift - detect drift between specs and source code

Usage:
  specdrift init
  specdrift check [--base <dir>] <file|glob>...
  specdrift update [--base <dir>] <file|glob>...

Commands:
  init    Create a .specdrift file in the current directory
  check   Check spec files for drifted source annotations
  update  Update source annotation hashes to match current files

Options:
  --base <dir>  Base directory for resolving source paths (default: .specdrift location or current directory)
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "init":
		runInit()
		return
	default:
	}

	var baseFlag string
	var rawArgs []string

	for i := 0; i < len(args); i++ {
		if args[i] == "--base" {
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "error: --base requires a directory argument")
				os.Exit(1)
			}
			baseFlag = args[i+1]
			i++
		} else {
			rawArgs = append(rawArgs, args[i])
		}
	}

	if len(rawArgs) == 0 {
		fmt.Fprintln(os.Stderr, "error: no spec files specified")
		os.Exit(1)
	}

	var files []string
	for _, arg := range rawArgs {
		if internal.IsGlobPattern(arg) {
			matches, err := internal.ExpandGlob(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: glob %q: %v\n", arg, err)
				os.Exit(1)
			}
			if len(matches) == 0 {
				fmt.Fprintf(os.Stderr, "warning: glob %q matched no files\n", arg)
			}
			files = append(files, matches...)
		} else {
			files = append(files, arg)
		}
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "error: no spec files matched")
		os.Exit(1)
	}

	basePath := resolveBasePath(baseFlag)

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

func resolveBasePath(baseFlag string) string {
	if baseFlag != "" {
		return baseFlag
	}
	if root := internal.FindProjectRoot("."); root != "" {
		return root
	}
	return "."
}

func runInit() {
	err := internal.Init(".")
	if err != nil {
		if os.IsExist(err) {
			fmt.Fprintln(os.Stderr, ".specdrift already exists")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("created .specdrift")
}

func runCheck(files []string, basePath string) int {
	graph := internal.BuildDocGraph(files, basePath)
	cycleErrors := internal.CycleErrorsByFile(graph)

	hasProblems := false
	for _, f := range files {
		rel := internal.Relativize(f, basePath)
		if cycleErr, ok := cycleErrors[rel]; ok {
			result := &internal.CheckResult{
				SpecFile: f,
				Status:   internal.CheckError,
				Error:    cycleErr,
			}
			internal.ReportText(os.Stdout, result)
			hasProblems = true
			continue
		}

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
			fmt.Printf("%s: skipped (no <!-- specdrift --> declaration)\n", f)
		} else if r.Updated > 0 {
			fmt.Printf("%s: updated %d annotation(s)\n", f, r.Updated)
		} else {
			fmt.Printf("%s: already up to date\n", f)
		}
	}
}
