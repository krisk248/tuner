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
		bold.Printf("Auto-detected profile: %s\n", p.Type)
		if p.PowerState != "" {
			bold.Printf("Power state: %s\n", p.PowerState)
		}
		fmt.Println()
	}

	powerInfo := detect.DetectPower()

	// Warnings section
	warnings := buildWarnings(powerInfo, p.Type)
	if len(warnings.Fields) > 0 {
		formatter := output.NewFormatter(outFormat, noColor)
		formatter.Format(os.Stdout, []output.Section{warnings})
		fmt.Println()
	}

	// CPU skip logic per profile:
	// Laptop: skip if TLP active (TLP manages governor/EPP)
	// Desktop: never skip (always plugged in, no power manager expected)
	// Server: skip if tuned active (tuned manages CPU)
	skipCPU := false
	switch p.Type {
	case profile.Laptop:
		skipCPU = powerInfo.TLP.Active
	case profile.Server:
		skipCPU = powerInfo.Tuned.Active
	}

	sections := buildSuggestions(p, skipCPU)

	// Server-specific suggestions
	if p.Type == profile.Server {
		if sec := suggestServer(); len(sec.Fields) > 0 {
			sections = append(sections, sec)
		}
	}

	if len(sections) == 0 {
		fmt.Println("System is already optimally configured for this profile.")
		return nil
	}

	formatter := output.NewFormatter(outFormat, noColor)
	return formatter.Format(os.Stdout, sections)
}

func buildWarnings(power detect.PowerInfo, pType profile.Type) output.Section {
	sec := output.Section{Title: "Warnings"}

	switch pType {
	case profile.Laptop:
		// Laptop expects TLP, ignores tuned
		if !power.TLP.Installed {
			sec.Fields = append(sec.Fields,
				output.Field{
					Key:    "TLP",
					Value:  "not installed - recommended for laptop power management",
					Status: output.StatusWarn,
				},
				output.Field{
					Key:    "",
					Value:  "Install: sudo dnf install tlp && sudo systemctl enable --now tlp",
					Status: output.StatusNone,
				},
			)
		} else if power.TLP.Enabled && !power.TLP.Active {
			sec.Fields = append(sec.Fields,
				output.Field{
					Key:    "TLP",
					Value:  "enabled but not running",
					Status: output.StatusWarn,
				},
				output.Field{
					Key:    "",
					Value:  "Run: sudo systemctl start tlp",
					Status: output.StatusNone,
				},
			)
		}
		if power.TLP.Active {
			sec.Fields = append(sec.Fields,
				output.Field{
					Key:    "TLP",
					Value:  "active - manages CPU governor/EPP, skipping CPU suggestions",
					Status: output.StatusInfo,
				},
			)
		}

	case profile.Desktop:
		// Desktop: no power manager expected, never mention TLP
		// Optionally note tuned if present
		if power.Tuned.Active {
			val := "active"
			if power.TunedProfile != "" {
				val += fmt.Sprintf(" [%s]", power.TunedProfile)
			}
			sec.Fields = append(sec.Fields,
				output.Field{
					Key:    "tuned",
					Value:  val + " - CPU suggestions still shown (desktop always plugged in)",
					Status: output.StatusInfo,
				},
			)
		}

	case profile.Server:
		// Server expects tuned, ignores TLP
		if !power.Tuned.Active {
			sec.Fields = append(sec.Fields,
				output.Field{
					Key:    "tuned",
					Value:  "not active - recommended for server workloads",
					Status: output.StatusWarn,
				},
				output.Field{
					Key:    "",
					Value:  "Run: sudo tuned-adm profile throughput-performance",
					Status: output.StatusNone,
				},
			)
		} else {
			val := "active"
			if power.TunedProfile != "" {
				val += fmt.Sprintf(" [%s]", power.TunedProfile)
			}
			sec.Fields = append(sec.Fields,
				output.Field{
					Key:    "tuned",
					Value:  val + " - manages CPU settings, skipping CPU suggestions",
					Status: output.StatusInfo,
				},
			)
		}
	}

	return sec
}

func buildSuggestions(p profile.Profile, skipCPU bool) []output.Section {
	var sections []output.Section

	if !skipCPU {
		if sec := suggestCPU(p); len(sec.Fields) > 0 {
			sections = append(sections, sec)
		}
	}

	for _, sec := range []output.Section{
		suggestMemory(p),
		suggestStorage(p),
		suggestNetwork(p),
	} {
		if len(sec.Fields) > 0 {
			sections = append(sections, sec)
		}
	}
	return sections
}

// suggestion adds a diff field with reason and benefit lines.
type suggestion struct {
	key     string
	current string
	target  string
	reason  string
	benefit string
}

func (s suggestion) fields() []output.Field {
	fields := []output.Field{
		{
			Key:    s.key,
			Value:  fmt.Sprintf("%s → %s", s.current, s.target),
			Status: output.StatusWarn,
		},
	}
	if s.reason != "" {
		fields = append(fields, output.Field{
			Key:    fmt.Sprintf("  %s", "Reason"),
			Value:  s.reason,
			Status: output.StatusNone,
		})
	}
	if s.benefit != "" {
		fields = append(fields, output.Field{
			Key:    fmt.Sprintf("  %s", "Benefit"),
			Value:  s.benefit,
			Status: output.StatusNone,
		})
	}
	return fields
}

func suggestCPU(p profile.Profile) output.Section {
	sec := output.Section{Title: "CPU Changes"}
	v := p.Values

	if gov, err := sysfs.ReadString(sysfs.CPUGovernor); err == nil && gov != v.Governor {
		s := suggestion{
			key: "Governor", current: gov, target: v.Governor,
		}
		switch {
		case v.Governor == "schedutil":
			s.reason = "Dynamic scaling saves power when CPU is idle"
			s.benefit = "~10-15% battery improvement, similar peak performance"
		case v.Governor == "performance":
			s.reason = "Fixed max frequency for maximum throughput"
			s.benefit = "Lowest latency, best for servers and heavy workloads"
		case v.Governor == "powersave":
			s.reason = "Minimum frequency to extend battery life"
			s.benefit = "Maximum battery savings on battery power"
		}
		sec.Fields = append(sec.Fields, s.fields()...)
	}

	if epp, err := sysfs.ReadString(sysfs.CPUEPP); err == nil && epp != v.EPP {
		s := suggestion{
			key: "Energy Perf Pref", current: epp, target: v.EPP,
		}
		switch {
		case v.EPP == "balance_performance":
			s.reason = "Balanced mode for mixed workloads"
			s.benefit = "Better thermal management, longer battery"
		case v.EPP == "balance_power":
			s.reason = "Favor power savings over raw speed"
			s.benefit = "Significant battery improvement on laptop"
		case v.EPP == "performance":
			s.reason = "Maximum CPU performance priority"
			s.benefit = "Best throughput for compute-heavy tasks"
		}
		sec.Fields = append(sec.Fields, s.fields()...)
	}

	// Turbo
	turboOn := false
	if val, err := sysfs.ReadInt(sysfs.IntelNoTurbo); err == nil {
		turboOn = val == 0
	} else if val, err := sysfs.ReadInt(sysfs.CPUBoost); err == nil {
		turboOn = val == 1
	}
	if turboOn != v.TurboOn {
		cur := "off"
		if turboOn {
			cur = "on"
		}
		rec := "off"
		if v.TurboOn {
			rec = "on"
		}
		s := suggestion{key: "Turbo Boost", current: cur, target: rec}
		if v.TurboOn {
			s.reason = "Allow CPU to exceed base frequency under load"
			s.benefit = "Better burst performance for short workloads"
		} else {
			s.reason = "Disable turbo to reduce heat and power draw"
			s.benefit = "Lower temperatures, quieter fan, longer battery"
		}
		sec.Fields = append(sec.Fields, s.fields()...)
	}

	return sec
}

func suggestMemory(p profile.Profile) output.Section {
	sec := output.Section{Title: "Memory Changes"}
	v := p.Values

	if cur, err := sysfs.ReadInt(sysfs.VMSwappiness); err == nil && cur != v.Swappiness {
		s := suggestion{
			key:     "Swappiness",
			current: fmt.Sprintf("%d", cur),
			target:  fmt.Sprintf("%d", v.Swappiness),
		}
		if v.Swappiness < cur {
			s.reason = "Reduce tendency to swap out active pages"
			s.benefit = "More responsive system, less disk I/O"
		} else {
			s.reason = "Allow more swapping to free RAM for caches"
			s.benefit = "Better memory utilization under pressure"
		}
		sec.Fields = append(sec.Fields, s.fields()...)
	}

	if cur, err := sysfs.ReadInt(sysfs.VMDirtyBgRatio); err == nil && cur != v.DirtyBgRatio {
		s := suggestion{
			key:     "Dirty BG Ratio",
			current: fmt.Sprintf("%d%%", cur),
			target:  fmt.Sprintf("%d%%", v.DirtyBgRatio),
			reason:  "Flush dirty pages to disk sooner in background",
			benefit: "Less data loss risk on crash, smoother I/O",
		}
		sec.Fields = append(sec.Fields, s.fields()...)
	}

	if cur, err := sysfs.ReadInt(sysfs.VMDirtyRatio); err == nil && cur != v.DirtyRatio {
		s := suggestion{
			key:     "Dirty Ratio",
			current: fmt.Sprintf("%d%%", cur),
			target:  fmt.Sprintf("%d%%", v.DirtyRatio),
			reason:  "Limit dirty page buildup before forced writeback",
			benefit: "Prevents I/O stalls during heavy writes",
		}
		sec.Fields = append(sec.Fields, s.fields()...)
	}

	if cur, err := sysfs.ReadBracketedValue(sysfs.THPEnabled); err == nil && cur != v.THPEnabled {
		s := suggestion{
			key: "THP", current: cur, target: v.THPEnabled,
		}
		if v.THPEnabled == "madvise" {
			s.reason = "Only use huge pages when applications request them"
			s.benefit = "Avoids THP compaction stalls, apps opt-in as needed"
		}
		sec.Fields = append(sec.Fields, s.fields()...)
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
			s := suggestion{
				key:     fmt.Sprintf("%s scheduler", disk.Name),
				current: disk.Scheduler,
				target:  recommended,
			}
			switch recommended {
			case "none":
				s.reason = "NVMe has internal scheduling, kernel scheduler adds overhead"
				s.benefit = "Lower latency, better IOPS on NVMe drives"
			case "kyber":
				s.reason = "Lightweight scheduler optimized for fast SSDs"
				s.benefit = "Good latency guarantees with minimal overhead"
			case "bfq":
				s.reason = "Fair queuing prevents I/O starvation"
				s.benefit = "Better interactive responsiveness during heavy I/O"
			case "mq-deadline":
				s.reason = "Deadline guarantees prevent request starvation"
				s.benefit = "Predictable latency for mixed read/write workloads"
			}
			sec.Fields = append(sec.Fields, s.fields()...)
		}
	}

	return sec
}

func suggestNetwork(p profile.Profile) output.Section {
	sec := output.Section{Title: "Network Changes"}
	v := p.Values

	if cur, err := sysfs.ReadString(sysfs.TCPCongestion); err == nil && cur != v.TCPCongestion {
		s := suggestion{
			key: "TCP Congestion", current: cur, target: v.TCPCongestion,
		}
		if v.TCPCongestion == "bbr" {
			s.reason = "BBR uses bandwidth estimation instead of loss-based signals"
			s.benefit = "Better throughput on lossy or high-latency links"
		}
		sec.Fields = append(sec.Fields, s.fields()...)
	}

	if cur, err := sysfs.ReadInt(sysfs.TCPFastOpen); err == nil && cur != v.TCPFastOpen {
		s := suggestion{
			key:     "TCP Fast Open",
			current: fmt.Sprintf("%d", cur),
			target:  fmt.Sprintf("%d", v.TCPFastOpen),
			reason:  "Enable TFO for both client and server connections",
			benefit: "Reduces connection latency by ~1 round-trip on repeat connections",
		}
		sec.Fields = append(sec.Fields, s.fields()...)
	}

	if cur, err := sysfs.ReadInt(sysfs.NetCoreBufMax); err == nil && cur != v.RmemMax {
		s := suggestion{
			key:     "Recv Buffer Max",
			current: detect.FormatBytes(cur),
			target:  detect.FormatBytes(v.RmemMax),
			reason:  "Allow larger TCP receive windows for fast connections",
			benefit: "Better throughput for large downloads and transfers",
		}
		sec.Fields = append(sec.Fields, s.fields()...)
	}

	if cur, err := sysfs.ReadInt(sysfs.NetCoreWBufMax); err == nil && cur != v.WmemMax {
		s := suggestion{
			key:     "Send Buffer Max",
			current: detect.FormatBytes(cur),
			target:  detect.FormatBytes(v.WmemMax),
			reason:  "Allow larger TCP send windows for fast connections",
			benefit: "Better upload throughput for large transfers",
		}
		sec.Fields = append(sec.Fields, s.fields()...)
	}

	return sec
}

func suggestServer() output.Section {
	sec := output.Section{Title: "Server Tuning"}
	info := detect.DetectServer()

	if info.Somaxconn < 4096 {
		s := suggestion{
			key:     "Somaxconn",
			current: fmt.Sprintf("%d", info.Somaxconn),
			target:  "4096",
			reason:  "Higher connection backlog for busy services",
			benefit: "Handles more concurrent connection attempts",
		}
		sec.Fields = append(sec.Fields, s.fields()...)
	}

	if info.FileMax < 1048576 {
		s := suggestion{
			key:     "File Max",
			current: fmt.Sprintf("%d", info.FileMax),
			target:  "1048576",
			reason:  "More open file descriptors for services",
			benefit: "Supports more connections and open files",
		}
		sec.Fields = append(sec.Fields, s.fields()...)
	}

	if info.PortRange != "" {
		// Check if port range is narrower than recommended "1024 65535"
		var low, high int
		fmt.Sscanf(info.PortRange, "%d %d", &low, &high)
		if low > 1024 || high < 65535 {
			s := suggestion{
				key:     "Port Range",
				current: info.PortRange,
				target:  "1024 65535",
				reason:  "Wider ephemeral port range for outbound connections",
				benefit: "More ports available for high-connection workloads",
			}
			sec.Fields = append(sec.Fields, s.fields()...)
		}
	}

	if info.ConntrackMax >= 0 && info.ConntrackMax < 262144 {
		s := suggestion{
			key:     "Conntrack Max",
			current: fmt.Sprintf("%d", info.ConntrackMax),
			target:  "262144",
			reason:  "Track more concurrent connections",
			benefit: "Prevents conntrack table overflow under load",
		}
		sec.Fields = append(sec.Fields, s.fields()...)
	}

	if info.IRQBalance.Installed && !info.IRQBalance.Active {
		sec.Fields = append(sec.Fields,
			output.Field{
				Key:    "IRQ Balance",
				Value:  "installed but not active",
				Status: output.StatusWarn,
			},
			output.Field{
				Key:    "",
				Value:  "Run: sudo systemctl enable --now irqbalance",
				Status: output.StatusNone,
			},
		)
	}

	powerInfo := detect.DetectPower()
	if powerInfo.Tuned.Installed && !powerInfo.Tuned.Active {
		sec.Fields = append(sec.Fields,
			output.Field{
				Key:    "tuned",
				Value:  "not active - recommended for server workloads",
				Status: output.StatusWarn,
			},
			output.Field{
				Key:    "",
				Value:  "Run: sudo tuned-adm profile throughput-performance",
				Status: output.StatusNone,
			},
		)
	}

	return sec
}
