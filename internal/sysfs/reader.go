package sysfs

import (
	"os"
	"strconv"
	"strings"
)

// Exists returns true if the path exists on the filesystem.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ReadString reads a sysfs/procfs file and returns its trimmed content.
func ReadString(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// ReadInt reads a sysfs/procfs file and parses it as an integer.
func ReadInt(path string) (int, error) {
	s, err := ReadString(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(s)
}

// ReadInt64 reads a sysfs/procfs file and parses it as an int64.
func ReadInt64(path string) (int64, error) {
	s, err := ReadString(path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(s, 10, 64)
}

// ReadBool reads a sysfs/procfs file and parses common boolean representations.
// "1", "Y", "y", "yes", "true" -> true; "0", "N", "n", "no", "false" -> false.
func ReadBool(path string) (bool, error) {
	s, err := ReadString(path)
	if err != nil {
		return false, err
	}
	switch strings.ToLower(s) {
	case "1", "y", "yes", "true":
		return true, nil
	case "0", "n", "no", "false":
		return false, nil
	default:
		return false, &os.PathError{Op: "parse", Path: path, Err: os.ErrInvalid}
	}
}

// ReadLines reads a file and returns non-empty lines.
func ReadLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

// ReadFields reads a file and splits the first line by whitespace.
func ReadFields(path string) ([]string, error) {
	s, err := ReadString(path)
	if err != nil {
		return nil, err
	}
	return strings.Fields(s), nil
}

// ReadBracketedValue reads a sysfs file where the active value is in [brackets].
// e.g. "always [madvise] never" returns "madvise".
func ReadBracketedValue(path string) (string, error) {
	s, err := ReadString(path)
	if err != nil {
		return "", err
	}
	start := strings.Index(s, "[")
	end := strings.Index(s, "]")
	if start < 0 || end < 0 || end <= start {
		return s, nil
	}
	return s[start+1 : end], nil
}
