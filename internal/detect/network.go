package detect

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/krisk248/tuner/internal/output"
	"github.com/krisk248/tuner/internal/sysfs"
)

// NetInterface holds per-interface info.
type NetInterface struct {
	Name      string
	Speed     int    // Mbps, -1 if unknown
	State     string // up, down, unknown
	Driver    string
	MTU       int
	MAC       string
	IsWireless bool
	IsVirtual  bool
}

// NetworkInfo holds network diagnostic data.
type NetworkInfo struct {
	Interfaces    []NetInterface
	TCPCongestion string
	AvailCong     []string
	TCPFastOpen   int
	TCPMTUProbing int
	RmemMax       int
	WmemMax       int
	TCPRmem       string
	TCPWmem       string
}

// DetectNetwork gathers network information.
func DetectNetwork() NetworkInfo {
	info := NetworkInfo{}

	// TCP parameters
	if v, err := sysfs.ReadString(sysfs.TCPCongestion); err == nil {
		info.TCPCongestion = v
	}
	if v, err := sysfs.ReadFields(sysfs.TCPAvailCong); err == nil {
		info.AvailCong = v
	}
	if v, err := sysfs.ReadInt(sysfs.TCPFastOpen); err == nil {
		info.TCPFastOpen = v
	}
	if v, err := sysfs.ReadInt(sysfs.TCPMTUProbing); err == nil {
		info.TCPMTUProbing = v
	}
	if v, err := sysfs.ReadInt(sysfs.NetCoreBufMax); err == nil {
		info.RmemMax = v
	}
	if v, err := sysfs.ReadInt(sysfs.NetCoreWBufMax); err == nil {
		info.WmemMax = v
	}
	if v, err := sysfs.ReadString(sysfs.TCPRmem); err == nil {
		info.TCPRmem = v
	}
	if v, err := sysfs.ReadString(sysfs.TCPWmem); err == nil {
		info.TCPWmem = v
	}

	// Network interfaces
	entries, err := os.ReadDir(sysfs.NetBase)
	if err != nil {
		return info
	}

	for _, e := range entries {
		name := e.Name()
		if name == "lo" {
			continue
		}

		iface := NetInterface{Name: name, Speed: -1}
		base := filepath.Join(sysfs.NetBase, name)

		// Operational state
		if v, err := sysfs.ReadString(filepath.Join(base, "operstate")); err == nil {
			iface.State = v
		}

		// Speed
		if v, err := sysfs.ReadInt(filepath.Join(base, "speed")); err == nil {
			iface.Speed = v
		}

		// MTU
		if v, err := sysfs.ReadInt(filepath.Join(base, "mtu")); err == nil {
			iface.MTU = v
		}

		// MAC
		if v, err := sysfs.ReadString(filepath.Join(base, "address")); err == nil {
			iface.MAC = v
		}

		// Wireless detection
		iface.IsWireless = sysfs.Exists(filepath.Join(base, "wireless")) ||
			sysfs.Exists(filepath.Join(base, "phy80211"))

		// Virtual detection
		link, err := os.Readlink(filepath.Join(base, "device"))
		if err != nil {
			iface.IsVirtual = true
		} else {
			iface.IsVirtual = strings.Contains(link, "virtual")
		}

		info.Interfaces = append(info.Interfaces, iface)
	}

	return info
}

// NetworkSection formats network info as an output section.
func NetworkSection(info NetworkInfo) output.Section {
	sec := output.Section{Title: "Network"}

	// TCP settings
	congStatus := output.StatusInfo
	if info.TCPCongestion == "bbr" {
		congStatus = output.StatusGood
	} else if info.TCPCongestion == "cubic" {
		congStatus = output.StatusWarn
	}
	sec.Fields = append(sec.Fields,
		output.Field{Key: "TCP Congestion", Value: info.TCPCongestion, Status: congStatus},
		output.Field{Key: "TCP Fast Open", Value: fmt.Sprintf("%d", info.TCPFastOpen), Status: output.StatusInfo},
		output.Field{Key: "MTU Probing", Value: fmt.Sprintf("%d", info.TCPMTUProbing), Status: output.StatusInfo},
		output.Field{Key: "Recv Buffer Max", Value: formatBytes(info.RmemMax), Status: output.StatusInfo},
		output.Field{Key: "Send Buffer Max", Value: formatBytes(info.WmemMax), Status: output.StatusInfo},
	)

	// Interfaces
	for _, iface := range info.Interfaces {
		if iface.IsVirtual {
			continue
		}
		ifType := "ethernet"
		if iface.IsWireless {
			ifType = "wireless"
		}

		speedStr := "unknown"
		if iface.Speed > 0 {
			speedStr = fmt.Sprintf("%d Mbps", iface.Speed)
		}

		status := output.StatusInfo
		if iface.State == "up" {
			status = output.StatusGood
		} else if iface.State == "down" {
			status = output.StatusWarn
		}

		sec.Fields = append(sec.Fields,
			output.Field{
				Key:    fmt.Sprintf("%s (%s)", iface.Name, ifType),
				Value:  fmt.Sprintf("%s, MTU %d, %s", iface.State, iface.MTU, speedStr),
				Status: status,
			},
		)
	}

	return sec
}

func formatBytes(b int) string {
	switch {
	case b >= 1<<20:
		return fmt.Sprintf("%d MB", b/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%d KB", b/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
