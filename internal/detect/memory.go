package detect

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/krisk248/tuner/internal/output"
	"github.com/krisk248/tuner/internal/sysfs"
)

// ZramInfo holds info about a single zram device.
type ZramInfo struct {
	Name         string
	DiskSizeGB   float64
	Algorithm    string
	OrigDataSize int64 // bytes
	ComprSize    int64 // bytes
}

// MemoryInfo holds memory diagnostic data.
type MemoryInfo struct {
	TotalKB     int64
	AvailableKB int64
	SwapTotalKB int64
	SwapFreeKB  int64
	Swappiness  int
	DirtyBgRatio int
	DirtyRatio  int
	DirtyExpire int
	DirtyWriteback int
	VFSCachePressure int
	THPEnabled  string
	THPDefrag   string
	ZswapEnabled bool
	ZswapCompressor string
	ZswapMaxPool int
	ZramDevices []ZramInfo
	HugePages   int
}

// DetectMemory gathers memory information.
func DetectMemory() MemoryInfo {
	info := MemoryInfo{}

	// Parse /proc/meminfo
	if lines, err := sysfs.ReadLines(sysfs.ProcMemInfo); err == nil {
		for _, line := range lines {
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			key := strings.TrimSuffix(parts[0], ":")
			val, _ := strconv.ParseInt(parts[1], 10, 64)

			switch key {
			case "MemTotal":
				info.TotalKB = val
			case "MemAvailable":
				info.AvailableKB = val
			case "SwapTotal":
				info.SwapTotalKB = val
			case "SwapFree":
				info.SwapFreeKB = val
			case "HugePages_Total":
				info.HugePages = int(val)
			}
		}
	}

	// VM parameters
	if v, err := sysfs.ReadInt(sysfs.VMSwappiness); err == nil {
		info.Swappiness = v
	}
	if v, err := sysfs.ReadInt(sysfs.VMDirtyBgRatio); err == nil {
		info.DirtyBgRatio = v
	}
	if v, err := sysfs.ReadInt(sysfs.VMDirtyRatio); err == nil {
		info.DirtyRatio = v
	}
	if v, err := sysfs.ReadInt(sysfs.VMDirtyExpire); err == nil {
		info.DirtyExpire = v
	}
	if v, err := sysfs.ReadInt(sysfs.VMDirtyWriteback); err == nil {
		info.DirtyWriteback = v
	}
	if v, err := sysfs.ReadInt(sysfs.VMVFSCachePressure); err == nil {
		info.VFSCachePressure = v
	}

	// THP
	if v, err := sysfs.ReadBracketedValue(sysfs.THPEnabled); err == nil {
		info.THPEnabled = v
	}
	if v, err := sysfs.ReadBracketedValue(sysfs.THPDefrag); err == nil {
		info.THPDefrag = v
	}

	// Zswap
	if v, err := sysfs.ReadBool(sysfs.ZswapEnabled); err == nil {
		info.ZswapEnabled = v
	}
	if v, err := sysfs.ReadString(sysfs.ZswapCompressor); err == nil {
		info.ZswapCompressor = v
	}
	if v, err := sysfs.ReadInt(sysfs.ZswapMaxPool); err == nil {
		info.ZswapMaxPool = v
	}

	// Zram devices
	entries, _ := os.ReadDir(sysfs.BlockBase)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "zram") {
			zram := ZramInfo{Name: e.Name()}
			base := filepath.Join(sysfs.BlockBase, e.Name())

			if v, err := sysfs.ReadInt64(filepath.Join(base, "disksize")); err == nil {
				zram.DiskSizeGB = float64(v) / (1024 * 1024 * 1024)
			}
			if v, err := sysfs.ReadString(filepath.Join(base, "comp_algorithm")); err == nil {
				// Format: "lzo lzo-rle [zstd] lz4" - extract active one
				if bracketed, err := sysfs.ReadBracketedValue(filepath.Join(base, "comp_algorithm")); err == nil {
					zram.Algorithm = bracketed
				} else {
					zram.Algorithm = v
				}
			}
			// mm_stat: orig_data_size compr_data_size mem_used_total ...
			if stat, err := sysfs.ReadString(filepath.Join(base, "mm_stat")); err == nil {
				fields := strings.Fields(stat)
				if len(fields) >= 2 {
					zram.OrigDataSize, _ = strconv.ParseInt(fields[0], 10, 64)
					zram.ComprSize, _ = strconv.ParseInt(fields[1], 10, 64)
				}
			}

			info.ZramDevices = append(info.ZramDevices, zram)
		}
	}

	return info
}

// MemorySection formats memory info as an output section.
func MemorySection(info MemoryInfo) output.Section {
	sec := output.Section{Title: "Memory"}

	totalGB := float64(info.TotalKB) / 1048576.0
	availGB := float64(info.AvailableKB) / 1048576.0
	usedPct := 0.0
	if info.TotalKB > 0 {
		usedPct = float64(info.TotalKB-info.AvailableKB) / float64(info.TotalKB) * 100
	}

	memStatus := output.StatusGood
	if usedPct > 90 {
		memStatus = output.StatusBad
	} else if usedPct > 75 {
		memStatus = output.StatusWarn
	}

	sec.Fields = append(sec.Fields,
		output.Field{Key: "Total RAM", Value: fmt.Sprintf("%.1f GB", totalGB), Status: output.StatusInfo},
		output.Field{Key: "Available", Value: fmt.Sprintf("%.1f GB (%.0f%% used)", availGB, usedPct), Status: memStatus},
	)

	// Swap
	if info.SwapTotalKB > 0 {
		swapGB := float64(info.SwapTotalKB) / 1048576.0
		swapUsedPct := float64(info.SwapTotalKB-info.SwapFreeKB) / float64(info.SwapTotalKB) * 100
		swapStatus := output.StatusGood
		if swapUsedPct > 50 {
			swapStatus = output.StatusWarn
		}
		sec.Fields = append(sec.Fields,
			output.Field{Key: "Swap", Value: fmt.Sprintf("%.1f GB (%.0f%% used)", swapGB, swapUsedPct), Status: swapStatus},
		)
	} else {
		sec.Fields = append(sec.Fields,
			output.Field{Key: "Swap", Value: "none", Status: output.StatusInfo},
		)
	}

	// VM tunables
	swapStatus := output.StatusGood
	if info.Swappiness > 30 {
		swapStatus = output.StatusWarn
	}
	sec.Fields = append(sec.Fields,
		output.Field{Key: "Swappiness", Value: fmt.Sprintf("%d", info.Swappiness), Status: swapStatus},
		output.Field{Key: "Dirty BG Ratio", Value: fmt.Sprintf("%d%%", info.DirtyBgRatio), Status: output.StatusInfo},
		output.Field{Key: "Dirty Ratio", Value: fmt.Sprintf("%d%%", info.DirtyRatio), Status: output.StatusInfo},
		output.Field{Key: "VFS Cache Pressure", Value: fmt.Sprintf("%d", info.VFSCachePressure), Status: output.StatusInfo},
	)

	// THP
	if info.THPEnabled != "" {
		thpStatus := output.StatusInfo
		if info.THPEnabled == "always" {
			thpStatus = output.StatusWarn
		}
		sec.Fields = append(sec.Fields,
			output.Field{Key: "THP", Value: info.THPEnabled, Status: thpStatus},
		)
	}

	// Zswap
	if info.ZswapCompressor != "" {
		zswapStr := "disabled"
		if info.ZswapEnabled {
			zswapStr = fmt.Sprintf("enabled (%s, %d%%)", info.ZswapCompressor, info.ZswapMaxPool)
		}
		sec.Fields = append(sec.Fields,
			output.Field{Key: "Zswap", Value: zswapStr, Status: output.StatusInfo},
		)
	}

	// Zram
	for _, z := range info.ZramDevices {
		val := fmt.Sprintf("%.1f GB", z.DiskSizeGB)
		if z.Algorithm != "" {
			val = fmt.Sprintf("%.1f GB (%s", z.DiskSizeGB, z.Algorithm)
			if z.ComprSize > 0 {
				ratio := float64(z.OrigDataSize) / float64(z.ComprSize)
				val = fmt.Sprintf("%s, %.1f:1 ratio)", val, ratio)
			} else {
				val += ")"
			}
		}
		sec.Fields = append(sec.Fields,
			output.Field{Key: z.Name, Value: val, Status: output.StatusInfo},
		)
	}

	return sec
}
