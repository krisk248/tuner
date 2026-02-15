package output

import (
	"fmt"
	"io"
)

// MarkdownFormatter renders sections as markdown tables.
type MarkdownFormatter struct{}

func (f *MarkdownFormatter) Format(w io.Writer, sections []Section) error {
	for i, sec := range sections {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "## %s\n\n", sec.Title)
		fmt.Fprintln(w, "| Parameter | Value | Status |")
		fmt.Fprintln(w, "|-----------|-------|--------|")
		for _, field := range sec.Fields {
			status := statusString(field.Status)
			if status == "" {
				status = "-"
			}
			fmt.Fprintf(w, "| %s | %s | %s |\n", field.Key, field.Value, status)
		}
	}
	return nil
}
