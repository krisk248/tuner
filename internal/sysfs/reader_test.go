package sysfs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadString(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test")
	os.WriteFile(path, []byte("hello\n"), 0644)

	got, err := ReadString(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Errorf("ReadString = %q, want %q", got, "hello")
	}
}

func TestReadInt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test")
	os.WriteFile(path, []byte("42\n"), 0644)

	got, err := ReadInt(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("ReadInt = %d, want 42", got)
	}
}

func TestReadBool(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test")

	tests := []struct {
		input string
		want  bool
	}{
		{"1\n", true},
		{"0\n", false},
		{"Y\n", true},
		{"N\n", false},
		{"yes\n", true},
		{"no\n", false},
	}

	for _, tt := range tests {
		os.WriteFile(path, []byte(tt.input), 0644)
		got, err := ReadBool(path)
		if err != nil {
			t.Errorf("ReadBool(%q) error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ReadBool(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test")
	os.WriteFile(path, []byte("x"), 0644)

	if !Exists(path) {
		t.Error("Exists returned false for existing file")
	}
	if Exists(filepath.Join(dir, "nonexistent")) {
		t.Error("Exists returned true for nonexistent file")
	}
}

func TestReadBracketedValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test")

	os.WriteFile(path, []byte("always [madvise] never\n"), 0644)
	got, err := ReadBracketedValue(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "madvise" {
		t.Errorf("ReadBracketedValue = %q, want %q", got, "madvise")
	}
}

func TestReadFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test")
	os.WriteFile(path, []byte("performance powersave schedutil\n"), 0644)

	got, err := ReadFields(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Errorf("ReadFields returned %d fields, want 3", len(got))
	}
	if got[0] != "performance" {
		t.Errorf("ReadFields[0] = %q, want %q", got[0], "performance")
	}
}

func TestReadLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test")
	os.WriteFile(path, []byte("line1\nline2\n\nline3\n"), 0644)

	got, err := ReadLines(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Errorf("ReadLines returned %d lines, want 3", len(got))
	}
}
