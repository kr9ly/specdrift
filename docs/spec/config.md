<!-- specdrift v1 -->

# Project Configuration

<!-- source: internal/config.go@68cd5acc -->

## .specdrift File

Marker file and configuration file at the project root. Created by `specdrift init` with `{}`.

### Configuration Fields

- `require_reason` (bool, optional) — require `--reason` for every `update` invocation

### Project Root Discovery

Walks up from the start directory toward the filesystem root.
Returns the first directory containing a `.specdrift` file.
Returns empty string if none found.

### Config Loading

Reads and parses the `.specdrift` JSON file. Returns zero config if the file doesn't exist.

<!-- /source -->

## .specdriftignore File

<!-- source: internal/ignore.go@56954b8e -->

Optional file at the project root (next to `.specdrift`).
Each line is a glob pattern for files to exclude from `coverage` and `graph --reverse`.

### Format

- One pattern per line
- Lines starting with `#` are comments
- Empty lines are ignored
- Patterns use `filepath.Match` syntax

### Matching

Patterns are matched against both the full relative path and the base filename.
For example, `*_test.go` matches both `checker_test.go` and `internal/checker_test.go`.

### Scope

- `coverage` — ignored files are excluded from the source file list before computing coverage
- `graph --reverse` — ignored source files are omitted from the output
- `check`, `update`, `graph` (forward) — not affected

<!-- /source -->

## Glob Expansion

<!-- source: internal/glob.go@fbaa886a -->

Expands file arguments containing glob metacharacters (`*`, `?`, `[`).

### Simple Globs

Delegated to `filepath.Glob` (e.g., `docs/*.md`).

### Recursive Globs

Patterns containing `**` are expanded via `filepath.WalkDir`.
The suffix after `**` is matched against filenames using `filepath.Match`.

<!-- /source -->
