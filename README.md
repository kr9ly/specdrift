# specdrift

Detect drift between spec documents and source code.

Embed source file references as hash annotations in Markdown specs. When the source changes, `specdrift check` catches it.

## Install

```bash
go install github.com/kr9ly/specdrift@latest
```

## Quick Start

Initialize at your project root:

```bash
specdrift init
```

Add a declaration and source annotations to your spec file:

```markdown
<!-- specdrift v1 -->

<!-- source: path/to/handler.go@TODO -->
Handler spec goes here.
<!-- /source -->
```

Resolve hashes and check for drift across the project:

```bash
specdrift update '**/*.md'
specdrift check '**/*.md'
```

You can also target specific files:

```bash
specdrift check docs/spec.md
```

## Annotation Format

See [docs/format/annotation-format.md](docs/format/annotation-format.md) for the full format reference.

Japanese version: [docs/format/annotation-format.ja.md](docs/format/annotation-format.ja.md)

## Agent Integration Guide

See [docs/guide/agent-setup.md](docs/guide/agent-setup.md) for how to set up specdrift with AI coding agents.

Japanese version: [docs/guide/agent-setup.ja.md](docs/guide/agent-setup.ja.md)
