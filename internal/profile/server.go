package profile

// ServerValues returns tuning values optimized for server workloads.
func ServerValues() Values {
	return Values{
		Governor:         "performance",
		EPP:              "performance",
		TurboOn:          true,
		Swappiness:       10,
		DirtyBgRatio:     1,
		DirtyRatio:       5,
		DirtyExpire:      500,
		DirtyWriteback:   100,
		VFSCachePressure: 50,
		THPEnabled:       "always",
		TCPCongestion:    "bbr",
		TCPFastOpen:      3,
		TCPMTUProbing:    1,
		RmemMax:          268435456, // 256 MB
		WmemMax:          268435456,
		TCPRmem:          "4096 1048576 268435456",
		TCPWmem:          "4096 1048576 268435456",
		SchedNVMe:        "none",
		SchedSSD:         "kyber",
		SchedHDD:         "bfq",
		ReadAhead:        256,
		SkipIfTLP:        false,
	}
}
