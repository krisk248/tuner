package detect

import (
	"strings"

	"github.com/krisk248/tuner/internal/output"
	"github.com/krisk248/tuner/internal/platform"
	"github.com/krisk248/tuner/internal/sysfs"
)

// KernelInfo holds kernel diagnostic data.
type KernelInfo struct {
	Version    platform.KernelVersion
	Cmdline    string
	Distro     platform.Distro
	Preempt    string
	Tainted    string
}

// DetectKernel gathers kernel and distro information.
func DetectKernel() KernelInfo {
	info := KernelInfo{
		Version: platform.DetectKernel(),
		Distro:  platform.DetectDistro(),
	}

	if cmdline, err := sysfs.ReadString(sysfs.ProcCmdline); err == nil {
		info.Cmdline = cmdline
	}

	// Detect preemption model from /proc/version or cmdline
	if strings.Contains(info.Version.Full, "PREEMPT_DYNAMIC") {
		info.Preempt = "dynamic"
	} else if strings.Contains(info.Version.Full, "PREEMPT_RT") {
		info.Preempt = "rt"
	} else if strings.Contains(info.Version.Full, "PREEMPT") {
		info.Preempt = "voluntary"
	} else {
		info.Preempt = "none"
	}

	if tainted, err := sysfs.ReadString("/proc/sys/kernel/tainted"); err == nil {
		if tainted == "0" {
			info.Tainted = "clean"
		} else {
			info.Tainted = tainted
		}
	}

	return info
}

// KernelSection formats kernel info as an output section.
func KernelSection(info KernelInfo) output.Section {
	sec := output.Section{Title: "Kernel"}

	sec.Fields = append(sec.Fields,
		output.Field{Key: "Kernel Version", Value: info.Version.Release, Status: output.StatusInfo},
		output.Field{Key: "Distribution", Value: info.Distro.PrettyName, Status: output.StatusInfo},
		output.Field{Key: "Family", Value: string(info.Distro.Family), Status: output.StatusInfo},
		output.Field{Key: "Preemption", Value: info.Preempt, Status: output.StatusInfo},
	)

	if info.Tainted != "" {
		status := output.StatusGood
		if info.Tainted != "clean" {
			status = output.StatusWarn
		}
		sec.Fields = append(sec.Fields,
			output.Field{Key: "Tainted", Value: info.Tainted, Status: status},
		)
	}

	return sec
}
