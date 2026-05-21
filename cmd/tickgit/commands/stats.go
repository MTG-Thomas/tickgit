package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/MTG-Thomas/tickgit/pkg/stats"
	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
)

var statsJSON bool

func init() {
	statsCmd.Flags().BoolVar(&statsJSON, "json", false, "output stats as JSON")
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Print historical stats for current TODOs",
	Long:  `Scans a given git repository for TODOs and prints current-state historical stats using git blame metadata.`,
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

		foundToDos, err := findToDos(context.Background(), dir, s)
		handleError(err, s)
		s.Stop()

		report := stats.Build(foundToDos, time.Now())
		if statsJSON {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			encoder.SetEscapeHTML(false)
			handleError(encoder.Encode(report), nil)
			return
		}

		handleError(writeStats(report), nil)
	},
}

func writeStats(report stats.Report) error {
	if _, err := fmt.Fprintf(os.Stdout, "%d TODOs Found\n\n", report.Total); err != nil {
		return err
	}
	if err := writeCounts("By phrase", report.ByPhrase); err != nil {
		return err
	}
	if err := writeCounts("By age", report.ByAgeBucket); err != nil {
		return err
	}
	if err := writeCounts("By directory", report.ByDirectory); err != nil {
		return err
	}

	if len(report.Oldest) == 0 {
		return nil
	}

	if _, err := fmt.Fprintln(os.Stdout, "Oldest findings:"); err != nil {
		return err
	}
	for _, finding := range report.Oldest {
		if _, err := fmt.Fprintf(
			os.Stdout,
			"  %s:%d %s by %s in %s (%s)\n",
			finding.FilePath,
			finding.Line,
			finding.Text,
			finding.Author,
			finding.IntroducedSHA,
			finding.AgeBucket,
		); err != nil {
			return err
		}
	}

	return nil
}

func writeCounts(title string, counts map[string]int) error {
	if _, err := fmt.Fprintln(os.Stdout, title+":"); err != nil {
		return err
	}
	for _, key := range sortedKeys(counts) {
		if _, err := fmt.Fprintf(os.Stdout, "  %s: %d\n", key, counts[key]); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(os.Stdout); err != nil {
		return err
	}
	return nil
}

func sortedKeys(counts map[string]int) []string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
