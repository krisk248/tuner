package detect

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/krisk248/tuner/internal/output"
	"github.com/krisk248/tuner/internal/sysfs"
)

// WifiInfo holds wireless link details from iw.
type WifiInfo struct {
	SSID      string
	Frequency int    // MHz
	Band      string // "2.4 GHz", "5 GHz", "6 GHz"
	Signal    int    // dBm (negative)
	Quality   string // excellent, good, fair, weak
	RxBitrate string
	TxBitrate string
}

// NICOffloads holds ethtool offload status.
type NICOffloads struct {
	TxChecksum bool
	RxChecksum bool
	TSO        bool // tcp-segmentation-offload
	GSO        bool // generic-segmentation-offload
	GRO        bool // generic-receive-offload
}

// NetInterface holds per-interface info.
type NetInterface struct {
	Name       string
	Speed      int    // Mbps, -1 if unknown
	State      string // up, down, unknown
	Driver     string
	MTU        int
	MAC        string
	IsWireless bool
	IsVirtual  bool
	Wifi       *WifiInfo
	Offloads   *NICOffloads
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

		// Wi-Fi details via iw
		if iface.IsWireless && iface.State == "up" {
			iface.Wifi = detectWifi(name)
		}

		// NIC offloads via ethtool
		if !iface.IsVirtual {
			iface.Offloads = detectOffloads(name)
		}

		info.Interfaces = append(info.Interfaces, iface)
	}

	return info
}

// NetworkSection formats network info as an output section.
// mode controls profile-aware filtering (e.g. server skips Wi-Fi details).
func NetworkSection(info NetworkInfo, mode output.DiagMode) output.Section {
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
		output.Field{Key: "Recv Buffer Max", Value: FormatBytes(info.RmemMax), Status: output.StatusInfo},
		output.Field{Key: "Send Buffer Max", Value: FormatBytes(info.WmemMax), Status: output.StatusInfo},
	)

	// Interfaces
	for _, iface := range info.Interfaces {
		if iface.IsVirtual {
			continue
		}
		// Server mode: skip wireless interfaces entirely
		if iface.IsWireless && mode == output.ModeServer {
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

		// Wi-Fi details (skip in server mode)
		if iface.Wifi != nil && mode != output.ModeServer {
			w := iface.Wifi
			sec.Fields = append(sec.Fields,
				output.Field{Key: "  SSID", Value: w.SSID, Status: output.StatusInfo},
				output.Field{Key: "  Frequency", Value: formatFreqBand(w.Frequency), Status: output.StatusInfo},
				output.Field{Key: "  Signal", Value: fmt.Sprintf("%d dBm (%s)", w.Signal, w.Quality), Status: signalStatus(w.Quality)},
			)
			if w.TxBitrate != "" {
				sec.Fields = append(sec.Fields,
					output.Field{Key: "  TX Bitrate", Value: w.TxBitrate, Status: output.StatusInfo},
				)
			}
			if w.RxBitrate != "" {
				sec.Fields = append(sec.Fields,
					output.Field{Key: "  RX Bitrate", Value: w.RxBitrate, Status: output.StatusInfo},
				)
			}
		}

		// NIC Offloads
		if iface.Offloads != nil {
			o := iface.Offloads
			sec.Fields = append(sec.Fields,
				output.Field{Key: "  TX Checksum", Value: offloadStr(o.TxChecksum), Status: offloadStatus(o.TxChecksum)},
				output.Field{Key: "  RX Checksum", Value: offloadStr(o.RxChecksum), Status: offloadStatus(o.RxChecksum)},
				output.Field{Key: "  TSO", Value: offloadStr(o.TSO), Status: offloadStatus(o.TSO)},
				output.Field{Key: "  GSO", Value: offloadStr(o.GSO), Status: offloadStatus(o.GSO)},
				output.Field{Key: "  GRO", Value: offloadStr(o.GRO), Status: offloadStatus(o.GRO)},
			)
		}
	}

	return sec
}

func FormatBytes(b int) string {
	switch {
	case b >= 1<<20:
		return fmt.Sprintf("%d MB", b/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%d KB", b/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func detectWifi(ifname string) *WifiInfo {
	out, err := exec.Command("iw", "dev", ifname, "link").Output()
	if err != nil {
		return nil
	}

	w := &WifiInfo{}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "SSID:"):
			w.SSID = strings.TrimSpace(strings.TrimPrefix(line, "SSID:"))
		case strings.HasPrefix(line, "freq:"):
			fmt.Sscanf(line, "freq: %d", &w.Frequency)
		case strings.HasPrefix(line, "signal:"):
			fmt.Sscanf(line, "signal: %d dBm", &w.Signal)
		case strings.HasPrefix(line, "rx bitrate:"):
			w.RxBitrate = strings.TrimSpace(strings.TrimPrefix(line, "rx bitrate:"))
		case strings.HasPrefix(line, "tx bitrate:"):
			w.TxBitrate = strings.TrimSpace(strings.TrimPrefix(line, "tx bitrate:"))
		}
	}

	// Determine band
	switch {
	case w.Frequency >= 2400 && w.Frequency <= 2500:
		w.Band = "2.4 GHz"
	case w.Frequency >= 5150 && w.Frequency <= 5900:
		w.Band = "5 GHz"
	case w.Frequency >= 5925:
		w.Band = "6 GHz"
	}

	// Signal quality
	switch {
	case w.Signal >= -50:
		w.Quality = "excellent"
	case w.Signal >= -60:
		w.Quality = "good"
	case w.Signal >= -70:
		w.Quality = "fair"
	default:
		w.Quality = "weak"
	}

	if w.SSID == "" {
		return nil
	}
	return w
}

func detectOffloads(ifname string) *NICOffloads {
	out, err := exec.Command("ethtool", "-k", ifname).Output()
	if err != nil {
		return nil
	}

	o := &NICOffloads{}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		on := strings.HasPrefix(val, "on")

		switch key {
		case "tx-checksumming":
			o.TxChecksum = on
		case "rx-checksumming":
			o.RxChecksum = on
		case "tcp-segmentation-offload":
			o.TSO = on
		case "generic-segmentation-offload":
			o.GSO = on
		case "generic-receive-offload":
			o.GRO = on
		}
	}
	return o
}

func offloadStr(on bool) string {
	if on {
		return "on"
	}
	return "off"
}

func offloadStatus(on bool) output.Status {
	if on {
		return output.StatusGood
	}
	return output.StatusWarn
}

func signalStatus(quality string) output.Status {
	switch quality {
	case "excellent", "good":
		return output.StatusGood
	case "fair":
		return output.StatusWarn
	default:
		return output.StatusBad
	}
}

func formatFreqBand(freq int) string {
	if freq == 0 {
		return "unknown"
	}
	mhz := strconv.Itoa(freq)
	switch {
	case freq >= 2400 && freq <= 2500:
		return mhz + " MHz (2.4 GHz)"
	case freq >= 5150 && freq <= 5900:
		return mhz + " MHz (5 GHz)"
	case freq >= 5925:
		return mhz + " MHz (6 GHz)"
	default:
		return mhz + " MHz"
	}
}
