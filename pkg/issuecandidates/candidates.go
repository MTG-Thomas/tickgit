package issuecandidates

import (
	"crypto/sha1"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Candidate is a tracker-ready issue candidate derived from one tickgit CSV row.
type Candidate struct {
	Repository string
	Text       string
	FilePath   string
	StartLine  string
	Author     string
	AuthorTime string
	Key        string
}

// ReadCSV reads tickgit CSV output and converts rows into issue candidates.
func ReadCSV(reader io.Reader, repository string) ([]Candidate, error) {
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}

	header := map[string]int{}
	for index, name := range records[0] {
		header[name] = index
	}

	required := []string{"text", "file_path", "start_line"}
	for _, name := range required {
		if _, ok := header[name]; !ok {
			return nil, fmt.Errorf("missing required CSV column %q", name)
		}
	}

	candidates := make([]Candidate, 0, len(records)-1)
	seen := map[string]struct{}{}
	for _, record := range records[1:] {
		candidate := Candidate{
			Repository: repository,
			Text:       csvValue(record, header, "text"),
			FilePath:   csvValue(record, header, "file_path"),
			StartLine:  csvValue(record, header, "start_line"),
			Author:     csvValue(record, header, "author"),
			AuthorTime: csvValue(record, header, "author_time"),
		}
		candidate.Key = StableKey(candidate)
		if _, ok := seen[candidate.Key]; ok {
			continue
		}
		seen[candidate.Key] = struct{}{}
		candidates = append(candidates, candidate)
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].FilePath == candidates[j].FilePath {
			return candidates[i].StartLine < candidates[j].StartLine
		}
		return candidates[i].FilePath < candidates[j].FilePath
	})

	return candidates, nil
}

// StableKey returns a deterministic duplicate-detection key for a candidate.
func StableKey(candidate Candidate) string {
	sum := sha1.Sum([]byte(strings.Join([]string{
		candidate.Repository,
		candidate.FilePath,
		candidate.StartLine,
		normalizeText(candidate.Text),
	}, "\x00")))
	return hex.EncodeToString(sum[:])[:12]
}

// WriteMarkdown writes issue-candidate markdown for human or agent curation.
func WriteMarkdown(writer io.Writer, candidates []Candidate) error {
	if _, err := fmt.Fprintf(writer, "# Tickgit issue candidates\n\n"); err != nil {
		return err
	}
	if len(candidates) == 0 {
		_, err := fmt.Fprintln(writer, "No candidates found.")
		return err
	}

	for _, candidate := range candidates {
		if _, err := fmt.Fprintf(writer, "## %s:%s\n\n", candidate.FilePath, candidate.StartLine); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(writer, "- duplicate_key: `%s`\n", candidate.Key); err != nil {
			return err
		}
		if candidate.Repository != "" {
			if _, err := fmt.Fprintf(writer, "- repo: `%s`\n", candidate.Repository); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(writer, "- evidence: `%s:%s`\n", candidate.FilePath, candidate.StartLine); err != nil {
			return err
		}
		if candidate.Author != "" || candidate.AuthorTime != "" {
			if _, err := fmt.Fprintf(writer, "- introduced_by: `%s` `%s`\n", candidate.Author, candidate.AuthorTime); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(writer, "- comment: %s\n\n", quoteMarkdown(candidate.Text)); err != nil {
			return err
		}
	}

	return nil
}

func csvValue(record []string, header map[string]int, name string) string {
	index, ok := header[name]
	if !ok || index >= len(record) {
		return ""
	}
	return record[index]
}

func normalizeText(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func quoteMarkdown(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	for index, line := range lines {
		lines[index] = "> " + line
	}
	return "\n" + strings.Join(lines, "\n")
}
