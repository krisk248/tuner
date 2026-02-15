package detect

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/krisk248/tuner/internal/output"
	"github.com/krisk248/tuner/internal/sysfs"
)

// SMARTInfo holds disk health data.
type SMARTInfo struct {
	PercentUsed    int // wear level 0-100
	PowerOnHours   int
	UnsafeShutdowns int
	MediaErrors    int
}

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
	SMART      *SMARTInfo
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

		// SMART data for NVMe
		if disk.Type == "nvme" {
			disk.SMART = detectNVMeSMART(name)
		}

		info.Disks = append(info.Disks, disk)
	}

	// Parse /proc/mounts
	if lines, err := sysfs.ReadLines("/proc/mounts"); err == nil {
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) < 3 {
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

		if disk.SMART != nil {
			s := disk.SMART
			wearStatus := output.StatusGood
			if s.PercentUsed > 90 {
				wearStatus = output.StatusBad
			} else if s.PercentUsed > 70 {
				wearStatus = output.StatusWarn
			}
			sec.Fields = append(sec.Fields,
				output.Field{Key: "  Wear Level", Value: fmt.Sprintf("%d%% used", s.PercentUsed), Status: wearStatus},
				output.Field{Key: "  Power On", Value: fmt.Sprintf("%d hours", s.PowerOnHours), Status: output.StatusInfo},
			)
			if s.UnsafeShutdowns > 0 {
				sec.Fields = append(sec.Fields,
					output.Field{Key: "  Unsafe Shutdowns", Value: fmt.Sprintf("%d", s.UnsafeShutdowns), Status: output.StatusWarn},
				)
			}
			if s.MediaErrors > 0 {
				sec.Fields = append(sec.Fields,
					output.Field{Key: "  Media Errors", Value: fmt.Sprintf("%d", s.MediaErrors), Status: output.StatusBad},
				)
			}
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

func detectNVMeSMART(diskName string) *SMARTInfo {
	// Try nvme smart-log (needs root, but best data)
	out, err := exec.Command("nvme", "smart-log", "/dev/"+diskName).Output()
	if err != nil {
		return nil
	}

	s := &SMARTInfo{}
	for _, line := range strings.Split(string(out), "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(strings.ToLower(parts[0]))
		val := strings.TrimSpace(parts[1])
		// Strip trailing % or commas
		val = strings.ReplaceAll(val, ",", "")
		val = strings.TrimSuffix(val, "%")
		val = strings.TrimSpace(val)

		switch key {
		case "percentage used":
			s.PercentUsed, _ = strconv.Atoi(val)
		case "power on hours":
			s.PowerOnHours, _ = strconv.Atoi(val)
		case "unsafe shutdowns":
			s.UnsafeShutdowns, _ = strconv.Atoi(val)
		case "media and data integrity errors":
			s.MediaErrors, _ = strconv.Atoi(val)
		}
	}
	return s
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
