package internal

import (
	"testing"
)

func TestParseAnnotations_single(t *testing.T) {
	content := `# Spec

<!-- source: src/auth/login.ts@a3f2b1c0 -->

Login spec here.

<!-- /source -->
`
	result, err := ParseAnnotations(content)
	if err != nil {
		t.Fatal(err)
	}
	roots := result.Annotations
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	a := roots[0]
	if len(a.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(a.Sources))
	}
	if a.Sources[0].Path != "src/auth/login.ts" {
		t.Errorf("path = %q, want %q", a.Sources[0].Path, "src/auth/login.ts")
	}
	if a.Sources[0].ExpectedHash != "a3f2b1c0" {
		t.Errorf("hash = %q, want %q", a.Sources[0].ExpectedHash, "a3f2b1c0")
	}
	if a.Line != 3 {
		t.Errorf("line = %d, want 3", a.Line)
	}
}

func TestParseAnnotations_multiple_sources(t *testing.T) {
	content := `<!-- source: src/a.ts@11111111, src/b.ts@22222222 -->
<!-- /source -->
`
	result, err := ParseAnnotations(content)
	if err != nil {
		t.Fatal(err)
	}
	a := result.Annotations[0]
	if len(a.Sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(a.Sources))
	}
	if a.Sources[0].Path != "src/a.ts" {
		t.Errorf("first source path = %q", a.Sources[0].Path)
	}
	if a.Sources[1].Path != "src/b.ts" {
		t.Errorf("second source path = %q", a.Sources[1].Path)
	}
	if a.Sources[1].ExpectedHash != "22222222" {
		t.Errorf("second source hash = %q", a.Sources[1].ExpectedHash)
	}
}

func TestParseAnnotations_nested(t *testing.T) {
	content := `<!-- source: src/auth/login.ts@a3f2b1c0 -->

Outer spec.

<!-- source: src/db/auth/token.ts@b4e5d2f1 -->

Inner spec.

<!-- /source -->

More outer spec.

<!-- /source -->
`
	result, err := ParseAnnotations(content)
	if err != nil {
		t.Fatal(err)
	}
	roots := result.Annotations
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	outer := roots[0]
	if outer.Sources[0].Path != "src/auth/login.ts" {
		t.Errorf("outer path = %q", outer.Sources[0].Path)
	}
	if len(outer.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(outer.Children))
	}
	inner := outer.Children[0]
	if inner.Sources[0].Path != "src/db/auth/token.ts" {
		t.Errorf("inner path = %q", inner.Sources[0].Path)
	}
}

func TestParseAnnotations_multiple_roots(t *testing.T) {
	content := `<!-- source: a.ts@11111111 -->
<!-- /source -->
<!-- source: b.ts@22222222 -->
<!-- /source -->
`
	result, err := ParseAnnotations(content)
	if err != nil {
		t.Fatal(err)
	}
	roots := result.Annotations
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(roots))
	}
	if roots[0].Sources[0].Path != "a.ts" {
		t.Errorf("first root path = %q", roots[0].Sources[0].Path)
	}
	if roots[1].Sources[0].Path != "b.ts" {
		t.Errorf("second root path = %q", roots[1].Sources[0].Path)
	}
}

func TestParseAnnotations_unclosed_error(t *testing.T) {
	content := `<!-- source: a.ts@11111111 -->
some text
`
	_, err := ParseAnnotations(content)
	if err == nil {
		t.Fatal("expected error for unclosed tag")
	}
}

func TestParseAnnotations_unmatched_close_error(t *testing.T) {
	content := `<!-- /source -->
`
	_, err := ParseAnnotations(content)
	if err == nil {
		t.Fatal("expected error for unmatched close tag")
	}
}

func TestParseAnnotations_deep_nesting(t *testing.T) {
	content := `<!-- source: a.ts@11111111 -->
<!-- source: b.ts@22222222 -->
<!-- source: c.ts@33333333 -->
<!-- /source -->
<!-- /source -->
<!-- /source -->
`
	result, err := ParseAnnotations(content)
	if err != nil {
		t.Fatal(err)
	}
	roots := result.Annotations
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	grandchild := roots[0].Children[0].Children[0]
	if grandchild.Sources[0].Path != "c.ts" {
		t.Errorf("grandchild path = %q", grandchild.Sources[0].Path)
	}
}

func TestParseAnnotations_declared(t *testing.T) {
	content := `<!-- spec-drift -->

<!-- source: a.ts@11111111 -->
<!-- /source -->
`
	result, err := ParseAnnotations(content)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Declared {
		t.Error("expected Declared = true")
	}
	if result.Version != 1 {
		t.Errorf("version = %d, want 1", result.Version)
	}
	if len(result.Annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(result.Annotations))
	}
}

func TestParseAnnotations_declared_with_version(t *testing.T) {
	content := `<!-- spec-drift v1 -->

<!-- source: a.ts@11111111 -->
<!-- /source -->
`
	result, err := ParseAnnotations(content)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Declared {
		t.Error("expected Declared = true")
	}
	if result.Version != 1 {
		t.Errorf("version = %d, want 1", result.Version)
	}
}

func TestParseAnnotations_future_version(t *testing.T) {
	content := `<!-- spec-drift v99 -->

<!-- source: a.ts@11111111 -->
<!-- /source -->
`
	result, err := ParseAnnotations(content)
	if err != nil {
		t.Fatal(err)
	}
	if result.Version != 99 {
		t.Errorf("version = %d, want 99", result.Version)
	}
}

func TestParseAnnotations_not_declared(t *testing.T) {
	content := `<!-- source: a.ts@11111111 -->
<!-- /source -->
`
	result, err := ParseAnnotations(content)
	if err != nil {
		t.Fatal(err)
	}
	if result.Declared {
		t.Error("expected Declared = false")
	}
}

func TestParseAnnotations_declared_no_annotations(t *testing.T) {
	content := `<!-- spec-drift -->

This file has no source annotations.
`
	result, err := ParseAnnotations(content)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Declared {
		t.Error("expected Declared = true")
	}
	if len(result.Annotations) != 0 {
		t.Errorf("expected 0 annotations, got %d", len(result.Annotations))
	}
}
