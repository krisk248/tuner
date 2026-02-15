package detect

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/krisk248/tuner/internal/output"
)

// ServiceInfo holds systemd service diagnostic data.
type ServiceInfo struct {
	BootTime      string
	KernelTime    string
	UserspaceTime string
	SlowUnits     []SlowUnit
	FailedUnits   []string
	RunningCount  int
}

// SlowUnit is a service that took a long time to start.
type SlowUnit struct {
	Name     string
	Duration string
	Seconds  float64
}

// DetectServices gathers systemd service information.
func DetectServices() ServiceInfo {
	info := ServiceInfo{}

	// Boot timing
	out, err := exec.Command("systemd-analyze").Output()
	if err == nil {
		line := strings.TrimSpace(string(out))
		info.BootTime = line
		// Parse kernel + userspace times from output like:
		// "Startup finished in 1.234s (kernel) + 5.678s (userspace) = 6.912s"
		if idx := strings.Index(line, "(kernel)"); idx > 0 {
			parts := strings.Fields(line[:idx])
			if len(parts) > 0 {
				info.KernelTime = parts[len(parts)-1]
			}
		}
		if idx := strings.Index(line, "(userspace)"); idx > 0 {
			before := line[:idx]
			parts := strings.Fields(before)
			if len(parts) > 0 {
				info.UserspaceTime = parts[len(parts)-1]
			}
		}
	}

	// Slow units (systemd-analyze blame)
	out, err = exec.Command("systemd-analyze", "blame").Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for i, line := range lines {
			if i >= 10 { // Top 10 slowest
				break
			}
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				unit := SlowUnit{
					Duration: fields[0],
					Name:     fields[1],
				}
				// Parse seconds
				dur := fields[0]
				dur = strings.TrimSuffix(dur, "s")
				if strings.Contains(dur, "min") {
					// Handle "1min 30.123s" format
					dur = strings.ReplaceAll(dur, "min", "")
				}
				fmt.Sscanf(dur, "%f", &unit.Seconds)

				info.SlowUnits = append(info.SlowUnits, unit)
			}
		}
	}

	// Failed units
	out, err = exec.Command("systemctl", "--failed", "--no-pager", "--plain", "--no-legend").Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) > 0 && fields[0] != "" {
				info.FailedUnits = append(info.FailedUnits, fields[0])
			}
		}
	}

	// Count running services
	out, err = exec.Command("systemctl", "list-units", "--type=service", "--state=running", "--no-pager", "--plain", "--no-legend").Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) > 0 && lines[0] != "" {
			info.RunningCount = len(lines)
		}
	}

	return info
}

// ServicesSection formats service info as an output section.
func ServicesSection(info ServiceInfo) output.Section {
	sec := output.Section{Title: "Services"}

	if info.BootTime != "" {
		sec.Fields = append(sec.Fields,
			output.Field{Key: "Boot Time", Value: info.BootTime, Status: output.StatusInfo},
		)
	}

	sec.Fields = append(sec.Fields,
		output.Field{Key: "Running Services", Value: fmt.Sprintf("%d", info.RunningCount), Status: output.StatusInfo},
	)

	// Failed units
	if len(info.FailedUnits) > 0 {
		sec.Fields = append(sec.Fields,
			output.Field{Key: "Failed Units", Value: strings.Join(info.FailedUnits, ", "), Status: output.StatusBad},
		)
	} else {
		sec.Fields = append(sec.Fields,
			output.Field{Key: "Failed Units", Value: "none", Status: output.StatusGood},
		)
	}

	// Slowest boot units
	if len(info.SlowUnits) > 0 {
		sec.Fields = append(sec.Fields,
			output.Field{Key: "Slowest Boot Units", Value: "", Status: output.StatusNone},
		)
		// Sort by duration descending
		sort.Slice(info.SlowUnits, func(i, j int) bool {
			return info.SlowUnits[i].Seconds > info.SlowUnits[j].Seconds
		})
		for _, u := range info.SlowUnits {
			status := output.StatusInfo
			if u.Seconds > 10 {
				status = output.StatusWarn
			}
			sec.Fields = append(sec.Fields,
				output.Field{Key: fmt.Sprintf("  %s", u.Name), Value: u.Duration, Status: status},
			)
		}
	}

	return sec
}
