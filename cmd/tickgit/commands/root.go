package commands

import (
	"fmt"
	"os"

	"github.com/briandowns/spinner"
)

// Track CLI error handling cleanup in https://github.com/MTG-Thomas/tickgit/issues/6.
func handleError(err error, spinner *spinner.Spinner) {
	if err != nil {
		if spinner != nil {
			// spinner.Suffix = ""
			spinner.FinalMSG = err.Error()
			spinner.Stop()
		} else {
			fmt.Println(err)
		}
		os.Exit(1)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if len(os.Args) > 1 && os.Args[1] == "stats" {
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
		if err := statsCmd.Execute(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	if err := todosCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
