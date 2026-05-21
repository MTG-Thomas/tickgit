package todos

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/MTG-Thomas/tickgit/pkg/comments"
	"github.com/augmentable-dev/lege"
)

const ansiYellow = "\u001b[33m"

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

func TestWriteTodosColorsPhrasesByDefault(t *testing.T) {
	restoreNoColor(t)
	found := ToDos{reportTodo("TODO color output", "TODO")}

	var buf bytes.Buffer
	if err := WriteTodos(found, &buf); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), ansiYellow+"TODO"+"\u001b[0m") {
		t.Fatalf("expected default output to colorize TODO, got:\n%s", buf.String())
	}
}

func TestWriteTodosRespectsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	found := ToDos{reportTodo("TODO plain output", "TODO")}

	var buf bytes.Buffer
	if err := WriteTodos(found, &buf); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(buf.String(), "\u001b[") {
		t.Fatalf("expected NO_COLOR output without ANSI escapes, got:\n%s", buf.String())
	}
}

func TestWriteTodosColorAlwaysOverridesNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	found := ToDos{reportTodo("TODO forced output", "TODO")}

	var buf bytes.Buffer
	err := WriteTodosWithOptions(found, &buf, ReportOptions{Color: ColorAlways})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), ansiYellow+"TODO"+"\u001b[0m") {
		t.Fatalf("expected color=always to colorize TODO, got:\n%s", buf.String())
	}
}

func TestWriteTodosColorNeverDisablesColor(t *testing.T) {
	restoreNoColor(t)
	found := ToDos{reportTodo("TODO plain output", "TODO")}

	var buf bytes.Buffer
	err := WriteTodosWithOptions(found, &buf, ReportOptions{Color: ColorNever})
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(buf.String(), "\u001b[") {
		t.Fatalf("expected color=never output without ANSI escapes, got:\n%s", buf.String())
	}
}

func TestWriteTodosDoesNotMutateTodoStrings(t *testing.T) {
	restoreNoColor(t)
	found := ToDos{reportTodo("TODO stable output", "TODO")}

	var buf bytes.Buffer
	if err := WriteTodos(found, &buf); err != nil {
		t.Fatal(err)
	}

	if got := found[0].String; got != "TODO stable output" {
		t.Fatalf("expected original todo string to remain unchanged, got %q", got)
	}
}

func TestParseColorMode(t *testing.T) {
	for _, value := range []string{"auto", "always", "never"} {
		if mode, ok := ParseColorMode(value); !ok || string(mode) != value {
			t.Fatalf("expected %q to parse, got mode=%q ok=%v", value, mode, ok)
		}
	}

	if mode, ok := ParseColorMode("sometimes"); ok || mode != "" {
		t.Fatalf("expected invalid mode to fail, got mode=%q ok=%v", mode, ok)
	}
}

func reportTodo(text string, phrase string) *ToDo {
	collection := lege.NewCollection(
		lege.Location{Line: 3, Pos: 3},
		lege.Location{Line: 3, Pos: 3 + len(text)},
		lege.Boundary{},
		text,
	)
	return &ToDo{
		Comment: comments.Comment{
			Collection: *collection,
			FilePath:   "sample.go",
		},
		String: text,
		Phrase: phrase,
	}
}

func restoreNoColor(t *testing.T) {
	t.Helper()

	value, ok := os.LookupEnv("NO_COLOR")
	if err := os.Unsetenv("NO_COLOR"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if ok {
			_ = os.Setenv("NO_COLOR", value)
		} else {
			_ = os.Unsetenv("NO_COLOR")
		}
	})
}
