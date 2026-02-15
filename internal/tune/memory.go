package tune

import (
	"fmt"

	"github.com/krisk248/tuner/internal/profile"
	"github.com/krisk248/tuner/internal/sysfs"
)

func computeMemoryChanges(v profile.Values) []Change {
	var changes []Change

	// Swappiness
	if cur, err := sysfs.ReadInt(sysfs.VMSwappiness); err == nil && cur != v.Swappiness {
		target := v.Swappiness
		changes = append(changes, Change{
			Subsystem: "memory",
			Parameter: "Swappiness",
			OldValue:  fmt.Sprintf("%d", cur),
			NewValue:  fmt.Sprintf("%d", target),
			Path:      sysfs.VMSwappiness,
			ApplyFunc: func() error {
				return sysfs.WriteInt(sysfs.VMSwappiness, target)
			},
		})
	}

	// Dirty background ratio
	if cur, err := sysfs.ReadInt(sysfs.VMDirtyBgRatio); err == nil && cur != v.DirtyBgRatio {
		target := v.DirtyBgRatio
		changes = append(changes, Change{
			Subsystem: "memory",
			Parameter: "Dirty BG Ratio",
			OldValue:  fmt.Sprintf("%d", cur),
			NewValue:  fmt.Sprintf("%d", target),
			Path:      sysfs.VMDirtyBgRatio,
			ApplyFunc: func() error {
				return sysfs.WriteInt(sysfs.VMDirtyBgRatio, target)
			},
		})
	}

	// Dirty ratio
	if cur, err := sysfs.ReadInt(sysfs.VMDirtyRatio); err == nil && cur != v.DirtyRatio {
		target := v.DirtyRatio
		changes = append(changes, Change{
			Subsystem: "memory",
			Parameter: "Dirty Ratio",
			OldValue:  fmt.Sprintf("%d", cur),
			NewValue:  fmt.Sprintf("%d", target),
			Path:      sysfs.VMDirtyRatio,
			ApplyFunc: func() error {
				return sysfs.WriteInt(sysfs.VMDirtyRatio, target)
			},
		})
	}

	// Dirty expire
	if cur, err := sysfs.ReadInt(sysfs.VMDirtyExpire); err == nil && cur != v.DirtyExpire {
		target := v.DirtyExpire
		changes = append(changes, Change{
			Subsystem: "memory",
			Parameter: "Dirty Expire",
			OldValue:  fmt.Sprintf("%d", cur),
			NewValue:  fmt.Sprintf("%d", target),
			Path:      sysfs.VMDirtyExpire,
			ApplyFunc: func() error {
				return sysfs.WriteInt(sysfs.VMDirtyExpire, target)
			},
		})
	}

	// Dirty writeback
	if cur, err := sysfs.ReadInt(sysfs.VMDirtyWriteback); err == nil && cur != v.DirtyWriteback {
		target := v.DirtyWriteback
		changes = append(changes, Change{
			Subsystem: "memory",
			Parameter: "Dirty Writeback",
			OldValue:  fmt.Sprintf("%d", cur),
			NewValue:  fmt.Sprintf("%d", target),
			Path:      sysfs.VMDirtyWriteback,
			ApplyFunc: func() error {
				return sysfs.WriteInt(sysfs.VMDirtyWriteback, target)
			},
		})
	}

	// VFS cache pressure
	if cur, err := sysfs.ReadInt(sysfs.VMVFSCachePressure); err == nil && cur != v.VFSCachePressure {
		target := v.VFSCachePressure
		changes = append(changes, Change{
			Subsystem: "memory",
			Parameter: "VFS Cache Pressure",
			OldValue:  fmt.Sprintf("%d", cur),
			NewValue:  fmt.Sprintf("%d", target),
			Path:      sysfs.VMVFSCachePressure,
			ApplyFunc: func() error {
				return sysfs.WriteInt(sysfs.VMVFSCachePressure, target)
			},
		})
	}

	// THP
	if sysfs.Exists(sysfs.THPEnabled) {
		if cur, err := sysfs.ReadBracketedValue(sysfs.THPEnabled); err == nil && cur != v.THPEnabled {
			target := v.THPEnabled
			changes = append(changes, Change{
				Subsystem: "memory",
				Parameter: "THP",
				OldValue:  cur,
				NewValue:  target,
				Path:      sysfs.THPEnabled,
				ApplyFunc: func() error {
					return sysfs.WriteString(sysfs.THPEnabled, target)
				},
			})
		}
	}

	return changes
}
