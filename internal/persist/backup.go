package persist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	BackupDir  = "/etc/tuner"
	BackupFile = "/etc/tuner/backup.json"
)

// BackupData holds the original values before tuning.
type BackupData struct {
	Timestamp string            `json:"timestamp"`
	Profile   string            `json:"profile"`
	Values    map[string]string `json:"values"` // path -> original value
}

// SaveBackup writes the backup data to disk.
func SaveBackup(profile string, values map[string]string) error {
	if err := os.MkdirAll(BackupDir, 0755); err != nil {
		return err
	}

	data := BackupData{
		Timestamp: time.Now().Format(time.RFC3339),
		Profile:   profile,
		Values:    values,
	}

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(BackupFile, b, 0644)
}

// LoadBackup reads the backup data from disk.
func LoadBackup() (*BackupData, error) {
	b, err := os.ReadFile(BackupFile)
	if err != nil {
		return nil, err
	}

	var data BackupData
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// BackupExists returns true if a backup file exists.
func BackupExists() bool {
	_, err := os.Stat(BackupFile)
	return err == nil
}

// RemoveBackup deletes the backup file.
func RemoveBackup() error {
	return os.Remove(BackupFile)
}

// SysctlPath returns the path for the tuner sysctl drop-in.
func SysctlPath() string {
	return filepath.Join("/etc/sysctl.d", "99-tuner.conf")
}

// UdevPath returns the path for the tuner udev rules.
func UdevPath() string {
	return filepath.Join("/etc/udev/rules.d", "99-tuner-disk.rules")
}
