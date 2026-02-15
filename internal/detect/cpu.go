package detect

import (
	"fmt"
	"strings"

	"github.com/krisk248/tuner/internal/output"
	"github.com/krisk248/tuner/internal/sysfs"
)

// CPUInfo holds CPU diagnostic data.
type CPUInfo struct {
	Model       string
	Cores       int
	Threads     int
	Governor    string
	AvailGovs   []string
	Driver      string
	EPP         string
	AvailEPP    []string
	CurFreqMHz  int
	MinFreqMHz  int
	MaxFreqMHz  int
	BaseFreqMHz int
	TurboEnabled bool
	TurboKnown   bool
	Architecture string
	Vendor       string
}

// DetectCPU gathers CPU information from sysfs and procfs.
func DetectCPU() CPUInfo {
	info := CPUInfo{}

	// Parse /proc/cpuinfo for model, vendor, core count
	if lines, err := sysfs.ReadLines(sysfs.ProcCPUInfo); err == nil {
		threads := 0
		coreIDs := make(map[string]bool)
		for _, line := range lines {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			switch key {
			case "model name":
				if info.Model == "" {
					info.Model = val
				}
			case "vendor_id":
				if info.Vendor == "" {
					info.Vendor = val
				}
			case "processor":
				threads++
			case "core id":
				coreIDs[val] = true
			}
		}
		info.Threads = threads
		info.Cores = len(coreIDs)
		if info.Cores == 0 {
			info.Cores = threads
		}
	}

	// Architecture
	if arch, err := sysfs.ReadString("/proc/sys/kernel/arch"); err == nil {
		info.Architecture = arch
	} else {
		info.Architecture = "x86_64"
	}

	// Frequency scaling
	if gov, err := sysfs.ReadString(sysfs.CPUGovernor); err == nil {
		info.Governor = gov
	}
	if govs, err := sysfs.ReadFields(sysfs.CPUAvailGovs); err == nil {
		info.AvailGovs = govs
	}
	if driver, err := sysfs.ReadString(sysfs.CPUDriver); err == nil {
		info.Driver = driver
	}

	// EPP (may not exist on older kernels)
	if epp, err := sysfs.ReadString(sysfs.CPUEPP); err == nil {
		info.EPP = epp
	}
	if epps, err := sysfs.ReadFields(sysfs.CPUAvailEPP); err == nil {
		info.AvailEPP = epps
	}

	// Frequencies (in KHz in sysfs, convert to MHz)
	if freq, err := sysfs.ReadInt(sysfs.CPUFreqCur); err == nil {
		info.CurFreqMHz = freq / 1000
	}
	if freq, err := sysfs.ReadInt(sysfs.CPUFreqMin); err == nil {
		info.MinFreqMHz = freq / 1000
	}
	if freq, err := sysfs.ReadInt(sysfs.CPUFreqMax); err == nil {
		info.MaxFreqMHz = freq / 1000
	}
	if freq, err := sysfs.ReadInt(sysfs.CPUFreqBaseFreq); err == nil {
		info.BaseFreqMHz = freq / 1000
	}

	// Turbo boost
	if sysfs.Exists(sysfs.IntelNoTurbo) {
		info.TurboKnown = true
		if val, err := sysfs.ReadInt(sysfs.IntelNoTurbo); err == nil {
			info.TurboEnabled = val == 0 // no_turbo=0 means turbo is ON
		}
	} else if sysfs.Exists(sysfs.CPUBoost) {
		info.TurboKnown = true
		if val, err := sysfs.ReadInt(sysfs.CPUBoost); err == nil {
			info.TurboEnabled = val == 1
		}
	}

	return info
}

// CPUSection formats CPU info as an output section.
func CPUSection(info CPUInfo) output.Section {
	sec := output.Section{Title: "CPU"}

	sec.Fields = append(sec.Fields,
		output.Field{Key: "Model", Value: info.Model, Status: output.StatusInfo},
		output.Field{Key: "Cores / Threads", Value: fmt.Sprintf("%d / %d", info.Cores, info.Threads), Status: output.StatusInfo},
		output.Field{Key: "Scaling Driver", Value: info.Driver, Status: output.StatusInfo},
		output.Field{Key: "Governor", Value: info.Governor, Status: governorStatus(info.Governor)},
	)

	if info.EPP != "" {
		sec.Fields = append(sec.Fields,
			output.Field{Key: "Energy Perf Pref", Value: info.EPP, Status: output.StatusInfo},
		)
	}

	if info.CurFreqMHz > 0 {
		sec.Fields = append(sec.Fields,
			output.Field{Key: "Frequency", Value: fmt.Sprintf("%d MHz (min: %d, max: %d)", info.CurFreqMHz, info.MinFreqMHz, info.MaxFreqMHz), Status: output.StatusInfo},
		)
	}

	if info.TurboKnown {
		turboStr := "disabled"
		status := output.StatusWarn
		if info.TurboEnabled {
			turboStr = "enabled"
			status = output.StatusGood
		}
		sec.Fields = append(sec.Fields,
			output.Field{Key: "Turbo Boost", Value: turboStr, Status: status},
		)
	}

	return sec
}

func governorStatus(gov string) output.Status {
	switch gov {
	case "performance", "schedutil":
		return output.StatusGood
	case "powersave":
		return output.StatusWarn
	default:
		return output.StatusInfo
	}
}
