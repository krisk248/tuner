package tune

import (
	"github.com/krisk248/tuner/internal/profile"
	"github.com/krisk248/tuner/internal/sysfs"
)

func computeCPUChanges(v profile.Values) []Change {
	var changes []Change

	// Governor
	if cur, err := sysfs.ReadString(sysfs.CPUGovernor); err == nil && cur != v.Governor {
		target := v.Governor
		changes = append(changes, Change{
			Subsystem: "cpu",
			Parameter: "CPU Governor",
			OldValue:  cur,
			NewValue:  target,
			Path:      sysfs.CPUGovernor,
			ApplyFunc: func() error {
				return sysfs.WriteAllCPUs("scaling_governor", target)
			},
		})
	}

	// EPP
	if sysfs.Exists(sysfs.CPUEPP) {
		if cur, err := sysfs.ReadString(sysfs.CPUEPP); err == nil && cur != v.EPP {
			target := v.EPP
			changes = append(changes, Change{
				Subsystem: "cpu",
				Parameter: "Energy Perf Pref",
				OldValue:  cur,
				NewValue:  target,
				Path:      sysfs.CPUEPP,
				ApplyFunc: func() error {
					return sysfs.WriteAllCPUs("energy_performance_preference", target)
				},
			})
		}
	}

	// Turbo boost
	type turboSource struct {
		path  string
		curOn bool
	}
	var turbo *turboSource
	if val, err := sysfs.ReadInt(sysfs.IntelNoTurbo); err == nil {
		turbo = &turboSource{path: sysfs.IntelNoTurbo, curOn: val == 0}
	} else if val, err := sysfs.ReadInt(sysfs.CPUBoost); err == nil {
		turbo = &turboSource{path: sysfs.CPUBoost, curOn: val == 1}
	}

	if turbo != nil && turbo.curOn != v.TurboOn {
		// For IntelNoTurbo: 0=on, 1=off; for CPUBoost: 1=on, 0=off
		var newVal string
		if turbo.path == sysfs.IntelNoTurbo {
			newVal = "1"
			if v.TurboOn {
				newVal = "0"
			}
		} else {
			newVal = "0"
			if v.TurboOn {
				newVal = "1"
			}
		}
		path := turbo.path
		changes = append(changes, Change{
			Subsystem: "cpu",
			Parameter: "Turbo Boost",
			OldValue:  boolToOnOff(turbo.curOn),
			NewValue:  boolToOnOff(v.TurboOn),
			Path:      path,
			ApplyFunc: func() error {
				return sysfs.WriteString(path, newVal)
			},
		})
	}

	return changes
}

func boolToOnOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}
