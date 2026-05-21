package stats

import (
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/MTG-Thomas/tickgit/pkg/todos"
)

// Report summarizes current findings with Git-derived history.
type Report struct {
	Total       int            `json:"total"`
	ByPhrase    map[string]int `json:"by_phrase"`
	ByAgeBucket map[string]int `json:"by_age_bucket"`
	ByDirectory map[string]int `json:"by_directory"`
	Oldest      []Finding      `json:"oldest"`
}

// Finding is a compact stats view of a current finding.
type Finding struct {
	Text          string    `json:"text"`
	Phrase        string    `json:"phrase"`
	FilePath      string    `json:"file_path"`
	Line          int       `json:"line"`
	Author        string    `json:"author"`
	AuthorEmail   string    `json:"author_email"`
	AuthorTime    time.Time `json:"author_time"`
	IntroducedSHA string    `json:"introduced_sha"`
	AgeBucket     string    `json:"age_bucket"`
}

// Build groups current findings into a deterministic stats report.
func Build(found todos.ToDos, now time.Time) Report {
	report := Report{
		Total:       len(found),
		ByPhrase:    make(map[string]int),
		ByAgeBucket: make(map[string]int),
		ByDirectory: make(map[string]int),
	}

	for _, todo := range found {
		report.ByPhrase[todo.Phrase]++
		report.ByDirectory[directoryBucket(todo.FilePath)]++

		ageBucket := "unknown"
		if todo.Blame != nil {
			ageBucket = bucketAge(now.Sub(todo.Blame.Author.When))
			report.Oldest = append(report.Oldest, Finding{
				Text:          todo.String,
				Phrase:        todo.Phrase,
				FilePath:      filepath.ToSlash(todo.FilePath),
				Line:          todo.StartLocation.Line,
				Author:        todo.Blame.Author.Name,
				AuthorEmail:   todo.Blame.Author.Email,
				AuthorTime:    todo.Blame.Author.When,
				IntroducedSHA: todo.Blame.SHA,
				AgeBucket:     ageBucket,
			})
		}
		report.ByAgeBucket[ageBucket]++
	}

	sort.Slice(report.Oldest, func(i, j int) bool {
		return report.Oldest[i].AuthorTime.Before(report.Oldest[j].AuthorTime)
	})
	if len(report.Oldest) > 10 {
		report.Oldest = report.Oldest[:10]
	}

	return report
}

func bucketAge(age time.Duration) string {
	days := int(age.Hours() / 24)
	switch {
	case days < 30:
		return "<30d"
	case days < 90:
		return "30-90d"
	case days < 365:
		return "90-365d"
	default:
		return ">365d"
	}
}

func directoryBucket(path string) string {
	path = filepath.ToSlash(path)
	dir, _, ok := strings.Cut(path, "/")
	if !ok || dir == "" {
		return "."
	}
	return dir
}
