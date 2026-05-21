package todos

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MTG-Thomas/tickgit/pkg/blame"
	"github.com/MTG-Thomas/tickgit/pkg/comments"
	"github.com/dustin/go-humanize"
)

// ToDo represents a ToDo item
type ToDo struct {
	comments.Comment
	String  string
	Phrase  string
	Blame   *blame.Blame
	Context []ContextLine
}

// ToDos represents a list of ToDo items
type ToDos []*ToDo

// ContextLine is one source line rendered around a finding.
type ContextLine struct {
	Line int
	Text string
}

// BlameLookupFailure is a non-fatal failure to find git blame for one file.
type BlameLookupFailure struct {
	FilePath string
	Lines    []int
	Err      error
}

func (f BlameLookupFailure) Error() string {
	return fmt.Sprintf("could not determine git blame for %s: %v", f.FilePath, f.Err)
}

func (f BlameLookupFailure) Unwrap() error {
	return f.Err
}

// BlameLookupFailures collects non-fatal blame lookup failures.
type BlameLookupFailures []BlameLookupFailure

func (f BlameLookupFailures) Error() string {
	if len(f) == 1 {
		return f[0].Error()
	}
	return fmt.Sprintf("could not determine git blame for %d files", len(f))
}

// DefaultMatchPhrases are the phrase markers matched when no custom list is supplied.
var DefaultMatchPhrases = []string{"TODO", "FIXME", "OPTIMIZE", "HACK", "XXX", "WTF", "LEGACY"}

// TimeAgo returns a human readable string indicating the time since the todo was added
func (t *ToDo) TimeAgo() string {
	if t.Blame == nil {
		return "<unknown>"
	}
	return humanize.Time(t.Blame.Author.When)
}

// NewToDo produces a pointer to a ToDo from a comment
func NewToDo(comment comments.Comment) *ToDo {
	return NewToDoWithPhrases(comment, DefaultMatchPhrases)
}

// NewToDoWithPhrases produces a pointer to a ToDo from a comment and explicit phrase list.
func NewToDoWithPhrases(comment comments.Comment, startingMatchPhrases []string) *ToDo {
	var matchPhrases []string
	for _, phrase := range startingMatchPhrases {
		phrase = strings.TrimSpace(phrase)
		if phrase == "" {
			continue
		}
		matchPhrases = append(matchPhrases, phrase, "@"+strings.ToLower(phrase))
	}

	for _, phrase := range matchPhrases {
		s := comment.String()
		if strings.Contains(s, phrase) {
			todo := ToDo{
				Comment: comment,
				String:  strings.Trim(s, " "),
				Phrase:  phrase,
			}
			return &todo
		}
	}

	return nil
}

// NewToDos produces a list of ToDos from a list of comments
func NewToDos(comments comments.Comments) ToDos {
	todos := make(ToDos, 0)
	for _, comment := range comments {
		todo := NewToDo(*comment)
		if todo != nil {
			todos = append(todos, todo)
		}
	}
	return todos
}

// Len returns the number of todos
func (t ToDos) Len() int {
	return len(t)
}

// Less compares two todos by their creation time
func (t ToDos) Less(i, j int) bool {
	first := t[i]
	second := t[j]
	if first.Blame == nil || second.Blame == nil {
		return false
	}
	return first.Blame.Author.When.Before(second.Blame.Author.When)
}

// Swap swaps two todos
func (t ToDos) Swap(i, j int) {
	temp := t[i]
	t[i] = t[j]
	t[j] = temp
}

// CountWithCommits returns the number of todos with an associated commit (in which that todo was added)
func (t ToDos) CountWithCommits() (count int) {
	for _, todo := range t {
		if todo.Blame != nil {
			count++
		}
	}
	return count
}

// FindBlame sets the blame information on each todo in a set of todos
func (t *ToDos) FindBlame(ctx context.Context, dir string) error {
	fileMap := make(map[string]ToDos)
	for _, todo := range *t {
		filePath := todo.FilePath
		if _, ok := fileMap[filePath]; !ok {
			fileMap[filePath] = make(ToDos, 0)
		}
		fileMap[filePath] = append(fileMap[filePath], todo)
	}

	var failures BlameLookupFailures
	for filePath, todos := range fileMap {
		lines := make([]int, 0)

		for _, todo := range todos {
			lines = append(lines, todo.StartLocation.Line)
		}
		blames, err := blame.Exec(ctx, filePath, &blame.Options{
			Directory: dir,
			Lines:     lines,
		})
		if err != nil {
			failedLines := append([]int(nil), lines...)
			failures = append(failures, BlameLookupFailure{
				FilePath: filePath,
				Lines:    failedLines,
				Err:      err,
			})
			continue
		}
		for line, blame := range blames {
			for _, todo := range todos {
				if todo.StartLocation.Line == line {
					b := blame
					todo.Blame = &b
				}
			}
		}
	}
	if len(failures) > 0 {
		return failures
	}
	return nil
}

// FindContext sets source context lines around each todo.
func (t ToDos) FindContext(dir string, contextLines int) error {
	if contextLines <= 0 {
		return nil
	}

	fileMap := make(map[string]ToDos)
	for _, todo := range t {
		fileMap[todo.FilePath] = append(fileMap[todo.FilePath], todo)
	}

	for filePath, todos := range fileMap {
		lines, err := readFileLines(filepath.Join(dir, filePath))
		if err != nil {
			return err
		}
		for _, todo := range todos {
			start := todo.StartLocation.Line - contextLines
			if start < 1 {
				start = 1
			}
			end := todo.EndLocation.Line + contextLines
			if end > len(lines) {
				end = len(lines)
			}
			todo.Context = make([]ContextLine, 0, end-start+1)
			for line := start; line <= end; line++ {
				todo.Context = append(todo.Context, ContextLine{
					Line: line,
					Text: lines[line-1],
				})
			}
		}
	}

	return nil
}

func readFileLines(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	rawLines := strings.Split(strings.TrimSuffix(string(content), "\n"), "\n")
	for i := range rawLines {
		rawLines[i] = strings.TrimSuffix(rawLines[i], "\r")
	}

	return rawLines, nil
}
