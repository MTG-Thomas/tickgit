package commands

import (
	"io"
	"os"

	"github.com/MTG-Thomas/tickgit/pkg/issuecandidates"
	"github.com/spf13/cobra"
)

var candidatesCSVFile string
var candidatesRepository string

func init() {
	candidatesCmd.Flags().StringVar(&candidatesCSVFile, "csv-file", "", "tickgit CSV file to read; defaults to stdin")
	candidatesCmd.Flags().StringVar(&candidatesRepository, "repo", "", "repository name to include in issue candidates")
}

var candidatesCmd = &cobra.Command{
	Use:   "candidates",
	Short: "Convert tickgit CSV output to issue-candidate markdown",
	Long:  `Reads tickgit CSV output and writes issue-candidate markdown with stable duplicate keys for curation workflows.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		reader, closeReader, err := openCandidatesInput(candidatesCSVFile)
		handleError(err, nil)
		defer func() {
			handleError(closeReader(), nil)
		}()

		candidates, err := issuecandidates.ReadCSV(reader, candidatesRepository)
		handleError(err, nil)
		handleError(issuecandidates.WriteMarkdown(os.Stdout, candidates), nil)
	},
}

func openCandidatesInput(path string) (io.Reader, func() error, error) {
	if path == "" {
		return os.Stdin, func() error { return nil }, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	return file, file.Close, nil
}
