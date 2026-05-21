package stats

import (
	"testing"
	"time"

	"github.com/MTG-Thomas/tickgit/pkg/blame"
	"github.com/MTG-Thomas/tickgit/pkg/comments"
	"github.com/MTG-Thomas/tickgit/pkg/todos"
	"github.com/augmentable-dev/lege"
)

func TestBuildGroupsFindingsByPhraseAgeAndDirectory(t *testing.T) {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	found := todos.ToDos{
		finding("TODO first", "pkg/one.go", 3, "TODO", now.AddDate(0, 0, -10), "a"),
		finding("FIXME second", "pkg/two.go", 4, "FIXME", now.AddDate(0, 0, -60), "b"),
		finding("TODO old", "cmd/app.go", 5, "TODO", now.AddDate(-2, 0, 0), "c"),
	}

	report := Build(found, now)

	if report.Total != 3 {
		t.Fatalf("expected total 3, got %d", report.Total)
	}
	if report.ByPhrase["TODO"] != 2 || report.ByPhrase["FIXME"] != 1 {
		t.Fatalf("unexpected phrase counts: %#v", report.ByPhrase)
	}
	if report.ByAgeBucket["<30d"] != 1 || report.ByAgeBucket["30-90d"] != 1 || report.ByAgeBucket[">365d"] != 1 {
		t.Fatalf("unexpected age buckets: %#v", report.ByAgeBucket)
	}
	if report.ByDirectory["pkg"] != 2 || report.ByDirectory["cmd"] != 1 {
		t.Fatalf("unexpected directory counts: %#v", report.ByDirectory)
	}
	if len(report.Oldest) != 3 || report.Oldest[0].Text != "TODO old" {
		t.Fatalf("expected oldest finding first, got %#v", report.Oldest)
	}
}

func TestBuildCountsUnknownAgeWhenBlameIsMissing(t *testing.T) {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	found := todos.ToDos{findingWithoutBlame("TODO unknown", "README.md", 1, "TODO")}

	report := Build(found, now)

	if report.ByAgeBucket["unknown"] != 1 {
		t.Fatalf("expected unknown age bucket, got %#v", report.ByAgeBucket)
	}
	if len(report.Oldest) != 0 {
		t.Fatalf("expected no oldest entries without blame, got %#v", report.Oldest)
	}
}

func finding(text string, path string, line int, phrase string, when time.Time, sha string) *todos.ToDo {
	todo := findingWithoutBlame(text, path, line, phrase)
	todo.Blame = &blame.Blame{
		SHA: sha,
		Author: blame.Event{
			Name:  "Test User",
			Email: "test@example.com",
			When:  when,
		},
	}
	return todo
}

func findingWithoutBlame(text string, path string, line int, phrase string) *todos.ToDo {
	collection := lege.NewCollection(
		lege.Location{Line: line, Pos: 1},
		lege.Location{Line: line, Pos: len(text)},
		lege.Boundary{},
		text,
	)
	return &todos.ToDo{
		Comment: comments.Comment{
			Collection: *collection,
			FilePath:   path,
		},
		String: text,
		Phrase: phrase,
	}
}
