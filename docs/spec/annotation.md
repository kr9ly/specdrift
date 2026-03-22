<!-- specdrift v1 -->

# Annotation Parser

<!-- source: internal/annotation.go@34c0086f -->

## Declaration Detection

Detects `<!-- specdrift -->` or `<!-- specdrift vN -->` declaration tag.
Files without a declaration are not processed.
Version number defaults to 1 when omitted.

## Source Tag Parsing

Parses source open/close tag pairs into a tree structure using a stack.

### Reference Formats

- `path@hash` — resolved reference (hash is SHA-256 first 8 chars)
- `path@TODO` — path known, hash unresolved
- `TODO` — bare placeholder, no path

### Multiple References

A single source tag can contain multiple comma-separated references.

### Nesting

Source tags can nest. Inner tags become children of the outer tag.

### Error Cases

- Closing tag without matching open tag
- Unclosed source tag at end of file
- Source tag with no valid references

<!-- /source -->
