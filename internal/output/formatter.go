package output

import "io"

// Section represents a named group of key-value pairs for display.
type Section struct {
	Title  string
	Fields []Field
}

// Field is a single key-value pair with optional status.
type Field struct {
	Key    string
	Value  string
	Status Status
}

// Status indicates the health/optimization state of a field.
type Status int

const (
	StatusNone Status = iota
	StatusGood
	StatusWarn
	StatusBad
	StatusInfo
)

// Formatter defines the interface for output rendering.
type Formatter interface {
	Format(w io.Writer, sections []Section) error
}

// NewFormatter returns a formatter for the given format string.
func NewFormatter(format string, noColor bool) Formatter {
	switch format {
	case "json":
		return &JSONFormatter{}
	case "markdown":
		return &MarkdownFormatter{}
	default:
		return &TableFormatter{NoColor: noColor}
	}
}
