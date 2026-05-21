package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/briandowns/spinner"
)

const commandErrorExitCode = 1

func handleError(err error, activeSpinner *spinner.Spinner) {
	if code := reportCommandError(err, activeSpinner, os.Stderr); code != 0 {
		os.Exit(code)
	}
}

func reportCommandError(err error, activeSpinner *spinner.Spinner, stderr io.Writer) int {
	if err == nil {
		return 0
	}

	stopSpinner(activeSpinner)
	if _, writeErr := fmt.Fprintln(stderr, err); writeErr != nil {
		return commandErrorExitCode
	}
	return commandErrorExitCode
}

func stopSpinner(activeSpinner *spinner.Spinner) {
	if activeSpinner == nil {
		return
	}

	activeSpinner.FinalMSG = ""
	activeSpinner.Stop()
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if len(os.Args) > 1 && os.Args[1] == "stats" {
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
		if err := statsCmd.Execute(); err != nil {
			handleError(err, nil)
		}
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "candidates" {
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
		if err := candidatesCmd.Execute(); err != nil {
			handleError(err, nil)
		}
		return
	}

	if err := todosCmd.Execute(); err != nil {
		handleError(err, nil)
	}
}
