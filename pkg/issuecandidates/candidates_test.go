package issuecandidates

import (
	"bytes"
	"strings"
	"testing"
)

func TestReadCSVDeduplicatesCandidatesByStableKey(t *testing.T) {
	input := strings.NewReader("text,file_path,start_line,start_position,end_line,end_position,author,author_email,author_sha,author_time\nTODO fix this,pkg/app.go,10,1,10,14,A,a@example.com,abc,2026-01-01T00:00:00Z\nTODO   fix this,pkg/app.go,10,1,10,17,A,a@example.com,abc,2026-01-01T00:00:00Z\nFIXME other,cmd/main.go,2,1,2,12,B,b@example.com,def,2026-01-02T00:00:00Z\n")

	candidates, err := ReadCSV(input, "MTG-Thomas/tickgit")
	if err != nil {
		t.Fatal(err)
	}

	if len(candidates) != 2 {
		t.Fatalf("expected 2 deduplicated candidates, got %d", len(candidates))
	}
	if candidates[0].FilePath != "cmd/main.go" || candidates[1].FilePath != "pkg/app.go" {
		t.Fatalf("expected candidates sorted by path, got %#v", candidates)
	}
}

func TestReadCSVRequiresTickgitColumns(t *testing.T) {
	_, err := ReadCSV(strings.NewReader("text,file_path\nTODO,pkg/app.go\n"), "repo")
	if err == nil {
		t.Fatal("expected missing column error")
	}
	if !strings.Contains(err.Error(), "start_line") {
		t.Fatalf("expected start_line error, got %v", err)
	}
}

func TestWriteMarkdownIncludesEvidenceAndDuplicateKeys(t *testing.T) {
	candidates := []Candidate{{
		Repository: "MTG-Thomas/tickgit",
		Text:       "TODO make this tracked",
		FilePath:   "pkg/app.go",
		StartLine:  "10",
		Author:     "A",
		AuthorTime: "2026-01-01T00:00:00Z",
	}}
	candidates[0].Key = StableKey(candidates[0])

	var output bytes.Buffer
	if err := WriteMarkdown(&output, candidates); err != nil {
		t.Fatal(err)
	}

	got := output.String()
	for _, want := range []string{
		"# Tickgit issue candidates",
		"duplicate_key",
		"MTG-Thomas/tickgit",
		"pkg/app.go:10",
		"> TODO make this tracked",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected markdown to contain %q, got:\n%s", want, got)
		}
	}
}
