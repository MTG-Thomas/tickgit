package commands

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/MTG-Thomas/tickgit/pkg/baseline"
	"github.com/MTG-Thomas/tickgit/pkg/comments"
	"github.com/MTG-Thomas/tickgit/pkg/todos"
	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
)

var csvOutput bool
var baselineFile string
var failOnNew bool
var contextLines int
var matchPhrases []string

func init() {
	todosCmd.Flags().BoolVar(&csvOutput, "csv-output", false, "specify whether or not output should be in CSV format")
	todosCmd.Flags().StringVar(&baselineFile, "baseline-file", "", "compare CSV output against a tickgit baseline file")
	todosCmd.Flags().BoolVar(&failOnNew, "fail-on-new", false, "exit with status 2 when baseline comparison finds new TODOs")
	todosCmd.Flags().IntVar(&contextLines, "context-lines", 0, "number of source lines to show before and after each TODO in human-readable output")
	todosCmd.Flags().StringSliceVar(&matchPhrases, "match-phrase", nil, "phrase to match as latent work; repeat or comma-separate to override defaults")
	statsCmd.Flags().StringSliceVar(&matchPhrases, "match-phrase", nil, "phrase to match as latent work; repeat or comma-separate to override defaults")
}

var todosCmd = &cobra.Command{
	Use:   "todos",
	Short: "Print a report of current TODOs",
	Long:  `Scans a given git repository looking for any code comments with TODOs. Displays a report of all the TODO items found.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
		s.HideCursor = true
		s.Suffix = " finding TODOs"
		s.Writer = os.Stderr
		s.Start()

		cwd, err := os.Getwd()
		handleError(err, s)

		dir, err := resolveSearchDir(cwd, args)
		handleError(err, s)

		validateDir(dir)

		if !csvOutput {
			s.Suffix = " finding TODOs"
		}
		foundToDos, err := findToDos(context.Background(), dir, s)
		handleError(err, s)

		if !csvOutput {
			err = foundToDos.FindContext(dir, contextLines)
			handleError(err, s)
		}

		s.Stop()

		if csvOutput {
			var buf bytes.Buffer
			err := writeCSV(&buf, foundToDos)
			handleError(err, s)

			_, err = os.Stdout.Write(buf.Bytes())
			handleError(err, s)

			handleBaselineComparison(baselineFile, failOnNew, buf.Bytes())
		} else {
			err := todos.WriteTodos(foundToDos, os.Stdout)
			handleError(err, s)
		}

	},
}

func findToDos(ctx context.Context, dir string, s *spinner.Spinner) (todos.ToDos, error) {
	foundToDos := make(todos.ToDos, 0)
	phrases := selectedMatchPhrases()
	err := comments.SearchDir(dir, func(comment *comments.Comment) {
		todo := todos.NewToDoWithPhrases(*comment, phrases)
		if todo != nil {
			foundToDos = append(foundToDos, todo)
			s.Suffix = fmt.Sprintf(" %d TODOs found", len(foundToDos))
		}
	})
	if err != nil {
		return nil, err
	}

	s.Suffix = fmt.Sprintf(" blaming %d TODOs", len(foundToDos))
	// timeout after 30 seconds
	// ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	// defer cancel()
	err = foundToDos.FindBlame(ctx, dir)
	sort.Sort(&foundToDos)
	if err != nil {
		return nil, err
	}

	return foundToDos, nil
}

func selectedMatchPhrases() []string {
	if len(matchPhrases) == 0 {
		return todos.DefaultMatchPhrases
	}
	return matchPhrases
}

func writeCSV(w io.Writer, foundToDos todos.ToDos) error {
	csvWriter := csv.NewWriter(w)
	err := csvWriter.Write([]string{
		"text", "file_path", "start_line", "start_position", "end_line", "end_position", "author", "author_email", "author_sha", "author_time",
	})
	if err != nil {
		return err
	}

	for _, todo := range foundToDos {
		record := []string{
			todo.String,
			normalizeCSVPath(todo.FilePath),
			strconv.Itoa(todo.StartLocation.Line),
			strconv.Itoa(todo.StartLocation.Pos),
			strconv.Itoa(todo.EndLocation.Line),
			strconv.Itoa(todo.EndLocation.Pos),
			"", "", "", "",
		}
		if todo.Blame != nil {
			record[6] = todo.Blame.Author.Name
			record[7] = todo.Blame.Author.Email
			record[8] = todo.Blame.SHA
			record[9] = todo.Blame.Author.When.Format(time.RFC3339)
		}
		err := csvWriter.Write(record)
		if err != nil {
			return err
		}
	}

	csvWriter.Flush()
	return csvWriter.Error()
}

func normalizeCSVPath(path string) string {
	return filepath.ToSlash(strings.ReplaceAll(path, "\\", string(filepath.Separator)))
}

func handleBaselineComparison(path string, shouldFail bool, currentCSV []byte) {
	if path == "" {
		return
	}

	baselineCSV, err := os.Open(path)
	handleError(err, nil)
	defer func() {
		handleError(baselineCSV.Close(), nil)
	}()

	result, err := baseline.CompareCSV(baselineCSV, bytes.NewReader(currentCSV))
	handleError(err, nil)

	if len(result.New) == 0 {
		fmt.Fprintln(os.Stderr, "tickgit baseline check: no new TODOs found")
		return
	}

	fmt.Fprintf(os.Stderr, "tickgit baseline check: %d new TODO(s) found\n", len(result.New))
	for _, finding := range result.New {
		fmt.Fprintf(os.Stderr, "%s:%s: %s\n", finding.FilePath, finding.StartLine, finding.Text)
	}

	if shouldFail {
		os.Exit(2)
	}
}
