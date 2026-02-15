package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
)

// TableFormatter renders sections as colored aligned text.
type TableFormatter struct {
	NoColor bool
}

func (f *TableFormatter) Format(w io.Writer, sections []Section) error {
	if f.NoColor {
		color.NoColor = true
	}

	for i, sec := range sections {
		if i > 0 {
			fmt.Fprintln(w)
		}

		header := color.New(color.Bold, color.FgCyan)
		header.Fprintf(w, "── %s ──\n", sec.Title)

		// Find max key width for alignment
		maxKey := 0
		for _, field := range sec.Fields {
			if len(field.Key) > maxKey {
				maxKey = len(field.Key)
			}
		}
		if maxKey > 30 {
			maxKey = 30
		}

		for _, field := range sec.Fields {
			indicator := f.statusIndicator(field.Status)
			val := f.colorize(field.Value, field.Status)
			padding := strings.Repeat(" ", max(1, maxKey-len(field.Key)+2))

			if field.Value == "" {
				fmt.Fprintf(w, "  %s\n", field.Key)
			} else {
				fmt.Fprintf(w, "  %s%s%s %s\n", field.Key, padding, val, indicator)
			}
		}
	}
	return nil
}

func (f *TableFormatter) colorize(val string, status Status) string {
	if f.NoColor {
		return val
	}
	switch status {
	case StatusGood:
		return color.GreenString(val)
	case StatusWarn:
		return color.YellowString(val)
	case StatusBad:
		return color.RedString(val)
	case StatusInfo:
		return color.CyanString(val)
	default:
		return val
	}
}

func (f *TableFormatter) statusIndicator(status Status) string {
	if f.NoColor {
		switch status {
		case StatusGood:
			return "[OK]"
		case StatusWarn:
			return "[WARN]"
		case StatusBad:
			return "[BAD]"
		default:
			return ""
		}
	}
	switch status {
	case StatusGood:
		return color.GreenString("●")
	case StatusWarn:
		return color.YellowString("●")
	case StatusBad:
		return color.RedString("●")
	default:
		return ""
	}
}
