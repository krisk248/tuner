package sysfs

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// WriteString writes a string value to a sysfs/procfs path.
func WriteString(path, value string) error {
	return os.WriteFile(path, []byte(value), 0644)
}

// WriteInt writes an integer value to a sysfs/procfs path.
func WriteInt(path string, value int) error {
	return WriteString(path, strconv.Itoa(value))
}

// WriteInt64 writes an int64 value to a sysfs/procfs path.
func WriteInt64(path string, value int64) error {
	return WriteString(path, strconv.FormatInt(value, 10))
}

// WriteSysctl writes a value to a /proc/sys path.
// key is in dotted form like "vm.swappiness", value is the string to write.
func WriteSysctl(key, value string) error {
	path := "/proc/sys/" + sysCtlKeyToPath(key)
	return WriteString(path, value)
}

func sysCtlKeyToPath(key string) string {
	return strings.ReplaceAll(key, ".", "/")
}

// WriteAllCPUs writes a value to a per-CPU sysfs attribute for all CPUs.
func WriteAllCPUs(attr, value string) error {
	entries, err := os.ReadDir(CPUBase)
	if err != nil {
		return err
	}
	var lastErr error
	for _, e := range entries {
		name := e.Name()
		if len(name) > 3 && name[:3] == "cpu" && name[3] >= '0' && name[3] <= '9' {
			path := fmt.Sprintf("%s/%s/cpufreq/%s", CPUBase, name, attr)
			if Exists(path) {
				if err := WriteString(path, value); err != nil {
					lastErr = err
				}
			}
		}
	}
	return lastErr
}
