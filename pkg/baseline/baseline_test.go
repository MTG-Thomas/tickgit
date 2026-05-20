package baseline

import (
	"strings"
	"testing"
)

func TestCompareCSVReportsFindingsMissingFromBaseline(t *testing.T) {
	baselineCSV := strings.NewReader("text,file_path,start_line,start_position,end_line,end_position,author,author_email,author_sha,author_time\nTODO old,pkg/old.go,10,1,10,9,A,a@example.com,abc,2026-01-01T00:00:00Z\n")
	currentCSV := strings.NewReader("text,file_path,start_line,start_position,end_line,end_position,author,author_email,author_sha,author_time\nTODO old,pkg/old.go,10,1,10,9,A,a@example.com,abc,2026-01-01T00:00:00Z\nFIXME new,pkg/new.go,4,1,4,10,B,b@example.com,def,2026-01-02T00:00:00Z\n")

	result, err := CompareCSV(baselineCSV, currentCSV)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.New) != 1 {
		t.Fatalf("expected one new finding, got %d", len(result.New))
	}

	got := result.New[0]
	if got.Text != "FIXME new" || got.FilePath != "pkg/new.go" || got.StartLine != "4" {
		t.Fatalf("unexpected new finding: %#v", got)
	}
}

func TestCompareCSVHandlesDuplicateFindingsByCount(t *testing.T) {
	baselineCSV := strings.NewReader("text,file_path,start_line,start_position,end_line,end_position,author,author_email,author_sha,author_time\nTODO repeated,pkg/file.go,10,1,10,9,A,a@example.com,abc,2026-01-01T00:00:00Z\n")
	currentCSV := strings.NewReader("text,file_path,start_line,start_position,end_line,end_position,author,author_email,author_sha,author_time\nTODO repeated,pkg/file.go,10,1,10,9,A,a@example.com,abc,2026-01-01T00:00:00Z\nTODO repeated,pkg/file.go,11,1,11,9,A,a@example.com,abc,2026-01-01T00:00:00Z\n")

	result, err := CompareCSV(baselineCSV, currentCSV)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.New) != 1 {
		t.Fatalf("expected duplicate over baseline to be reported, got %d new findings", len(result.New))
	}

	if result.New[0].StartLine != "11" {
		t.Fatalf("expected line 11 to be new, got line %s", result.New[0].StartLine)
	}
}
