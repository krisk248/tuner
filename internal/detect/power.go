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

// ServiceState holds systemd service status.
type ServiceState struct {
	Installed bool
	Active    bool // is-active == "active"
	Enabled   bool // is-enabled == "enabled"
}

// PowerInfo holds power/battery diagnostic data.
type PowerInfo struct {
	HasBattery    bool
	OnAC          bool
	BatteryPct    int
	BatteryStatus string // Charging, Discharging, Full, Not charging
	TLP           ServiceState
	Tuned         ServiceState
	TunedProfile  string
	PPD           ServiceState
	PowerProfile  string
	ACAdapters    []string
	Batteries     []BatteryInfo
}


// BatteryInfo holds info about a single battery.
type BatteryInfo struct {
	Name             string
	Status           string
	Capacity         int
	EnergyNow        int64
	EnergyFull       int64
	EnergyFullDesign int64
	CycleCount       int
	HealthPct        float64 // (energy_full / energy_full_design) * 100
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
			if v, err := sysfs.ReadInt64(filepath.Join(base, "energy_full_design")); err == nil {
				bat.EnergyFullDesign = v
			}
			if v, err := sysfs.ReadInt(filepath.Join(base, "cycle_count")); err == nil {
				bat.CycleCount = v
			}

			// Calculate battery health
			if bat.EnergyFullDesign > 0 && bat.EnergyFull > 0 {
				bat.HealthPct = float64(bat.EnergyFull) / float64(bat.EnergyFullDesign) * 100
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
	info.TLP = checkService("tlp")

	// Check for tuned - only read profile if actually running
	info.Tuned = checkService("tuned")
	if info.Tuned.Active {
		if profileOut, err := exec.Command("tuned-adm", "active").Output(); err == nil {
			line := strings.TrimSpace(string(profileOut))
			if idx := strings.LastIndex(line, ": "); idx != -1 {
				info.TunedProfile = line[idx+2:]
			}
		}
	}

	// Check for power-profiles-daemon
	info.PPD = checkService("power-profiles-daemon")
	if info.PPD.Active {
		if profileOut, err := exec.Command("powerprofilesctl", "get").Output(); err == nil {
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
			if bat.HealthPct > 0 {
				healthStatus := output.StatusGood
				if bat.HealthPct < 80 {
					healthStatus = output.StatusBad
				} else if bat.HealthPct < 90 {
					healthStatus = output.StatusWarn
				}
				healthVal := fmt.Sprintf("%.0f%%", bat.HealthPct)
				if bat.CycleCount > 0 {
					healthVal = fmt.Sprintf("%.0f%% (%d cycles)", bat.HealthPct, bat.CycleCount)
				}
				sec.Fields = append(sec.Fields,
					output.Field{Key: "  Battery Health", Value: healthVal, Status: healthStatus},
				)
			} else if bat.CycleCount > 0 {
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
	if info.TLP.Installed {
		sec.Fields = append(sec.Fields,
			output.Field{Key: "TLP", Value: serviceStr(info.TLP), Status: serviceStatus(info.TLP)},
		)
	}
	if info.Tuned.Installed {
		val := serviceStr(info.Tuned)
		if info.Tuned.Active && info.TunedProfile != "" {
			val += fmt.Sprintf(" [%s]", info.TunedProfile)
		}
		sec.Fields = append(sec.Fields,
			output.Field{Key: "tuned", Value: val, Status: serviceStatus(info.Tuned)},
		)
		if info.Tuned.Active && info.TLP.Active {
			sec.Fields = append(sec.Fields,
				output.Field{Key: "  Warning", Value: "TLP and tuned both active - may conflict", Status: output.StatusBad},
			)
		}
	}
	if info.PPD.Installed {
		val := serviceStr(info.PPD)
		if info.PPD.Active && info.PowerProfile != "" {
			val += fmt.Sprintf(" [%s]", info.PowerProfile)
		}
		sec.Fields = append(sec.Fields,
			output.Field{Key: "power-profiles-daemon", Value: val, Status: serviceStatus(info.PPD)},
		)
	}

	return sec
}

// checkService checks if a systemd service is installed, active, and enabled.
func checkService(name string) ServiceState {
	s := ServiceState{}

	// is-active: "active", "inactive", "failed", etc. Exit code != 0 if not active.
	activeOut, _ := exec.Command("systemctl", "is-active", name).Output()
	activeStr := strings.TrimSpace(string(activeOut))

	// If systemctl returns nothing or "unknown", service isn't installed
	if activeStr == "" {
		return s
	}

	s.Installed = true
	s.Active = activeStr == "active"

	// is-enabled: "enabled", "disabled", "masked", "static", etc.
	enabledOut, _ := exec.Command("systemctl", "is-enabled", name).Output()
	enabledStr := strings.TrimSpace(string(enabledOut))
	s.Enabled = enabledStr == "enabled"

	return s
}

func serviceStr(s ServiceState) string {
	active := "inactive"
	if s.Active {
		active = "active"
	}
	enabled := "disabled"
	if s.Enabled {
		enabled = "enabled"
	}
	return fmt.Sprintf("%s (%s)", active, enabled)
}

func serviceStatus(s ServiceState) output.Status {
	if s.Active {
		return output.StatusWarn
	}
	return output.StatusInfo
}
