package tune

import (
	"fmt"
	"path/filepath"

	"github.com/krisk248/tuner/internal/detect"
	"github.com/krisk248/tuner/internal/profile"
	"github.com/krisk248/tuner/internal/sysfs"
)

func computeStorageChanges(v profile.Values) []Change {
	var changes []Change

	storageInfo := detect.DetectStorage()

	for _, disk := range storageInfo.Disks {
		recommended := v.IOScheduler(disk.Type)

		// Scheduler
		if disk.Scheduler != recommended {
			schedPath := filepath.Join(sysfs.BlockBase, disk.Name, "queue/scheduler")
			target := recommended
			old := disk.Scheduler
			changes = append(changes, Change{
				Subsystem: "storage",
				Parameter: fmt.Sprintf("%s scheduler", disk.Name),
				OldValue:  old,
				NewValue:  target,
				Path:      schedPath,
				ApplyFunc: func() error {
					return sysfs.WriteString(schedPath, target)
				},
			})
		}

		// Read ahead
		if v.ReadAhead > 0 && disk.ReadAhead != v.ReadAhead {
			raPath := filepath.Join(sysfs.BlockBase, disk.Name, "queue/read_ahead_kb")
			target := v.ReadAhead
			old := disk.ReadAhead
			changes = append(changes, Change{
				Subsystem: "storage",
				Parameter: fmt.Sprintf("%s read_ahead_kb", disk.Name),
				OldValue:  fmt.Sprintf("%d", old),
				NewValue:  fmt.Sprintf("%d", target),
				Path:      raPath,
				ApplyFunc: func() error {
					return sysfs.WriteInt(raPath, target)
				},
			})
		}
	}

	return changes
}
