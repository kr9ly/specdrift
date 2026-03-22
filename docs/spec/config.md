<!-- specdrift v1 -->

# Project Configuration

<!-- source: internal/config.go@06b50e71 -->

## .specdrift File

Marker file at the project root. Created by `specdrift init`.
Contains `{}` (empty JSON object, reserved for future configuration).

## Project Root Discovery

Walks up from the start directory toward the filesystem root.
Returns the first directory containing a `.specdrift` file.
Returns empty string if none found.

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
