<!-- specdrift v1 -->

# CLI Commands

<!-- source: main.go@080c5de0 -->

## Subcommand Dispatch

Dispatches to `init`, `check`, or `update` based on the first argument.
Prints usage and exits with code 1 when no arguments or unknown command.

### Argument Parsing

- `--base <dir>` — override base directory for source path resolution
- Remaining arguments are file paths or glob patterns

### Glob Expansion

Arguments containing `*`, `?`, or `[` are expanded by the tool.
Supports `**` for recursive directory matching.

<!-- source: internal/glob.go@fbaa886a -->

Glob expansion is handled internally, not delegated to the shell.
`**` is implemented via `filepath.WalkDir` + `filepath.Match`.

<!-- /source -->

<!-- /source -->

## check

<!-- source: internal/checker.go@913de33e -->

Parses each spec file and compares source annotation hashes against actual file hashes.

### Circular Reference Detection

<!-- source: internal/cycle.go@ff0533b4 -->

Before checking individual files, builds a dependency graph from `.md`-to-`.md` references and detects cycles.
Files participating in a cycle are reported as errors and skipped from normal checking.
Non-participating files are checked normally.

The graph is built transitively: if A references B.md, B.md is also parsed for its `.md` references, even if B.md was not explicitly passed as an argument.

<!-- /source -->

### Check Statuses

- `OK` — hash matches
- `DRIFT` — hash mismatch
- `MISSING` — source file not found
- `TODO` — unresolved placeholder

Exits with code 1 if any file has DRIFT, MISSING, TODO, empty declaration, parse error, or circular reference.

<!-- /source -->

### Report Output

<!-- source: internal/reporter.go@42a8b733 -->

Prints a tree-structured report for each spec file.
Summary line shows counts by status.

Skipped files (no declaration) and empty declarations are reported as single-line messages.

<!-- /source -->

## update

<!-- source: internal/updater.go@a22df0ea -->

Rewrites source annotation hashes in-place to match current file contents.

- `path@hash` — recalculates hash, writes file only if changed
- `path@TODO` — resolves hash if file exists, keeps TODO if not
- bare `TODO` — skipped (no path to resolve)
- Missing files — keeps existing hash unchanged

<!-- /source -->

## init

<!-- source: internal/config.go@06b50e71 -->

Creates a `.specdrift` marker file (`{}`) in the current directory.
Errors if the file already exists.

### Project Root Discovery

`check` and `update` walk up from the current directory to find `.specdrift`.
The directory containing it becomes the default base path.
`--base` flag overrides this.

<!-- /source -->
