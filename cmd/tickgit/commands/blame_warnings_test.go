package commands

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/briandowns/spinner"
)

func TestFindToDosReportsBlameWarningsAndContinues(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "tracked.md"), []byte("# Notes\n\n- TODO tracked\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	initCommandsGitRepo(t, dir)
	err = os.WriteFile(filepath.Join(dir, "untracked.md"), []byte("# Notes\n\n- TODO untracked\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var warnings bytes.Buffer
	previousWarningWriter := blameWarningWriter
	previousMatchPhrases := matchPhrases
	blameWarningWriter = &warnings
	matchPhrases = nil
	t.Cleanup(func() {
		blameWarningWriter = previousWarningWriter
		matchPhrases = previousMatchPhrases
	})

	s := spinner.New(spinner.CharSets[9], time.Millisecond)
	found, err := findToDos(context.Background(), dir, s)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 2 {
		t.Fatalf("expected both tracked and untracked TODOs, got %d", len(found))
	}
	if got := found.CountWithCommits(); got != 1 {
		t.Fatalf("expected blame for tracked TODO only, got %d blamed TODOs", got)
	}
	if !strings.Contains(warnings.String(), "tickgit warning:") {
		t.Fatalf("expected warning prefix, got %q", warnings.String())
	}
	if !strings.Contains(warnings.String(), "untracked.md") {
		t.Fatalf("expected warning to include untracked file path, got %q", warnings.String())
	}
}

func initCommandsGitRepo(t *testing.T, dir string) {
	t.Helper()
	runCommandsGit(t, dir, "init")
	runCommandsGit(t, dir, "config", "user.email", "test@example.com")
	runCommandsGit(t, dir, "config", "user.name", "Test User")
	runCommandsGit(t, dir, "add", ".")
	runCommandsGit(t, dir, "commit", "-m", "initial")
}

func runCommandsGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, output)
	}
}
