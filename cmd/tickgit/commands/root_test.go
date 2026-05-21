package commands

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/briandowns/spinner"
)

func TestReportCommandErrorStopsSpinnerAndWritesToStderr(t *testing.T) {
	var spinnerOutput bytes.Buffer
	var stderr bytes.Buffer
	s := spinner.New(spinner.CharSets[9], time.Millisecond)
	s.Writer = &spinnerOutput
	s.FinalMSG = "stale final message"
	s.Start()

	code := reportCommandError(errors.New("scan failed"), s, &stderr)

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if s.Active() {
		t.Fatal("expected spinner to stop")
	}
	if got := stderr.String(); got != "scan failed\n" {
		t.Fatalf("expected error on stderr, got %q", got)
	}
	if strings.Contains(spinnerOutput.String(), "stale final message") {
		t.Fatalf("expected spinner final message to be suppressed, got %q", spinnerOutput.String())
	}
}

func TestReportCommandErrorIgnoresNilError(t *testing.T) {
	var stderr bytes.Buffer

	code := reportCommandError(nil, nil, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}
