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
	if sysfs.Exists(sysfs.IntelNoTurbo) {
		if cur, err := sysfs.ReadInt(sysfs.IntelNoTurbo); err == nil {
			curOn := cur == 0
			if curOn != v.TurboOn {
				newVal := "1" // no_turbo=1 means turbo OFF
				if v.TurboOn {
					newVal = "0"
				}
				curStr := "on"
				if !curOn {
					curStr = "off"
				}
				newStr := "off"
				if v.TurboOn {
					newStr = "on"
				}
				changes = append(changes, Change{
					Subsystem: "cpu",
					Parameter: "Turbo Boost",
					OldValue:  curStr,
					NewValue:  newStr,
					Path:      sysfs.IntelNoTurbo,
					ApplyFunc: func() error {
						return sysfs.WriteString(sysfs.IntelNoTurbo, newVal)
					},
				})
			}
		}
	} else if sysfs.Exists(sysfs.CPUBoost) {
		if cur, err := sysfs.ReadInt(sysfs.CPUBoost); err == nil {
			curOn := cur == 1
			if curOn != v.TurboOn {
				newVal := "0"
				if v.TurboOn {
					newVal = "1"
				}
				curStr := "on"
				if !curOn {
					curStr = "off"
				}
				newStr := "off"
				if v.TurboOn {
					newStr = "on"
				}
				changes = append(changes, Change{
					Subsystem: "cpu",
					Parameter: "Turbo Boost",
					OldValue:  curStr,
					NewValue:  newStr,
					Path:      sysfs.CPUBoost,
					ApplyFunc: func() error {
						return sysfs.WriteString(sysfs.CPUBoost, newVal)
					},
				})
			}
		}
	}

	return changes
}
