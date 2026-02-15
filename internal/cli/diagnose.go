package cli

import (
	"os"

	"github.com/krisk248/tuner/internal/detect"
	"github.com/krisk248/tuner/internal/output"
	"github.com/krisk248/tuner/internal/profile"
	"github.com/spf13/cobra"
)

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "Diagnose system state across all subsystems",
	Long:  "Detects hardware/software state for CPU, memory, storage, network, power, services, kernel, and GPU.",
	RunE:  runDiagnose,
}

var (
	diagCPU      bool
	diagMemory   bool
	diagStorage  bool
	diagNetwork  bool
	diagPower    bool
	diagServices bool
	diagKernel   bool
	diagGPU      bool
	diagProfile  string
)

func init() {
	diagnoseCmd.Flags().BoolVar(&diagCPU, "cpu", false, "show CPU info only")
	diagnoseCmd.Flags().BoolVar(&diagMemory, "memory", false, "show memory info only")
	diagnoseCmd.Flags().BoolVar(&diagStorage, "storage", false, "show storage info only")
	diagnoseCmd.Flags().BoolVar(&diagNetwork, "network", false, "show network info only")
	diagnoseCmd.Flags().BoolVar(&diagPower, "power", false, "show power info only")
	diagnoseCmd.Flags().BoolVar(&diagServices, "services", false, "show services info only")
	diagnoseCmd.Flags().BoolVar(&diagKernel, "kernel", false, "show kernel info only")
	diagnoseCmd.Flags().BoolVar(&diagGPU, "gpu", false, "show GPU info only")
	diagnoseCmd.Flags().StringVar(&diagProfile, "profile", "", "filter output for profile (laptop, desktop, server, auto)")
	rootCmd.AddCommand(diagnoseCmd)
}

func runDiagnose(cmd *cobra.Command, args []string) error {
	// Resolve the effective mode
	mode := resolveMode(diagProfile)

	// If subsystem flags are set, they override profile filtering
	hasSubsystemFlag := diagCPU || diagMemory || diagStorage || diagNetwork ||
		diagPower || diagServices || diagKernel || diagGPU

	showAll := !hasSubsystemFlag

	var sections []output.Section

	if showAll || diagKernel {
		sections = append(sections, detect.KernelSection(detect.DetectKernel()))
	}
	if showAll || diagCPU {
		sections = append(sections, detect.CPUSection(detect.DetectCPU()))
	}
	if showAll || diagMemory {
		sections = append(sections, detect.MemorySection(detect.DetectMemory()))
	}
	if showAll || diagStorage {
		sections = append(sections, detect.StorageSection(detect.DetectStorage()))
	}
	if showAll || diagNetwork {
		sections = append(sections, detect.NetworkSection(detect.DetectNetwork(), mode))
	}
	if (showAll && mode != output.ModeServer) || diagPower {
		sections = append(sections, detect.PowerSection(detect.DetectPower()))
	}
	if showAll || diagServices {
		sections = append(sections, detect.ServicesSection(detect.DetectServices()))
	}
	if (showAll && mode != output.ModeServer) || diagGPU {
		sections = append(sections, detect.GPUSection(detect.DetectGPU()))
	}

	// Server extras
	if showAll && mode == output.ModeServer {
		sections = append(sections, detect.ServerSection(detect.DetectServer()))
	}

	formatter := output.NewFormatter(outFormat, noColor)
	return formatter.Format(os.Stdout, sections)
}

// resolveMode converts the --profile flag to a DiagMode, auto-detecting if needed.
func resolveMode(flag string) output.DiagMode {
	switch output.DiagMode(flag) {
	case output.ModeLaptop:
		return output.ModeLaptop
	case output.ModeDesktop:
		return output.ModeDesktop
	case output.ModeServer:
		return output.ModeServer
	case output.ModeAuto:
		return autoDetectMode()
	default:
		if flag == "" {
			return output.ModeAuto
		}
		return autoDetectMode()
	}
}

func autoDetectMode() output.DiagMode {
	p := profile.AutoDetect()
	switch p.Type {
	case profile.Laptop:
		return output.ModeLaptop
	case profile.Server:
		return output.ModeServer
	default:
		return output.ModeDesktop
	}
}
