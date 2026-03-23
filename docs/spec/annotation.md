<!-- specdrift v1 -->

# Annotation Parser

<!-- source: internal/annotation.go@1f856111 -->

## Declaration Detection

Detects `<!-- specdrift -->` or `<!-- specdrift vN -->` declaration tag.
Files without a declaration are not processed.
Version number defaults to 1 when omitted.

## Source Tag Parsing

Parses source open/close tag pairs into a tree structure using a stack.

### Reference Formats

<!-- source: internal/hasher.go@5315ca07 -->

- `path@hash` — resolved reference (hash is SHA-256 first 8 chars)
- `path@TODO` — path known, hash unresolved
- `TODO` — bare placeholder, no path

<!-- /source -->

### Multiple References

A single source tag can contain multiple comma-separated references.

### Nesting

Source tags can nest. Inner tags become children of the outer tag.

### Code Block Skipping

Fenced code blocks (`` ``` `` or `~~~`) are skipped entirely.
Closing fences must use the same character and be at least as long as the opening fence (CommonMark compliant).

Inline code spans (backtick-delimited) within a line are stripped before matching,
so annotations inside inline code are ignored.

### Document-to-Document References

Source tags can reference other `.md` files, enabling dependency tracking between documents.

**Circular references are not allowed.** If document A references document B and B references A (directly or transitively), `check` reports an error on all participating files. This prevents infinite update loops where updating one document changes its hash and triggers drift in the other.

To resolve a circular reference, remove one direction of the dependency so the relationship becomes one-way.

### Error Cases

- Closing tag without matching open tag
- Unclosed source tag at end of file
- Source tag with no valid references
- Circular references between documents

<!-- /source -->
