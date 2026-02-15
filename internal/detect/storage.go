package detect

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/krisk248/tuner/internal/output"
	"github.com/krisk248/tuner/internal/sysfs"
)

// DiskInfo holds per-device storage info.
type DiskInfo struct {
	Name       string
	Type       string // nvme, ssd, hdd
	Scheduler  string
	AvailScheds []string
	Rotational bool
	SizeGB     float64
	Model      string
	NrRequests int
	ReadAhead  int
}

// StorageInfo holds overall storage diagnostic data.
type StorageInfo struct {
	Disks      []DiskInfo
	Mounts     []MountInfo
	FileSystems map[string]string // device -> fstype
}

// MountInfo holds mount point info.
type MountInfo struct {
	Device     string
	MountPoint string
	FSType     string
	Options    string
}

// DetectStorage gathers storage device information.
func DetectStorage() StorageInfo {
	info := StorageInfo{
		FileSystems: make(map[string]string),
	}

	// Read block devices
	entries, err := os.ReadDir(sysfs.BlockBase)
	if err != nil {
		return info
	}

	for _, e := range entries {
		name := e.Name()
		// Skip loop, ram, dm, zram devices
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") ||
			strings.HasPrefix(name, "zram") || strings.HasPrefix(name, "dm-") {
			continue
		}

		disk := DiskInfo{Name: name}
		base := filepath.Join(sysfs.BlockBase, name)

		// Rotational
		if v, err := sysfs.ReadInt(filepath.Join(base, "queue/rotational")); err == nil {
			disk.Rotational = v == 1
		}

		// Determine type
		if strings.HasPrefix(name, "nvme") {
			disk.Type = "nvme"
		} else if disk.Rotational {
			disk.Type = "hdd"
		} else {
			disk.Type = "ssd"
		}

		// Scheduler
		schedPath := filepath.Join(base, "queue/scheduler")
		if sched, err := sysfs.ReadBracketedValue(schedPath); err == nil {
			disk.Scheduler = sched
		}
		if raw, err := sysfs.ReadString(schedPath); err == nil {
			raw = strings.ReplaceAll(raw, "[", "")
			raw = strings.ReplaceAll(raw, "]", "")
			disk.AvailScheds = strings.Fields(raw)
		}

		// Size (in 512-byte sectors)
		if sectors, err := sysfs.ReadInt64(filepath.Join(base, "size")); err == nil {
			disk.SizeGB = float64(sectors) * 512 / (1024 * 1024 * 1024)
		}

		// Model
		if model, err := sysfs.ReadString(filepath.Join(base, "device/model")); err == nil {
			disk.Model = model
		}

		// Nr requests
		if v, err := sysfs.ReadInt(filepath.Join(base, "queue/nr_requests")); err == nil {
			disk.NrRequests = v
		}

		// Read ahead
		if v, err := sysfs.ReadInt(filepath.Join(base, "queue/read_ahead_kb")); err == nil {
			disk.ReadAhead = v
		}

		info.Disks = append(info.Disks, disk)
	}

	// Parse /proc/mounts
	if lines, err := sysfs.ReadLines("/proc/mounts"); err == nil {
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) < 4 {
				continue
			}
			// Only show real filesystems
			if !strings.HasPrefix(fields[0], "/") {
				continue
			}
			mount := MountInfo{
				Device:     fields[0],
				MountPoint: fields[1],
				FSType:     fields[2],
				Options:    fields[3],
			}
			info.Mounts = append(info.Mounts, mount)
			info.FileSystems[fields[0]] = fields[2]
		}
	}

	return info
}

// StorageSection formats storage info as an output section.
func StorageSection(info StorageInfo) output.Section {
	sec := output.Section{Title: "Storage"}

	for _, disk := range info.Disks {
		prefix := fmt.Sprintf("%s (%s)", disk.Name, disk.Type)
		if disk.Model != "" {
			prefix = fmt.Sprintf("%s [%s]", prefix, disk.Model)
		}

		sec.Fields = append(sec.Fields,
			output.Field{Key: prefix, Value: fmt.Sprintf("%.0f GB", disk.SizeGB), Status: output.StatusInfo},
			output.Field{Key: "  Scheduler", Value: disk.Scheduler, Status: schedulerStatus(disk.Type, disk.Scheduler)},
		)

		if disk.ReadAhead > 0 {
			sec.Fields = append(sec.Fields,
				output.Field{Key: "  Read Ahead", Value: fmt.Sprintf("%d KB", disk.ReadAhead), Status: output.StatusInfo},
			)
		}
	}

	// Show filesystems
	for _, m := range info.Mounts {
		if m.MountPoint == "/" || m.MountPoint == "/home" || m.MountPoint == "/boot" {
			sec.Fields = append(sec.Fields,
				output.Field{Key: m.MountPoint, Value: fmt.Sprintf("%s (%s)", m.Device, m.FSType), Status: output.StatusInfo},
			)
		}
	}

	return sec
}

func schedulerStatus(diskType, sched string) output.Status {
	switch diskType {
	case "nvme":
		if sched == "none" {
			return output.StatusGood
		}
		return output.StatusWarn
	case "ssd":
		if sched == "kyber" || sched == "bfq" || sched == "mq-deadline" {
			return output.StatusGood
		}
		return output.StatusWarn
	case "hdd":
		if sched == "bfq" || sched == "mq-deadline" {
			return output.StatusGood
		}
		return output.StatusWarn
	}
	return output.StatusInfo
}
