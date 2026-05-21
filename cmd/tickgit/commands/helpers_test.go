package commands

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MTG-Thomas/tickgit/pkg/comments"
	"github.com/MTG-Thomas/tickgit/pkg/todos"
	"github.com/augmentable-dev/lege"
)

func TestResolveSearchDirAcceptsCurrentDirectory(t *testing.T) {
	cwd := t.TempDir()

	dir, err := resolveSearchDir(cwd, []string{"."})
	if err != nil {
		t.Fatal(err)
	}

	if dir != "." {
		t.Fatalf("expected current directory to resolve to '.', got %q", dir)
	}
}

func TestResolveSearchDirMakesAbsoluteArgRelativeToCwd(t *testing.T) {
	cwd := t.TempDir()
	arg := filepath.Join(cwd, "pkg")

	dir, err := resolveSearchDir(cwd, []string{arg})
	if err != nil {
		t.Fatal(err)
	}

	if dir != "pkg" {
		t.Fatalf("expected pkg, got %q", dir)
	}
}

func TestWriteCSVNormalizesFilePathSeparators(t *testing.T) {
	collection := lege.NewCollection(
		lege.Location{Line: 3, Pos: 1},
		lege.Location{Line: 3, Pos: 10},
		lege.Boundary{},
		"TODO normalize path",
	)
	found := todos.ToDos{
		{
			Comment: comments.Comment{
				Collection: *collection,
				FilePath:   "pkg\\file.go",
			},
			String: "TODO normalize path",
		},
	}

	var buf bytes.Buffer
	if err := writeCSV(&buf, found); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(buf.String(), "pkg\\file.go") {
		t.Fatalf("expected slash-normalized path, got CSV:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "pkg/file.go") {
		t.Fatalf("expected pkg/file.go in CSV, got:\n%s", buf.String())
	}
}

func TestSelectedIgnorePatternsIsEmptyWithoutConfiguredPatterns(t *testing.T) {
	ignorePaths = nil

	patterns := selectedIgnorePatterns()

	if len(patterns) != 0 {
		t.Fatalf("expected no configured ignore patterns, got %v", patterns)
	}
}

func TestSelectedIgnorePatternsAddsConfiguredPatterns(t *testing.T) {
	ignorePaths = []string{"fixtures", "docs/generated"}
	t.Cleanup(func() {
		ignorePaths = nil
	})

	patterns := selectedIgnorePatterns()

	if !containsString(patterns, "fixtures") || !containsString(patterns, "docs/generated") {
		t.Fatalf("expected configured ignore patterns, got %v", patterns)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
