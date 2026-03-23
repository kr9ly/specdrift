package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/kr9ly/specdrift/internal"
)

const usage = `specdrift - detect drift between specs and source code

Usage:
  specdrift init
  specdrift check [--base <dir>] <file|glob>...
  specdrift update [--base <dir>] [-i] <file|glob>...
  specdrift graph [--base <dir>] [--reverse] <file|glob>...
  specdrift coverage [--base <dir>] --src <glob>... <file|glob>...

Commands:
  init      Create a .specdrift file in the current directory
  check     Check spec files for drifted source annotations
  update    Update source annotation hashes to match current files
  graph     Show dependency graph between spec files and source files
  coverage  Show documentation coverage of source files

Options:
  --base <dir>   Base directory for resolving source paths (default: .specdrift location or current directory)
  -i             Interactive update: prompt for each drifted annotation
  --reverse      Show reverse graph (source -> specs that reference it)
  --src <glob>   Source files to measure coverage against (repeatable, used by coverage)
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
	var reverseFlag bool
	var interactiveFlag bool
	var srcPatterns []string
	var rawArgs []string

	for i := 0; i < len(args); i++ {
		if args[i] == "--base" {
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "error: --base requires a directory argument")
				os.Exit(1)
			}
			baseFlag = args[i+1]
			i++
		} else if args[i] == "--reverse" {
			reverseFlag = true
		} else if args[i] == "-i" {
			interactiveFlag = true
		} else if args[i] == "--src" {
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "error: --src requires a glob argument")
				os.Exit(1)
			}
			srcPatterns = append(srcPatterns, args[i+1])
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

	ignorePatterns, err := internal.LoadIgnorePatterns(basePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: reading %s: %v\n", internal.IgnoreFileName, err)
		os.Exit(1)
	}

	// Load config for update_mode
	cfg, cfgErr := internal.LoadConfig(basePath)
	if cfgErr != nil {
		fmt.Fprintf(os.Stderr, "warning: reading config: %v\n", cfgErr)
		cfg = &internal.Config{}
	}

	switch cmd {
	case "check":
		exitCode := runCheck(files, basePath)
		os.Exit(exitCode)
	case "update":
		interactive := interactiveFlag || cfg.UpdateMode == "interactive"
		if interactive {
			runInteractiveUpdate(files, basePath)
		} else {
			runUpdate(files, basePath)
		}
	case "graph":
		runGraph(files, basePath, reverseFlag, ignorePatterns)
	case "coverage":
		exitCode := runCoverage(files, basePath, srcPatterns, ignorePatterns)
		os.Exit(exitCode)
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

func runGraph(files []string, basePath string, reverse bool, ignorePatterns []string) {
	g := internal.BuildFullGraph(files, basePath)

	var m map[string][]string
	if reverse {
		m = g.Reverse
	} else {
		m = g.Forward
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if reverse && internal.IsIgnored(k, ignorePatterns) {
			continue
		}
		fmt.Println(k)
		for _, v := range m[k] {
			fmt.Printf("  -> %s\n", v)
		}
	}
}

func runCoverage(specFiles []string, basePath string, srcPatterns []string, ignorePatterns []string) int {
	if len(srcPatterns) == 0 {
		fmt.Fprintln(os.Stderr, "error: --src is required for coverage command")
		return 1
	}

	// Expand source patterns
	var srcFiles []string
	for _, pat := range srcPatterns {
		if internal.IsGlobPattern(pat) {
			matches, err := internal.ExpandGlob(pat)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: src glob %q: %v\n", pat, err)
				return 1
			}
			srcFiles = append(srcFiles, matches...)
		} else {
			srcFiles = append(srcFiles, pat)
		}
	}

	if len(srcFiles) == 0 {
		fmt.Fprintln(os.Stderr, "error: no source files matched")
		return 1
	}

	// Relativize and filter source files
	var relSrcFiles []string
	for _, f := range srcFiles {
		relSrcFiles = append(relSrcFiles, internal.Relativize(f, basePath))
	}
	relSrcFiles = internal.FilterIgnored(relSrcFiles, ignorePatterns)

	if len(relSrcFiles) == 0 {
		fmt.Fprintln(os.Stderr, "error: all source files excluded by .specdriftignore")
		return 1
	}

	g := internal.BuildFullGraph(specFiles, basePath)
	result := internal.ComputeCoverage(g.Reverse, relSrcFiles)

	// Print coverage percentage
	covered := len(result.Covered)
	fmt.Printf("Coverage: %d/%d", covered, result.Total)
	if result.Total > 0 {
		fmt.Printf(" (%.1f%%)", float64(covered)/float64(result.Total)*100)
	}
	fmt.Println()

	// Print covered files
	if len(result.Covered) > 0 {
		fmt.Println("\nCovered:")
		for _, cf := range result.Covered {
			specs := ""
			for i, s := range cf.Specs {
				if i > 0 {
					specs += ", "
				}
				specs += s
			}
			fmt.Printf("  %s  <- %s\n", cf.Source, specs)
		}
	}

	// Print not covered files
	if len(result.NotCovered) > 0 {
		fmt.Println("\nNot covered:")
		for _, f := range result.NotCovered {
			fmt.Printf("  %s\n", f)
		}
	}

	return 0
}

func runInteractiveUpdate(files []string, basePath string) {
	for _, f := range files {
		r, err := internal.InteractiveUpdate(f, basePath, os.Stdin, os.Stdout)
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
