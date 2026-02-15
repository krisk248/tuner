package tune

import (
	"fmt"

	"github.com/krisk248/tuner/internal/profile"
	"github.com/krisk248/tuner/internal/sysfs"
)

func computeNetworkChanges(v profile.Values) []Change {
	var changes []Change

	// TCP congestion control
	if cur, err := sysfs.ReadString(sysfs.TCPCongestion); err == nil && cur != v.TCPCongestion {
		target := v.TCPCongestion
		changes = append(changes, Change{
			Subsystem: "network",
			Parameter: "TCP Congestion",
			OldValue:  cur,
			NewValue:  target,
			Path:      sysfs.TCPCongestion,
			ApplyFunc: func() error {
				return sysfs.WriteString(sysfs.TCPCongestion, target)
			},
		})
	}

	// TCP Fast Open
	if cur, err := sysfs.ReadInt(sysfs.TCPFastOpen); err == nil && cur != v.TCPFastOpen {
		target := v.TCPFastOpen
		changes = append(changes, Change{
			Subsystem: "network",
			Parameter: "TCP Fast Open",
			OldValue:  fmt.Sprintf("%d", cur),
			NewValue:  fmt.Sprintf("%d", target),
			Path:      sysfs.TCPFastOpen,
			ApplyFunc: func() error {
				return sysfs.WriteInt(sysfs.TCPFastOpen, target)
			},
		})
	}

	// MTU probing
	if cur, err := sysfs.ReadInt(sysfs.TCPMTUProbing); err == nil && cur != v.TCPMTUProbing {
		target := v.TCPMTUProbing
		changes = append(changes, Change{
			Subsystem: "network",
			Parameter: "TCP MTU Probing",
			OldValue:  fmt.Sprintf("%d", cur),
			NewValue:  fmt.Sprintf("%d", target),
			Path:      sysfs.TCPMTUProbing,
			ApplyFunc: func() error {
				return sysfs.WriteInt(sysfs.TCPMTUProbing, target)
			},
		})
	}

	// Receive buffer max
	if cur, err := sysfs.ReadInt(sysfs.NetCoreBufMax); err == nil && cur != v.RmemMax {
		target := v.RmemMax
		changes = append(changes, Change{
			Subsystem: "network",
			Parameter: "Recv Buffer Max",
			OldValue:  fmt.Sprintf("%d", cur),
			NewValue:  fmt.Sprintf("%d", target),
			Path:      sysfs.NetCoreBufMax,
			ApplyFunc: func() error {
				return sysfs.WriteInt(sysfs.NetCoreBufMax, target)
			},
		})
	}

	// Send buffer max
	if cur, err := sysfs.ReadInt(sysfs.NetCoreWBufMax); err == nil && cur != v.WmemMax {
		target := v.WmemMax
		changes = append(changes, Change{
			Subsystem: "network",
			Parameter: "Send Buffer Max",
			OldValue:  fmt.Sprintf("%d", cur),
			NewValue:  fmt.Sprintf("%d", target),
			Path:      sysfs.NetCoreWBufMax,
			ApplyFunc: func() error {
				return sysfs.WriteInt(sysfs.NetCoreWBufMax, target)
			},
		})
	}

	// TCP rmem
	if cur, err := sysfs.ReadString(sysfs.TCPRmem); err == nil && cur != v.TCPRmem {
		target := v.TCPRmem
		changes = append(changes, Change{
			Subsystem: "network",
			Parameter: "TCP Rmem",
			OldValue:  cur,
			NewValue:  target,
			Path:      sysfs.TCPRmem,
			ApplyFunc: func() error {
				return sysfs.WriteString(sysfs.TCPRmem, target)
			},
		})
	}

	// TCP wmem
	if cur, err := sysfs.ReadString(sysfs.TCPWmem); err == nil && cur != v.TCPWmem {
		target := v.TCPWmem
		changes = append(changes, Change{
			Subsystem: "network",
			Parameter: "TCP Wmem",
			OldValue:  cur,
			NewValue:  target,
			Path:      sysfs.TCPWmem,
			ApplyFunc: func() error {
				return sysfs.WriteString(sysfs.TCPWmem, target)
			},
		})
	}

	return changes
}
