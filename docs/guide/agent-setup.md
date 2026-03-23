# Setting Up specdrift for AI Coding Agents

This guide covers installing specdrift and integrating it into your development workflow with an AI coding agent.

If your project already has documentation (design docs, rule definitions, guides, etc.), you don't need to write new spec files. You can add specdrift annotations directly to your existing documents.

## Part 1: Tool Setup

### Install

```bash
go install github.com/kr9ly/specdrift@latest
```

### Initialize

Run at the root of your project. This creates a `.specdrift` marker file used to resolve source paths.

```bash
cd /path/to/your/project
specdrift init
```

### Add Annotations to Existing Documents

If your project already has documentation, this is the quickest way to get started.

1. Add the specdrift declaration at the top of the document:

```markdown
<!-- specdrift v1 -->
```

2. Wrap sections that describe specific source files with source annotations:

```markdown
<!-- source: src/auth/handler.go@TODO -->
This handler validates credentials and returns a JWT token.
<!-- /source -->
```

> **Important:** The `@TODO` or `@hash` suffix is required. Writing just `src/auth/handler.go` without the `@` suffix will produce an error.

3. Run `specdrift update` to resolve hashes for existing files:

```bash
specdrift update docs/design/auth.md
```

4. Verify with `specdrift check`:

```bash
specdrift check docs/design/auth.md
```

### Write a New Spec (Spec-First Workflow)

Alternatively, create a new spec file (e.g., `docs/spec/auth.md`) and add the specdrift declaration:

```markdown
<!-- specdrift v1 -->

# Authentication

<!-- source: TODO -->
Users authenticate via email and password.
The handler validates input, checks credentials, and returns a token.
<!-- /source -->
```

This is the **spec-first workflow**: write the spec before the implementation exists. The `TODO` placeholder signals that the source file hasn't been created yet.

### Evolve the Spec as Code Takes Shape

Once you know the file path:

```markdown
<!-- source: src/auth/handler.go@TODO -->
```

Once the file exists, run `update` to resolve the hash:

```bash
specdrift update docs/spec/auth.md
```

The annotation becomes:

```markdown
<!-- source: src/auth/handler.go@a1b2c3d4 -->
```

From this point, `specdrift check` detects when the source file changes.

### Verify

```bash
specdrift check docs/spec/auth.md
```

## Part 2: Integrating into Your Development Process

### The Core Principle

specdrift enforces one rule: **when source code changes, the spec must be reviewed.** A hash mismatch (DRIFT) doesn't necessarily mean the spec is wrong — it means someone needs to decide whether the spec still accurately describes the code.

For AI coding agents, this is particularly valuable. Agents can modify code freely, but specdrift forces a checkpoint: the spec must be explicitly reconciled before a commit goes through.

### Add specdrift to Your Commit Checks

Add `specdrift check` to your pre-commit workflow. The exact mechanism depends on your agent and project, but the pattern is the same:

**Run checks in order. Stop at the first failure.**

1. Formatter / linter (e.g., `gofmt`, `eslint`, `ruff`)
2. Static analysis (e.g., `go vet`, `tsc --noEmit`)
3. Tests
4. `specdrift check 'docs/spec/*.md'`
5. Commit

The key point: specdrift check runs **after tests pass** but **before the commit is created**. This ensures the agent addresses spec drift while the changes are fresh.

Choose the glob pattern that fits your project layout:

- `'docs/spec/*.md'` — when specs live in a dedicated directory
- `'docs/**/*.md'` — when annotations are spread across existing documents (files without the `<!-- specdrift v1 -->` declaration are silently skipped)

#### Example: Claude Code custom command

Create `.claude/commands/commit.md` in your project:

```markdown
Run checks and commit. Execute in order, stop at first failure.

1. **format**: run your formatter
2. **lint/vet**: run static analysis
3. **test**: run your test suite
4. **specdrift check**: `specdrift check 'docs/spec/*.md'`
   - If DRIFT is detected, do NOT just run update. First read the drifted spec
     and the changed source, then decide whether the spec text needs revision.
     Update the spec if needed, then run `specdrift update` to sync hashes.
5. **commit**: stage and commit with a descriptive message
```

#### Example: Git pre-commit hook

```bash
#!/bin/sh
specdrift check 'docs/spec/*.md'
```

### Handling DRIFT: The Review Rule

When `specdrift check` reports DRIFT, the correct response is:

1. **Read the drifted spec section** — understand what the spec says
2. **Read the changed source code** — understand what actually changed
3. **Decide**: does the spec still match the code's behavior?
   - If yes → run `specdrift update` to sync the hash
   - If no → revise the spec text, then run `specdrift update`

**Never run `specdrift update` without reviewing the spec first.** Silent updates defeat the entire purpose. This is the single most important rule to convey to your agent.

For Claude Code, you can enforce this in CLAUDE.md:

```markdown
When specdrift check detects DRIFT, read the spec and the changed source before updating.
Do not run `specdrift update` without reviewing whether the spec text needs revision.
```

### Documenting specdrift in Your Project Instructions

Add specdrift to whatever file your agent reads for project context (e.g., `CLAUDE.md`, `AGENTS.md`, `CONVENTIONS.md`). Include:

1. **How to build and check**: the commands to run
2. **The review rule**: never silent-update
3. **Spec file locations**: where specs live (e.g., `docs/spec/`)

Example block for CLAUDE.md:

```markdown
## Spec Drift Detection

This project uses specdrift to track whether spec documents are in sync with source code.

- Check: `specdrift check 'docs/spec/*.md'`
- Update (after review): `specdrift update 'docs/spec/*.md'`
- When DRIFT is detected, read both the spec and the changed source before updating.
```

### The Spec-First Workflow

specdrift supports writing specs before code exists, using TODO placeholders:

1. **Write the spec** with `<!-- source: TODO -->` — describe what the code should do
2. **Implement the code** — the agent writes the source file
3. **Fill in the path** — change to `<!-- source: path/to/file.go@TODO -->`
4. **Run update** — `specdrift update` resolves the hash

This workflow pairs naturally with AI coding agents. The spec serves as a structured prompt: it describes the desired behavior, and the agent implements it. The hash annotation then locks the spec to the implementation, catching future drift.

### Keeping Specs Maintainable

A few practical tips:

- **One spec file per module or feature** — avoid monolithic spec documents
- **Scope annotations narrowly** — annotate the specific section that discusses a file, not the entire document
- **Use nesting for hierarchical specs** — outer scope for the module, inner scopes for individual files
- **Don't annotate every file** — focus on files where the intent matters (handlers, core logic), not boilerplate
- **Multiple documents can reference the same source** — this is normal and expected (e.g., a design doc and a rule doc may both track the same file)

#### Choosing What to Annotate

- Annotate when a document describes **how to use** a specific source file
- Don't annotate when a document explains **general concepts** without depending on specific implementations
- Don't annotate indirectly related files — it creates noise and false drift signals
