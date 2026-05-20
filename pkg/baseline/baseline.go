package baseline

import (
	"encoding/csv"
	"fmt"
	"io"
)

// Finding is one row from tickgit CSV output.
type Finding struct {
	Text      string
	FilePath  string
	StartLine string
	Raw       []string
}

// Result contains findings that exist in current output but not baseline.
type Result struct {
	New []Finding
}

// CompareCSV returns findings present in currentCSV beyond the counts present in baselineCSV.
func CompareCSV(baselineCSV io.Reader, currentCSV io.Reader) (Result, error) {
	baselineFindings, err := readFindings(baselineCSV)
	if err != nil {
		return Result{}, fmt.Errorf("read baseline CSV: %w", err)
	}

	currentFindings, err := readFindings(currentCSV)
	if err != nil {
		return Result{}, fmt.Errorf("read current CSV: %w", err)
	}

	remaining := make(map[string]int)
	for _, finding := range baselineFindings {
		remaining[finding.fingerprint()]++
	}

	var result Result
	for _, finding := range currentFindings {
		key := finding.fingerprint()
		if remaining[key] > 0 {
			remaining[key]--
			continue
		}
		result.New = append(result.New, finding)
	}

	return result, nil
}

func readFindings(r io.Reader) ([]Finding, error) {
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}

	findings := make([]Finding, 0, len(records)-1)
	for i, record := range records {
		if i == 0 {
			continue
		}
		if len(record) < 3 {
			return nil, fmt.Errorf("row %d has %d columns, expected at least 3", i+1, len(record))
		}
		findings = append(findings, Finding{
			Text:      record[0],
			FilePath:  record[1],
			StartLine: record[2],
			Raw:       record,
		})
	}

	return findings, nil
}

func (f Finding) fingerprint() string {
	return f.FilePath + "\x00" + f.StartLine + "\x00" + f.Text
}
