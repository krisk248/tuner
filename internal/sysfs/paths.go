package sysfs

const (
	// CPU
	CPUBase          = "/sys/devices/system/cpu"
	CPUFreqBase      = "/sys/devices/system/cpu/cpu0/cpufreq"
	CPUGovernor      = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor"
	CPUAvailGovs     = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_available_governors"
	CPUDriver        = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_driver"
	CPUEPP           = "/sys/devices/system/cpu/cpu0/cpufreq/energy_performance_preference"
	CPUAvailEPP      = "/sys/devices/system/cpu/cpu0/cpufreq/energy_performance_available_preferences"
	CPUFreqMin       = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_min_freq"
	CPUFreqMax       = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_max_freq"
	CPUFreqCur       = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq"
	CPUFreqBaseFreq  = "/sys/devices/system/cpu/cpu0/cpufreq/base_frequency"
	CPUBoost         = "/sys/devices/system/cpu/cpufreq/boost"
	IntelNoTurbo     = "/sys/devices/system/cpu/intel_pstate/no_turbo"
	IntelPStateStatus = "/sys/devices/system/cpu/intel_pstate/status"
	ProcCPUInfo      = "/proc/cpuinfo"

	// Memory
	ProcMemInfo   = "/proc/meminfo"
	VMSwappiness  = "/proc/sys/vm/swappiness"
	VMDirtyBgRatio = "/proc/sys/vm/dirty_background_ratio"
	VMDirtyRatio  = "/proc/sys/vm/dirty_ratio"
	VMDirtyExpire = "/proc/sys/vm/dirty_expire_centisecs"
	VMDirtyWriteback = "/proc/sys/vm/dirty_writeback_centisecs"
	VMVFSCachePressure = "/proc/sys/vm/vfs_cache_pressure"
	THPEnabled    = "/sys/kernel/mm/transparent_hugepage/enabled"
	THPDefrag     = "/sys/kernel/mm/transparent_hugepage/defrag"
	ZswapEnabled  = "/sys/module/zswap/parameters/enabled"
	ZswapCompressor = "/sys/module/zswap/parameters/compressor"
	ZswapMaxPool  = "/sys/module/zswap/parameters/max_pool_percent"

	// Storage
	BlockBase     = "/sys/block"

	// Network
	NetBase       = "/sys/class/net"
	TCPCongestion = "/proc/sys/net/ipv4/tcp_congestion_control"
	TCPAvailCong  = "/proc/sys/net/ipv4/tcp_available_congestion_control"
	TCPRmem       = "/proc/sys/net/ipv4/tcp_rmem"
	TCPWmem       = "/proc/sys/net/ipv4/tcp_wmem"
	TCPFastOpen   = "/proc/sys/net/ipv4/tcp_fastopen"
	TCPMTUProbing = "/proc/sys/net/ipv4/tcp_mtu_probing"
	NetCoreBufMax = "/proc/sys/net/core/rmem_max"
	NetCoreWBufMax = "/proc/sys/net/core/wmem_max"

	// Power
	PowerSupplyBase = "/sys/class/power_supply"
	ChassisType     = "/sys/class/dmi/id/chassis_type"

	// Kernel
	ProcVersion  = "/proc/version"
	ProcCmdline  = "/proc/cmdline"

	// GPU
	DRMBase      = "/sys/class/drm"
)
