package detect

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/krisk248/tuner/internal/output"
	"github.com/krisk248/tuner/internal/sysfs"
)

// PowerInfo holds power/battery diagnostic data.
type PowerInfo struct {
	HasBattery    bool
	OnAC          bool
	BatteryPct    int
	BatteryStatus string // Charging, Discharging, Full, Not charging
	TLPActive     bool
	PPDActive     bool // power-profiles-daemon
	PowerProfile  string
	ACAdapters    []string
	Batteries     []BatteryInfo
}

// BatteryInfo holds info about a single battery.
type BatteryInfo struct {
	Name      string
	Status    string
	Capacity  int
	EnergyNow int64
	EnergyFull int64
	CycleCount int
}

// DetectPower gathers power supply and management information.
func DetectPower() PowerInfo {
	info := PowerInfo{}

	entries, err := os.ReadDir(sysfs.PowerSupplyBase)
	if err != nil {
		return info
	}

	for _, e := range entries {
		name := e.Name()
		base := filepath.Join(sysfs.PowerSupplyBase, name)

		psType, err := sysfs.ReadString(filepath.Join(base, "type"))
		if err != nil {
			continue
		}

		switch strings.ToLower(psType) {
		case "mains":
			info.ACAdapters = append(info.ACAdapters, name)
			if online, err := sysfs.ReadInt(filepath.Join(base, "online")); err == nil {
				info.OnAC = online == 1
			}

		case "battery":
			info.HasBattery = true
			bat := BatteryInfo{Name: name}

			if v, err := sysfs.ReadString(filepath.Join(base, "status")); err == nil {
				bat.Status = v
				if info.BatteryStatus == "" {
					info.BatteryStatus = v
				}
			}
			if v, err := sysfs.ReadInt(filepath.Join(base, "capacity")); err == nil {
				bat.Capacity = v
				if info.BatteryPct == 0 {
					info.BatteryPct = v
				}
			}
			if v, err := sysfs.ReadInt64(filepath.Join(base, "energy_now")); err == nil {
				bat.EnergyNow = v
			}
			if v, err := sysfs.ReadInt64(filepath.Join(base, "energy_full")); err == nil {
				bat.EnergyFull = v
			}
			if v, err := sysfs.ReadInt(filepath.Join(base, "cycle_count")); err == nil {
				bat.CycleCount = v
			}

			info.Batteries = append(info.Batteries, bat)
		}
	}

	// If no AC adapters found but has battery, check AC status via battery
	if info.HasBattery && len(info.ACAdapters) == 0 {
		if info.BatteryStatus != "Discharging" {
			info.OnAC = true
		}
	}

	// Check for TLP
	if _, err := exec.LookPath("tlp"); err == nil {
		// Check if TLP service is active
		out, err := exec.Command("systemctl", "is-active", "tlp").Output()
		if err == nil && strings.TrimSpace(string(out)) == "active" {
			info.TLPActive = true
		}
	}

	// Check for power-profiles-daemon
	out, err := exec.Command("systemctl", "is-active", "power-profiles-daemon").Output()
	if err == nil && strings.TrimSpace(string(out)) == "active" {
		info.PPDActive = true
		// Get current profile
		profileOut, err := exec.Command("powerprofilesctl", "get").Output()
		if err == nil {
			info.PowerProfile = strings.TrimSpace(string(profileOut))
		}
	}

	return info
}

// PowerSection formats power info as an output section.
func PowerSection(info PowerInfo) output.Section {
	sec := output.Section{Title: "Power"}

	if info.HasBattery {
		acStr := "battery"
		if info.OnAC {
			acStr = "AC"
		}
		sec.Fields = append(sec.Fields,
			output.Field{Key: "Power Source", Value: acStr, Status: output.StatusInfo},
		)

		for _, bat := range info.Batteries {
			status := output.StatusGood
			if bat.Capacity < 20 {
				status = output.StatusBad
			} else if bat.Capacity < 50 {
				status = output.StatusWarn
			}
			sec.Fields = append(sec.Fields,
				output.Field{Key: bat.Name, Value: fmt.Sprintf("%d%% (%s)", bat.Capacity, bat.Status), Status: status},
			)
			if bat.CycleCount > 0 {
				sec.Fields = append(sec.Fields,
					output.Field{Key: "  Cycle Count", Value: fmt.Sprintf("%d", bat.CycleCount), Status: output.StatusInfo},
				)
			}
		}
	} else {
		sec.Fields = append(sec.Fields,
			output.Field{Key: "Power Source", Value: "AC (no battery)", Status: output.StatusInfo},
		)
	}

	// Power management daemons
	if info.TLPActive {
		sec.Fields = append(sec.Fields,
			output.Field{Key: "TLP", Value: "active", Status: output.StatusWarn},
		)
	}
	if info.PPDActive {
		sec.Fields = append(sec.Fields,
			output.Field{Key: "power-profiles-daemon", Value: fmt.Sprintf("active (%s)", info.PowerProfile), Status: output.StatusInfo},
		)
	}

	return sec
}
