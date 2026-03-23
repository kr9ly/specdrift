# specdrift

Go CLI tool to detect drift between spec documents and source code via embedded hash annotations.

## Build & Test

```bash
go build -o specdrift .
go test ./...
```

No external dependencies (stdlib only).

## Documentation

- English is the default: `docs/<name>.md`
- Japanese translation alongside: `docs/<name>.ja.md`
- When creating or updating docs, maintain both versions.

## Versioning

Semantic versioning. The boundary is the specdrift annotation format.

- **Major**: annotation format changes (breaking for existing spec files)
- **Minor**: new features that don't change the annotation format
- **Patch**: bug fixes, documentation-only changes
