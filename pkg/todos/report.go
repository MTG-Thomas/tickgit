package todos

import (
	"io"
	"os"
	"strings"
	"text/template"
)

const (
	ColorAuto   ColorMode = "auto"
	ColorAlways ColorMode = "always"
	ColorNever  ColorMode = "never"
)

// ColorMode controls ANSI color rendering in human-readable reports.
type ColorMode string

// ReportOptions configures human-readable report rendering.
type ReportOptions struct {
	Color ColorMode
}

const defaultTemplate = `
{{- range $index, $todo := . }}
{{ .String }}
  => {{ .Comment.FilePath }}:{{ .Comment.StartLocation.Line }}:{{ .Comment.StartLocation.Pos }}
  {{- if .Context }}
  => context:
    {{- range .Context }}
     {{ printf "%6d" .Line }} | {{ .Text }}
    {{- end }}
  {{- end }}
  {{- if .Blame }}
  => added {{ .TimeAgo }} by {{ .Blame.Author }} in {{ .Blame.SHA }}
  {{- end }}
{{ else }}
no todos 🎉
{{- end }}
{{ len . }} TODOs Found 📝
`

// WriteTodos renders a report of todos
func WriteTodos(todos ToDos, writer io.Writer) error {
	return WriteTodosWithOptions(todos, writer, ReportOptions{Color: ColorAuto})
}

// WriteTodosWithOptions renders a report of todos with explicit report options.
func WriteTodosWithOptions(todos ToDos, writer io.Writer, options ReportOptions) error {

	t, err := template.New("todos").Parse(defaultTemplate)
	if err != nil {
		return err
	}

	renderTodos := todosForReport(todos, options.colorEnabled())

	err = t.Execute(writer, renderTodos)
	if err != nil {
		return err
	}

	return nil
}

// ParseColorMode converts a CLI color mode value to a report color mode.
func ParseColorMode(value string) (ColorMode, bool) {
	switch ColorMode(value) {
	case ColorAuto, ColorAlways, ColorNever:
		return ColorMode(value), true
	default:
		return "", false
	}
}

func (options ReportOptions) colorEnabled() bool {
	switch options.Color {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	case ColorAuto, "":
		_, noColor := os.LookupEnv("NO_COLOR")
		return !noColor
	default:
		return true
	}
}

func todosForReport(found ToDos, colorEnabled bool) ToDos {
	renderTodos := make(ToDos, 0, len(found))
	for _, todo := range found {
		renderTodo := *todo
		if colorEnabled {
			renderTodo.String = strings.Replace(renderTodo.String, renderTodo.Phrase, "\u001b[33m"+renderTodo.Phrase+"\u001b[0m", 1)
		}
		renderTodos = append(renderTodos, &renderTodo)
	}
	return renderTodos
}
