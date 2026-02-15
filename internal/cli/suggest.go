package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/krisk248/tuner/internal/detect"
	"github.com/krisk248/tuner/internal/output"
	"github.com/krisk248/tuner/internal/profile"
	"github.com/krisk248/tuner/internal/sysfs"
	"github.com/spf13/cobra"
)

var suggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Show recommended changes for a profile",
	RunE:  runSuggest,
}

var suggestProfile string

func init() {
	suggestCmd.Flags().StringVar(&suggestProfile, "profile", "", "profile to suggest for (server, desktop, laptop)")
	rootCmd.AddCommand(suggestCmd)
}

func runSuggest(cmd *cobra.Command, args []string) error {
	var p profile.Profile
	if suggestProfile != "" {
		p = profile.ForType(profile.Type(suggestProfile))
	} else {
		p = profile.AutoDetect()
		bold := color.New(color.Bold)
		bold.Printf("Auto-detected profile: %s\n\n", p.Type)
	}

	sections := buildSuggestions(p)

	if len(sections) == 0 {
		fmt.Println("System is already optimally configured for this profile.")
		return nil
	}

	formatter := output.NewFormatter(outFormat, noColor)
	return formatter.Format(os.Stdout, sections)
}

func buildSuggestions(p profile.Profile) []output.Section {
	var sections []output.Section

	// CPU suggestions
	cpuSec := suggestCPU(p)
	if len(cpuSec.Fields) > 0 {
		sections = append(sections, cpuSec)
	}

	// Memory suggestions
	memSec := suggestMemory(p)
	if len(memSec.Fields) > 0 {
		sections = append(sections, memSec)
	}

	// Storage suggestions
	storageSec := suggestStorage(p)
	if len(storageSec.Fields) > 0 {
		sections = append(sections, storageSec)
	}

	// Network suggestions
	netSec := suggestNetwork(p)
	if len(netSec.Fields) > 0 {
		sections = append(sections, netSec)
	}

	return sections
}

func suggestCPU(p profile.Profile) output.Section {
	sec := output.Section{Title: "CPU Changes"}
	v := p.Values

	if gov, err := sysfs.ReadString(sysfs.CPUGovernor); err == nil && gov != v.Governor {
		sec.Fields = append(sec.Fields, diffField("Governor", gov, v.Governor))
	}

	if sysfs.Exists(sysfs.CPUEPP) {
		if epp, err := sysfs.ReadString(sysfs.CPUEPP); err == nil && epp != v.EPP {
			sec.Fields = append(sec.Fields, diffField("Energy Perf Pref", epp, v.EPP))
		}
	}

	// Turbo
	turboOn := false
	if sysfs.Exists(sysfs.IntelNoTurbo) {
		if val, err := sysfs.ReadInt(sysfs.IntelNoTurbo); err == nil {
			turboOn = val == 0
		}
	} else if sysfs.Exists(sysfs.CPUBoost) {
		if val, err := sysfs.ReadInt(sysfs.CPUBoost); err == nil {
			turboOn = val == 1
		}
	}
	if turboOn != v.TurboOn {
		cur := "off"
		rec := "off"
		if turboOn {
			cur = "on"
		}
		if v.TurboOn {
			rec = "on"
		}
		sec.Fields = append(sec.Fields, diffField("Turbo Boost", cur, rec))
	}

	return sec
}

func suggestMemory(p profile.Profile) output.Section {
	sec := output.Section{Title: "Memory Changes"}
	v := p.Values

	if cur, err := sysfs.ReadInt(sysfs.VMSwappiness); err == nil && cur != v.Swappiness {
		sec.Fields = append(sec.Fields, diffField("Swappiness", fmt.Sprintf("%d", cur), fmt.Sprintf("%d", v.Swappiness)))
	}
	if cur, err := sysfs.ReadInt(sysfs.VMDirtyBgRatio); err == nil && cur != v.DirtyBgRatio {
		sec.Fields = append(sec.Fields, diffField("Dirty BG Ratio", fmt.Sprintf("%d", cur), fmt.Sprintf("%d", v.DirtyBgRatio)))
	}
	if cur, err := sysfs.ReadInt(sysfs.VMDirtyRatio); err == nil && cur != v.DirtyRatio {
		sec.Fields = append(sec.Fields, diffField("Dirty Ratio", fmt.Sprintf("%d", cur), fmt.Sprintf("%d", v.DirtyRatio)))
	}

	if sysfs.Exists(sysfs.THPEnabled) {
		if cur, err := sysfs.ReadBracketedValue(sysfs.THPEnabled); err == nil && cur != v.THPEnabled {
			sec.Fields = append(sec.Fields, diffField("THP", cur, v.THPEnabled))
		}
	}

	return sec
}

func suggestStorage(p profile.Profile) output.Section {
	sec := output.Section{Title: "Storage Changes"}
	v := p.Values

	storageInfo := detect.DetectStorage()
	for _, disk := range storageInfo.Disks {
		recommended := v.IOScheduler(disk.Type)
		if disk.Scheduler != recommended {
			sec.Fields = append(sec.Fields, diffField(
				fmt.Sprintf("%s scheduler", disk.Name),
				disk.Scheduler,
				recommended,
			))
		}
	}

	return sec
}

func suggestNetwork(p profile.Profile) output.Section {
	sec := output.Section{Title: "Network Changes"}
	v := p.Values

	if cur, err := sysfs.ReadString(sysfs.TCPCongestion); err == nil && cur != v.TCPCongestion {
		sec.Fields = append(sec.Fields, diffField("TCP Congestion", cur, v.TCPCongestion))
	}
	if cur, err := sysfs.ReadInt(sysfs.TCPFastOpen); err == nil && cur != v.TCPFastOpen {
		sec.Fields = append(sec.Fields, diffField("TCP Fast Open", fmt.Sprintf("%d", cur), fmt.Sprintf("%d", v.TCPFastOpen)))
	}
	if cur, err := sysfs.ReadInt(sysfs.NetCoreBufMax); err == nil && cur != v.RmemMax {
		sec.Fields = append(sec.Fields, diffField("Recv Buffer Max", formatSize(cur), formatSize(v.RmemMax)))
	}

	return sec
}

func diffField(key, current, recommended string) output.Field {
	return output.Field{
		Key:    key,
		Value:  fmt.Sprintf("%s → %s", current, recommended),
		Status: output.StatusWarn,
	}
}

func formatSize(bytes int) string {
	switch {
	case bytes >= 1<<20:
		return fmt.Sprintf("%d MB", bytes/(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%d KB", bytes/(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
