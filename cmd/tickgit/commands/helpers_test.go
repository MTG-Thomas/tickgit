package commands

import (
	"path/filepath"
	"testing"
)

func TestResolveSearchDirAcceptsCurrentDirectory(t *testing.T) {
	cwd := filepath.Clean("C:/work/tickgit")

	dir, err := resolveSearchDir(cwd, []string{"."})
	if err != nil {
		t.Fatal(err)
	}

	if dir != "." {
		t.Fatalf("expected current directory to resolve to '.', got %q", dir)
	}
}

func TestResolveSearchDirMakesAbsoluteArgRelativeToCwd(t *testing.T) {
	cwd := filepath.Clean("C:/work/tickgit")
	arg := filepath.Join(cwd, "pkg")

	dir, err := resolveSearchDir(cwd, []string{arg})
	if err != nil {
		t.Fatal(err)
	}

	if dir != "pkg" {
		t.Fatalf("expected pkg, got %q", dir)
	}
}
