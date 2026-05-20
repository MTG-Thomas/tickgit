package comments

import (
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

func commentFilePaths(comments Comments) []string {
	paths := make([]string, 0, len(comments))
	for _, comment := range comments {
		paths = append(paths, comment.FilePath)
	}
	return paths
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

	// Track comment-type test organization and parser limits in
	// https://github.com/MTG-Thomas/tickgit/issues/9.
	if len(comments) != 21 {
		t.Fail()
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
