package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

func validateDir(dir string) {
	if dir == "" {
		cwd, err := os.Getwd()
		handleError(err, nil)
		dir = cwd
	}

	abs, err := filepath.Abs(filepath.Join(dir, ".git"))
	handleError(err, nil)

	if _, err := os.Stat(abs); os.IsNotExist(err) {
		handleError(fmt.Errorf("%s is not a git repository", abs), nil)
	}
}

func resolveSearchDir(cwd string, args []string) (string, error) {
	if len(args) == 0 {
		return cwd, nil
	}

	arg := args[0]
	if !filepath.IsAbs(arg) {
		arg = filepath.Join(cwd, arg)
	}

	arg, err := filepath.Abs(arg)
	if err != nil {
		return "", err
	}

	return filepath.Rel(cwd, arg)
}
