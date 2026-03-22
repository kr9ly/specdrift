# specdrift Annotation Format (v1)

Embed source code references in Markdown spec files to detect drift.

## Declaration

Add the following near the top of the file. Files without this declaration are ignored by specdrift.

```markdown
<!-- specdrift v1 -->
```

The version number is optional (defaults to v1).

## Source References

Wrap spec sections with `source` tags to associate them with source files.

### Basic Workflow

Start with a TODO placeholder when writing specs before implementation. Resolve hashes later with `update`.

```markdown
<!-- source: TODO -->
Spec for this feature. Implementation TBD.
<!-- /source -->
```

Once the path is known, fill it in. The hash can remain `TODO`.

```markdown
<!-- source: path/to/handler.go@TODO -->
Handler spec.
<!-- /source -->
```

Run `specdrift update` to fill in the hash (first 8 characters of SHA-256).

```markdown
<!-- source: path/to/handler.go@a1b2c3d4 -->
Handler spec.
<!-- /source -->
```

From this point, `specdrift check` detects DRIFT when the source file changes.

### Multiple File References

Specify multiple files in a single tag, separated by commas.

```markdown
<!-- source: handler.go@a1b2c3d4, handler_test.go@e5f6a7b8 -->
```

### Nesting

Source tags can be nested to separate outer and inner scopes.

```markdown
<!-- source: api/router.go@a1b2c3d4 -->
Routing spec.

  <!-- source: api/middleware.go@e5f6a7b8 -->
  Middleware spec.
  <!-- /source -->

<!-- /source -->
```

## Setup

Run `init` at the project root. This creates a `.specdrift` file that serves as the base directory for resolving source paths.

```bash
specdrift init
```

## Commands

```bash
# Check for drift
specdrift check docs/spec.md

# Update hashes to match current files (also resolves path@TODO)
specdrift update docs/spec.md

# Explicitly specify base directory (overrides .specdrift)
specdrift check --base /path/to/repo docs/spec.md
```

## Check Results

| Status  | Meaning                              |
| ------- | ------------------------------------ |
| OK      | Hash matches                         |
| DRIFT   | Hash mismatch (source was modified)  |
| MISSING | Source file does not exist            |
| TODO    | Unimplemented placeholder            |

Exits with code 1 if any DRIFT, MISSING, or TODO is found.
