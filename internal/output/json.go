package output

import (
	"encoding/json"
	"io"
)

// JSONFormatter renders sections as JSON.
type JSONFormatter struct{}

type jsonSection struct {
	Title  string      `json:"title"`
	Fields []jsonField `json:"fields"`
}

type jsonField struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Status string `json:"status,omitempty"`
}

func (f *JSONFormatter) Format(w io.Writer, sections []Section) error {
	out := make([]jsonSection, len(sections))
	for i, sec := range sections {
		out[i] = jsonSection{Title: sec.Title}
		out[i].Fields = make([]jsonField, len(sec.Fields))
		for j, field := range sec.Fields {
			out[i].Fields[j] = jsonField{
				Key:    field.Key,
				Value:  field.Value,
				Status: statusString(field.Status),
			}
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func statusString(s Status) string {
	switch s {
	case StatusGood:
		return "good"
	case StatusWarn:
		return "warn"
	case StatusBad:
		return "bad"
	case StatusInfo:
		return "info"
	default:
		return ""
	}
}
