package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func testSections() []Section {
	return []Section{
		{
			Title: "Test",
			Fields: []Field{
				{Key: "Key1", Value: "Value1", Status: StatusGood},
				{Key: "Key2", Value: "Value2", Status: StatusWarn},
			},
		},
	}
}

func TestJSONFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := &JSONFormatter{}
	if err := f.Format(&buf, testSections()); err != nil {
		t.Fatal(err)
	}

	// Should be valid JSON
	var result []jsonSection
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 section, got %d", len(result))
	}
	if result[0].Title != "Test" {
		t.Errorf("title = %q, want Test", result[0].Title)
	}
	if len(result[0].Fields) != 2 {
		t.Errorf("fields = %d, want 2", len(result[0].Fields))
	}
}

func TestMarkdownFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := &MarkdownFormatter{}
	if err := f.Format(&buf, testSections()); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "## Test") {
		t.Error("markdown missing section header")
	}
	if !strings.Contains(out, "| Key1 |") {
		t.Error("markdown missing field")
	}
}

func TestTableFormatter(t *testing.T) {
	var buf bytes.Buffer
	f := &TableFormatter{NoColor: true}
	if err := f.Format(&buf, testSections()); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Test") {
		t.Error("table missing section title")
	}
	if !strings.Contains(out, "Key1") {
		t.Error("table missing field key")
	}
}
