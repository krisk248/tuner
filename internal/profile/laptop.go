package profile

// LaptopValues returns tuning values for laptops, adjusted by power state.
func LaptopValues(state PowerState) Values {
	if state == OnBattery {
		return laptopBatteryValues()
	}
	return laptopACValues()
}

func laptopACValues() Values {
	return Values{
		Governor:         "schedutil",
		EPP:              "balance_performance",
		TurboOn:          true,
		Swappiness:       10,
		DirtyBgRatio:     5,
		DirtyRatio:       15,
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
		SchedSSD:         "bfq",
		SchedHDD:         "bfq",
		ReadAhead:        256,
		SkipIfTLP:        true,
	}
}

func laptopBatteryValues() Values {
	return Values{
		Governor:         "powersave",
		EPP:              "balance_power",
		TurboOn:          false,
		Swappiness:       30,
		DirtyBgRatio:     5,
		DirtyRatio:       15,
		DirtyExpire:      6000,
		DirtyWriteback:   1500,
		VFSCachePressure: 100,
		THPEnabled:       "madvise",
		TCPCongestion:    "bbr",
		TCPFastOpen:      3,
		TCPMTUProbing:    1,
		RmemMax:          67108864,
		WmemMax:          67108864,
		TCPRmem:          "4096 131072 67108864",
		TCPWmem:          "4096 131072 67108864",
		SchedNVMe:        "none",
		SchedSSD:         "bfq",
		SchedHDD:         "bfq",
		ReadAhead:        128,
		SkipIfTLP:        true,
	}
}
