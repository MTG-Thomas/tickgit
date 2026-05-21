package comments

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJSFiles(t *testing.T) {
	var comments Comments
	err := SearchDir("testdata/javascript", func(comment *Comment) {
		comments = append(comments, comment)
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) != 3 {
		t.Fatalf("expected 3 JavaScript comments, got %d from files %v", len(comments), commentFilePaths(comments))
	}
}

func TestSearchFileIgnoresVendorPathsWithOSSeparators(t *testing.T) {
	path := filepath.Join("node_modules", "index.js")
	fixture := "// the comments in this file should be ignored\n"

	var comments Comments
	err := SearchFile(path, strings.NewReader(fixture), func(comment *Comment) {
		comments = append(comments, comment)
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) != 0 {
		t.Fatalf("expected vendor path to be ignored, got %d comments", len(comments))
	}
}

func TestSearchDirUsesDefaultIgnorePatterns(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "src", "app.md", "# TODO keep this\n")
	writeTestFile(t, dir, ".git", "HEAD.md", "# TODO ignore git\n")
	writeTestFile(t, dir, "build", "generated.md", "# TODO ignore build\n")
	writeTestFile(t, dir, "node_modules", "pkg", "index.md", "# TODO ignore dependencies\n")

	var comments Comments
	err := SearchDir(dir, func(comment *Comment) {
		comments = append(comments, comment)
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) != 1 {
		t.Fatalf("expected only non-ignored comments, got %d from files %v", len(comments), commentFilePaths(comments))
	}
	if got, want := filepath.ToSlash(comments[0].FilePath), "src/app.md"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestSearchDirUsesConfiguredIgnorePatterns(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "src", "app.md", "# TODO keep this\n")
	writeTestFile(t, dir, "fixtures", "ignored.md", "# TODO ignore fixtures\n")

	var comments Comments
	err := SearchDirWithOptions(dir, SearchOptions{IgnorePatterns: []string{"fixtures"}}, func(comment *Comment) {
		comments = append(comments, comment)
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) != 1 {
		t.Fatalf("expected configured ignore to skip fixtures, got %d comments from files %v", len(comments), commentFilePaths(comments))
	}
	if got, want := filepath.ToSlash(comments[0].FilePath), "src/app.md"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestSearchFileTreatsMarkdownLinesAsComments(t *testing.T) {
	fixture := "# Notes\n\n- TODO capture this work\n\nPlain paragraph\n"
	var comments Comments

	err := SearchFile("README.md", strings.NewReader(fixture), func(comment *Comment) {
		comments = append(comments, comment)
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) != 3 {
		t.Fatalf("expected 3 Markdown content lines, got %d", len(comments))
	}
	if got, want := comments[1].String(), "- TODO capture this work"; got != want {
		t.Fatalf("expected second content line %q, got %q", want, got)
	}
	if got, want := comments[1].StartLocation.Line, 3; got != want {
		t.Fatalf("expected line %d, got %d", want, got)
	}
}

func TestSearchFileFallsBackToCStyleCommentsForUnknownLanguages(t *testing.T) {
	fixture := "plain text\n// TODO capture this unknown file\n/* FIXME capture this block */\n"
	var comments Comments

	err := SearchFile("notes.unknownext", strings.NewReader(fixture), func(comment *Comment) {
		comments = append(comments, comment)
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) != 2 {
		t.Fatalf("expected 2 fallback comments, got %d", len(comments))
	}
	if got, want := comments[0].String(), " TODO capture this unknown file"; got != want {
		t.Fatalf("expected first fallback comment %q, got %q", want, got)
	}
	if got, want := comments[1].String(), " FIXME capture this block "; got != want {
		t.Fatalf("expected second fallback comment %q, got %q", want, got)
	}
}

func commentFilePaths(comments Comments) []string {
	paths := make([]string, 0, len(comments))
	for _, comment := range comments {
		paths = append(paths, comment.FilePath)
	}
	return paths
}

func commentStrings(comments Comments) []string {
	strings := make([]string, 0, len(comments))
	for _, comment := range comments {
		strings = append(strings, comment.String())
	}
	return strings
}

func writeTestFile(t *testing.T, dir string, parts ...string) {
	t.Helper()
	content := parts[len(parts)-1]
	path := filepath.Join(append([]string{dir}, parts[:len(parts)-1]...)...)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLispFiles(t *testing.T) {
	var comments Comments
	err := SearchDir("testdata/lisp", func(comment *Comment) {
		comments = append(comments, comment)
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) != 1 {
		t.Fail()
	}
}

func TestRustFiles(t *testing.T) {
	var comments Comments
	err := SearchDir("testdata/rust", func(comment *Comment) {
		comments = append(comments, comment)
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) != 21 {
		t.Fatalf("expected 21 Rust comments, got %d", len(comments))
	}
}

func TestSearchFileDocumentsRustDocCommentPrefixBoundaryLimitations(t *testing.T) {
	fixture := "/// module docs\n//! crate docs\n// plain line\n"
	var comments Comments

	err := SearchFile("lib.rs", strings.NewReader(fixture), func(comment *Comment) {
		comments = append(comments, comment)
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) != 2 {
		t.Fatalf("expected 2 Rust comments, got %d: %q", len(comments), commentStrings(comments))
	}
	expected := []string{" module docs", "! crate docs"}
	for i, want := range expected {
		if got := comments[i].String(); got != want {
			t.Fatalf("expected Rust comment %d to be %q, got %q", i, want, got)
		}
	}
}

func TestPHPFiles(t *testing.T) {
	var comments Comments
	err := SearchDir("testdata/php", func(comment *Comment) {
		comments = append(comments, comment)
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) != 3 {
		t.Fail()
	}
}

func TestKotlinFiles(t *testing.T) {
	var comments Comments
	err := SearchDir("testdata/kotlin", func(comment *Comment) {
		comments = append(comments, comment)
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) != 2 {
		t.Fail()
	}
}

func TestJuliaFiles(t *testing.T) {
	var comments Comments
	err := SearchDir("testdata/julia", func(comment *Comment) {
		comments = append(comments, comment)
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) != 3 {
		t.Fail()
	}
}

func TestSearchFileDocumentsJuliaBlockCommentBoundaryLimitations(t *testing.T) {
	fixture := "#= block line\nsecond line =#\n# line comment\n"
	var comments Comments

	err := SearchFile("script.jl", strings.NewReader(fixture), func(comment *Comment) {
		comments = append(comments, comment)
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) != 1 {
		t.Fatalf("expected 1 Julia comment, got %d: %q", len(comments), commentStrings(comments))
	}
	if got, want := comments[0].String(), " block line\nsecond line "; got != want {
		t.Fatalf("expected Julia block comment %q, got %q", want, got)
	}
}

func TestElixirFiles(t *testing.T) {
	var comments Comments
	err := SearchDir("testdata/elixir", func(comment *Comment) {
		comments = append(comments, comment)
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) != 2 {
		t.Fail()
	}
}

func TestHaskellFiles(t *testing.T) {
	var comments Comments
	err := SearchDir("testdata/haskell", func(comment *Comment) {
		comments = append(comments, comment)
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) != 2 {
		t.Fail()
	}
}
