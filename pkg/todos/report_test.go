package todos

import (
	"bytes"
	"strings"
	"testing"

	"github.com/MTG-Thomas/tickgit/pkg/comments"
	"github.com/augmentable-dev/lege"
)

func TestWriteTodosIncludesContextLines(t *testing.T) {
	phrase := "TO" + "DO"
	todoText := phrase + " wire context"
	sourceLine := "// " + todoText
	collection := lege.NewCollection(
		lege.Location{Line: 3, Pos: 3},
		lege.Location{Line: 3, Pos: 22},
		lege.Boundary{},
		todoText,
	)
	found := ToDos{
		{
			Comment: comments.Comment{
				Collection: *collection,
				FilePath:   "sample.go",
			},
			String: todoText,
			Phrase: phrase,
			Context: []ContextLine{
				{Line: 2, Text: ""},
				{Line: 3, Text: sourceLine},
				{Line: 4, Text: "func main() {}"},
			},
		},
	}

	var buf bytes.Buffer
	if err := WriteTodos(found, &buf); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	for _, want := range []string{
		"  => context:",
		"     2 | ",
		"     3 | " + sourceLine,
		"     4 | func main() {}",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, output)
		}
	}
}
