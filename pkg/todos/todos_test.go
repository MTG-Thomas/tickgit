package todos

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MTG-Thomas/tickgit/pkg/comments"
	"github.com/augmentable-dev/lege"
)

func TestNewToDoNil(t *testing.T) {
	collection := lege.NewCollection(lege.Location{}, lege.Location{}, lege.Boundary{}, "Hello World")
	comment := comments.Comment{
		Collection: *collection,
	}
	todo := NewToDo(comment)

	if todo != nil {
		t.Fatalf("did not expect a TODO, got: %v", todo)
	}
}

func TestNewToDo(t *testing.T) {
	collection := lege.NewCollection(lege.Location{}, lege.Location{}, lege.Boundary{}, "TODO Hello World")
	comment := comments.Comment{
		Collection: *collection,
	}
	todo := NewToDo(comment)

	if todo == nil {
		t.Fatalf("expected a TODO, got: %v", todo)
	}

	if todo.Phrase != "TODO" {
		t.Fatalf("expected matched phrase to be TODO, got: %s", todo.Phrase)
	}
}

func TestNewToDoWithCustomMatchPhrases(t *testing.T) {
	collection := lege.NewCollection(lege.Location{}, lege.Location{}, lege.Boundary{}, "NOTE Hello World")
	comment := comments.Comment{
		Collection: *collection,
	}

	todo := NewToDoWithPhrases(comment, []string{"NOTE"})
	if todo == nil {
		t.Fatalf("expected a TODO from custom NOTE phrase")
	}
	if todo.Phrase != "NOTE" {
		t.Fatalf("expected matched phrase to be NOTE, got: %s", todo.Phrase)
	}

	defaultTodo := NewToDo(comment)
	if defaultTodo != nil {
		t.Fatalf("did not expect default phrases to match NOTE, got: %v", defaultTodo)
	}
}

func TestNewToDoWithCustomMatchPhrasesAddsAtLowercaseVariant(t *testing.T) {
	collection := lege.NewCollection(lege.Location{}, lege.Location{}, lege.Boundary{}, "@note Hello World")
	comment := comments.Comment{
		Collection: *collection,
	}

	todo := NewToDoWithPhrases(comment, []string{"NOTE"})
	if todo == nil {
		t.Fatalf("expected @note to match custom NOTE phrase")
	}
	if todo.Phrase != "@note" {
		t.Fatalf("expected matched phrase to be @note, got: %s", todo.Phrase)
	}
}

func TestFindContextAddsSurroundingLines(t *testing.T) {
	dir := t.TempDir()
	phrase := "TO" + "DO"
	todoText := phrase + " wire context"
	sourceLine := "// " + todoText
	err := os.WriteFile(filepath.Join(dir, "sample.go"), []byte("package main\n\n"+sourceLine+"\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

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
		},
	}

	if err := found.FindContext(dir, 1); err != nil {
		t.Fatal(err)
	}

	want := []ContextLine{
		{Line: 2, Text: ""},
		{Line: 3, Text: sourceLine},
		{Line: 4, Text: "func main() {}"},
	}
	if len(found[0].Context) != len(want) {
		t.Fatalf("expected %d context lines, got %d", len(want), len(found[0].Context))
	}
	for i := range want {
		if found[0].Context[i] != want[i] {
			t.Fatalf("context line %d: expected %#v, got %#v", i, want[i], found[0].Context[i])
		}
	}
}

func TestFindContextClampsToFileBoundaries(t *testing.T) {
	dir := t.TempDir()
	phrase := "TO" + "DO"
	firstText := phrase + " first"
	lastText := phrase + " last"
	err := os.WriteFile(filepath.Join(dir, "sample.go"), []byte("// "+firstText+"\npackage main\nfunc main() {}\n// "+lastText+"\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	firstCollection := lege.NewCollection(
		lege.Location{Line: 1, Pos: 3},
		lege.Location{Line: 1, Pos: 15},
		lege.Boundary{},
		firstText,
	)
	lastCollection := lege.NewCollection(
		lege.Location{Line: 4, Pos: 3},
		lege.Location{Line: 4, Pos: 14},
		lege.Boundary{},
		lastText,
	)
	found := ToDos{
		{
			Comment: comments.Comment{
				Collection: *firstCollection,
				FilePath:   "sample.go",
			},
			String: firstText,
			Phrase: phrase,
		},
		{
			Comment: comments.Comment{
				Collection: *lastCollection,
				FilePath:   "sample.go",
			},
			String: lastText,
			Phrase: phrase,
		},
	}

	if err := found.FindContext(dir, 2); err != nil {
		t.Fatal(err)
	}

	if got, want := found[0].Context[0].Line, 1; got != want {
		t.Fatalf("first context should start at line %d, got %d", want, got)
	}
	if got, want := found[1].Context[len(found[1].Context)-1].Line, 4; got != want {
		t.Fatalf("last context should end at line %d, got %d", want, got)
	}
}
