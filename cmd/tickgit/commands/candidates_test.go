package commands

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestOpenCandidatesInputUsesStdinByDefault(t *testing.T) {
	reader, closeReader, err := openCandidatesInput("")
	if err != nil {
		t.Fatal(err)
	}
	if reader != os.Stdin {
		t.Fatalf("expected stdin reader, got %#v", reader)
	}
	if err := closeReader(); err != nil {
		t.Fatal(err)
	}
}

func TestOpenCandidatesInputReadsFile(t *testing.T) {
	path := t.TempDir() + "/tickgit.csv"
	if err := os.WriteFile(path, []byte("text,file_path,start_line\nTODO,pkg/app.go,1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	reader, closeReader, err := openCandidatesInput(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := closeReader(); err != nil {
			t.Fatal(err)
		}
	}()

	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "TODO") {
		t.Fatalf("expected file content, got %q", string(content))
	}
}
