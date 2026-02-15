package profile

// Values holds all tuning parameters for a profile.
type Values struct {
	// CPU
	Governor string
	EPP      string
	TurboOn  bool

	// Memory
	Swappiness       int
	DirtyBgRatio     int
	DirtyRatio       int
	DirtyExpire      int // centisecs
	DirtyWriteback   int // centisecs
	VFSCachePressure int
	THPEnabled       string // always, madvise, never

	// Network
	TCPCongestion string
	TCPFastOpen   int
	TCPMTUProbing int
	RmemMax       int // bytes
	WmemMax       int // bytes
	TCPRmem       string // "min default max"
	TCPWmem       string // "min default max"

	// Storage (scheduler recommendations by device type)
	SchedNVMe string
	SchedSSD  string
	SchedHDD  string
	ReadAhead int // KB

	// Power
	SkipIfTLP bool // don't touch power if TLP is active
}

// IOScheduler returns the recommended scheduler for a device type.
func (v Values) IOScheduler(diskType string) string {
	switch diskType {
	case "nvme":
		return v.SchedNVMe
	case "ssd":
		return v.SchedSSD
	case "hdd":
		return v.SchedHDD
	default:
		return v.SchedHDD
	}
}
