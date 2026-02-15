package detect

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/krisk248/tuner/internal/output"
	"github.com/krisk248/tuner/internal/sysfs"
)

// GPUInfo holds GPU diagnostic data.
type GPUInfo struct {
	Cards []GPUCard
}

// GPUCard holds info about a single GPU.
type GPUCard struct {
	Name         string
	Driver       string
	Vendor       string
	PowerProfile string // auto, high, low (amdgpu)
	MemTotalMB   int
	MemUsedMB    int
}

// DetectGPU gathers GPU information from /sys/class/drm/.
func DetectGPU() GPUInfo {
	info := GPUInfo{}

	entries, err := os.ReadDir(sysfs.DRMBase)
	if err != nil {
		return info
	}

	for _, e := range entries {
		name := e.Name()
		// Only look at card0, card1, etc. (not card0-DP-1 etc.)
		if !strings.HasPrefix(name, "card") || strings.Contains(name, "-") {
			continue
		}

		base := filepath.Join(sysfs.DRMBase, name)
		deviceBase := filepath.Join(base, "device")

		card := GPUCard{Name: name}

		// Driver via symlink
		driverLink, err := os.Readlink(filepath.Join(deviceBase, "driver"))
		if err == nil {
			card.Driver = filepath.Base(driverLink)
		}

		// Vendor (PCI)
		if v, err := sysfs.ReadString(filepath.Join(deviceBase, "vendor")); err == nil {
			switch v {
			case "0x8086":
				card.Vendor = "Intel"
			case "0x1002":
				card.Vendor = "AMD"
			case "0x10de":
				card.Vendor = "NVIDIA"
			default:
				card.Vendor = v
			}
		}

		// AMDGPU power profile
		if v, err := sysfs.ReadString(filepath.Join(deviceBase, "power_dpm_force_performance_level")); err == nil {
			card.PowerProfile = v
		}

		// VRAM (amdgpu/i915)
		if v, err := sysfs.ReadInt64(filepath.Join(deviceBase, "mem_info_vram_total")); err == nil {
			card.MemTotalMB = int(v / (1024 * 1024))
		}
		if v, err := sysfs.ReadInt64(filepath.Join(deviceBase, "mem_info_vram_used")); err == nil {
			card.MemUsedMB = int(v / (1024 * 1024))
		}

		info.Cards = append(info.Cards, card)
	}

	return info
}

// GPUSection formats GPU info as an output section.
func GPUSection(info GPUInfo) output.Section {
	sec := output.Section{Title: "GPU"}

	if len(info.Cards) == 0 {
		sec.Fields = append(sec.Fields,
			output.Field{Key: "GPU", Value: "none detected", Status: output.StatusInfo},
		)
		return sec
	}

	for _, card := range info.Cards {
		sec.Fields = append(sec.Fields,
			output.Field{Key: card.Name, Value: fmt.Sprintf("%s (%s)", card.Vendor, card.Driver), Status: output.StatusInfo},
		)

		if card.MemTotalMB > 0 {
			sec.Fields = append(sec.Fields,
				output.Field{Key: "  VRAM", Value: fmt.Sprintf("%d MB (%d MB used)", card.MemTotalMB, card.MemUsedMB), Status: output.StatusInfo},
			)
		}

		if card.PowerProfile != "" {
			sec.Fields = append(sec.Fields,
				output.Field{Key: "  Power Profile", Value: card.PowerProfile, Status: output.StatusInfo},
			)
		}
	}

	return sec
}
