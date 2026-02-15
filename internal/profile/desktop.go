package profile

// DesktopValues returns tuning values optimized for desktop workloads.
func DesktopValues() Values {
	return Values{
		Governor:         "performance",
		EPP:              "balance_performance",
		TurboOn:          true,
		Swappiness:       10,
		DirtyBgRatio:     10,
		DirtyRatio:       20,
		DirtyExpire:      3000,
		DirtyWriteback:   500,
		VFSCachePressure: 100,
		THPEnabled:       "madvise",
		TCPCongestion:    "bbr",
		TCPFastOpen:      3,
		TCPMTUProbing:    1,
		RmemMax:          67108864, // 64 MB
		WmemMax:          67108864,
		TCPRmem:          "4096 131072 67108864",
		TCPWmem:          "4096 131072 67108864",
		SchedNVMe:        "none",
		SchedSSD:         "kyber",
		SchedHDD:         "bfq",
		ReadAhead:        256,
		SkipIfTLP:        false,
	}
}
