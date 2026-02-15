package platform

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/krisk248/tuner/internal/sysfs"
)

// KernelVersion holds parsed kernel version info.
type KernelVersion struct {
	Major  int
	Minor  int
	Patch  int
	Full   string // raw version string from /proc/version
	Release string // e.g. "6.12.5-200.fc41.x86_64"
}

// DetectKernel parses the running kernel version.
func DetectKernel() KernelVersion {
	kv := KernelVersion{}

	full, err := sysfs.ReadString(sysfs.ProcVersion)
	if err != nil {
		return kv
	}
	kv.Full = full

	// /proc/version format: "Linux version 6.12.5-200.fc41.x86_64 ..."
	fields := strings.Fields(full)
	if len(fields) < 3 {
		return kv
	}
	kv.Release = fields[2]

	// Parse major.minor.patch
	parts := strings.SplitN(kv.Release, "-", 2)
	version := parts[0]
	nums := strings.Split(version, ".")
	if len(nums) >= 1 {
		kv.Major, _ = strconv.Atoi(nums[0])
	}
	if len(nums) >= 2 {
		kv.Minor, _ = strconv.Atoi(nums[1])
	}
	if len(nums) >= 3 {
		kv.Patch, _ = strconv.Atoi(nums[2])
	}

	return kv
}

// AtLeast returns true if the kernel version is >= major.minor.
func (kv KernelVersion) AtLeast(major, minor int) bool {
	if kv.Major > major {
		return true
	}
	if kv.Major == major && kv.Minor >= minor {
		return true
	}
	return false
}

// String returns a human-readable version string.
func (kv KernelVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", kv.Major, kv.Minor, kv.Patch)
}
