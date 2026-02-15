package detect

import (
	"fmt"

	"github.com/krisk248/tuner/internal/output"
	"github.com/krisk248/tuner/internal/sysfs"
)

// ServerInfo holds server-specific kernel parameters.
type ServerInfo struct {
	FileMax      int
	Somaxconn    int
	ConntrackMax int    // -1 if not available
	PortRange    string // e.g. "32768\t60999"
	IRQBalance   ServiceState
}

// DetectServer gathers server-specific kernel tuning info.
func DetectServer() ServerInfo {
	info := ServerInfo{ConntrackMax: -1}

	if v, err := sysfs.ReadInt(sysfs.FileMax); err == nil {
		info.FileMax = v
	}
	if v, err := sysfs.ReadInt(sysfs.Somaxconn); err == nil {
		info.Somaxconn = v
	}
	if sysfs.Exists(sysfs.ConntrackMax) {
		if v, err := sysfs.ReadInt(sysfs.ConntrackMax); err == nil {
			info.ConntrackMax = v
		}
	}
	if v, err := sysfs.ReadString(sysfs.PortRange); err == nil {
		info.PortRange = v
	}

	info.IRQBalance = checkService("irqbalance")

	return info
}

// ServerSection formats server info as an output section.
func ServerSection(info ServerInfo) output.Section {
	sec := output.Section{Title: "Server Tuning"}

	sec.Fields = append(sec.Fields,
		output.Field{Key: "File Max", Value: fmt.Sprintf("%d", info.FileMax), Status: fileMaxStatus(info.FileMax)},
		output.Field{Key: "Somaxconn", Value: fmt.Sprintf("%d", info.Somaxconn), Status: somaxconnStatus(info.Somaxconn)},
	)

	if info.ConntrackMax >= 0 {
		sec.Fields = append(sec.Fields,
			output.Field{Key: "Conntrack Max", Value: fmt.Sprintf("%d", info.ConntrackMax), Status: output.StatusInfo},
		)
	}

	if info.PortRange != "" {
		sec.Fields = append(sec.Fields,
			output.Field{Key: "Port Range", Value: info.PortRange, Status: output.StatusInfo},
		)
	}

	sec.Fields = append(sec.Fields,
		output.Field{Key: "IRQ Balance", Value: serviceStr(info.IRQBalance), Status: irqBalanceStatus(info.IRQBalance)},
	)

	return sec
}

func fileMaxStatus(v int) output.Status {
	if v >= 1048576 {
		return output.StatusGood
	}
	return output.StatusWarn
}

func somaxconnStatus(v int) output.Status {
	if v >= 4096 {
		return output.StatusGood
	}
	return output.StatusWarn
}

func irqBalanceStatus(s ServiceState) output.Status {
	if s.Active {
		return output.StatusGood
	}
	if s.Installed {
		return output.StatusWarn
	}
	return output.StatusInfo
}
