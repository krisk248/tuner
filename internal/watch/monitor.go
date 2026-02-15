package watch

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/krisk248/tuner/internal/sysfs"
)

// Config holds watch mode configuration.
type Config struct {
	CPU     bool
	Memory  bool
	Network bool
	All     bool
}

// Run starts the live monitoring loop.
func Run(cfg Config) error {
	if cfg.All || (!cfg.CPU && !cfg.Memory && !cfg.Network) {
		cfg.CPU = true
		cfg.Memory = true
		cfg.Network = true
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var prevNetRx, prevNetTx int64
	firstRun := true

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigCh:
			fmt.Print("\033[?25h") // show cursor
			fmt.Println("\nExiting watch mode.")
			return nil
		case <-ticker.C:
			var lines []string

			// Clear screen and move to top
			fmt.Print("\033[2J\033[H")
			fmt.Print("\033[?25l") // hide cursor

			header := color.New(color.Bold, color.FgCyan)
			header.Println("tuner watch - press Ctrl+C to exit")
			fmt.Println(strings.Repeat("─", 60))

			if cfg.CPU {
				lines = append(lines, cpuLines()...)
			}
			if cfg.Memory {
				lines = append(lines, memoryLines()...)
			}
			if cfg.Network {
				netLines, rx, tx := networkLines(prevNetRx, prevNetTx, firstRun)
				lines = append(lines, netLines...)
				prevNetRx = rx
				prevNetTx = tx
			}

			for _, line := range lines {
				fmt.Println(line)
			}

			firstRun = false
		}
	}
}

func cpuLines() []string {
	var lines []string
	bold := color.New(color.Bold)
	lines = append(lines, bold.Sprint("CPU"))

	// Governor
	if gov, err := sysfs.ReadString(sysfs.CPUGovernor); err == nil {
		lines = append(lines, fmt.Sprintf("  Governor: %s", gov))
	}

	// Current frequency
	if freq, err := sysfs.ReadInt(sysfs.CPUFreqCur); err == nil {
		lines = append(lines, fmt.Sprintf("  Frequency: %d MHz", freq/1000))
	}

	// Load average
	if load, err := sysfs.ReadString("/proc/loadavg"); err == nil {
		fields := strings.Fields(load)
		if len(fields) >= 3 {
			lines = append(lines, fmt.Sprintf("  Load: %s %s %s", fields[0], fields[1], fields[2]))
		}
	}

	// CPU usage from /proc/stat (simplified - total across all cores)
	if stat, err := sysfs.ReadLines("/proc/stat"); err == nil && len(stat) > 0 {
		fields := strings.Fields(stat[0])
		if len(fields) >= 5 && fields[0] == "cpu" {
			var user, nice, system, idle int64
			fmt.Sscanf(fields[1], "%d", &user)
			fmt.Sscanf(fields[2], "%d", &nice)
			fmt.Sscanf(fields[3], "%d", &system)
			fmt.Sscanf(fields[4], "%d", &idle)
			total := user + nice + system + idle
			if total > 0 {
				usePct := float64(user+nice+system) / float64(total) * 100
				bar := renderBar(usePct, 30)
				lines = append(lines, fmt.Sprintf("  Usage: %s %.0f%%", bar, usePct))
			}
		}
	}

	lines = append(lines, "")
	return lines
}

func memoryLines() []string {
	var lines []string
	bold := color.New(color.Bold)
	lines = append(lines, bold.Sprint("Memory"))

	memInfo := parseMemInfo()

	totalKB := memInfo["MemTotal"]
	availKB := memInfo["MemAvailable"]
	usedKB := totalKB - availKB

	if totalKB > 0 {
		usePct := float64(usedKB) / float64(totalKB) * 100
		bar := renderBar(usePct, 30)
		lines = append(lines, fmt.Sprintf("  RAM: %s %.0f%% (%.1f/%.1f GB)",
			bar, usePct, float64(usedKB)/1048576, float64(totalKB)/1048576))
	}

	swapTotal := memInfo["SwapTotal"]
	swapFree := memInfo["SwapFree"]
	if swapTotal > 0 {
		swapUsed := swapTotal - swapFree
		swapPct := float64(swapUsed) / float64(swapTotal) * 100
		bar := renderBar(swapPct, 30)
		lines = append(lines, fmt.Sprintf("  Swap: %s %.0f%% (%.1f/%.1f GB)",
			bar, swapPct, float64(swapUsed)/1048576, float64(swapTotal)/1048576))
	}

	lines = append(lines, "")
	return lines
}

func networkLines(prevRx, prevTx int64, first bool) ([]string, int64, int64) {
	var lines []string
	bold := color.New(color.Bold)
	lines = append(lines, bold.Sprint("Network"))

	totalRx, totalTx := getNetBytes()

	if !first && prevRx > 0 {
		rxRate := float64(totalRx-prevRx) / 1024 // KB/s
		txRate := float64(totalTx-prevTx) / 1024
		rxStr := formatRate(rxRate)
		txStr := formatRate(txRate)
		lines = append(lines, fmt.Sprintf("  RX: %s/s  TX: %s/s", rxStr, txStr))
	} else {
		lines = append(lines, "  Measuring...")
	}

	if cong, err := sysfs.ReadString(sysfs.TCPCongestion); err == nil {
		lines = append(lines, fmt.Sprintf("  TCP: %s", cong))
	}

	lines = append(lines, "")
	return lines, totalRx, totalTx
}

func getNetBytes() (rx, tx int64) {
	entries, err := os.ReadDir(sysfs.NetBase)
	if err != nil {
		return 0, 0
	}
	for _, e := range entries {
		name := e.Name()
		if name == "lo" {
			continue
		}
		if v, err := sysfs.ReadInt64(fmt.Sprintf("%s/%s/statistics/rx_bytes", sysfs.NetBase, name)); err == nil {
			rx += v
		}
		if v, err := sysfs.ReadInt64(fmt.Sprintf("%s/%s/statistics/tx_bytes", sysfs.NetBase, name)); err == nil {
			tx += v
		}
	}
	return
}

func parseMemInfo() map[string]int64 {
	result := make(map[string]int64)
	lines, err := sysfs.ReadLines(sysfs.ProcMemInfo)
	if err != nil {
		return result
	}
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		var val int64
		fmt.Sscanf(fields[1], "%d", &val)
		result[key] = val
	}
	return result
}

func renderBar(pct float64, width int) string {
	filled := int(pct / 100 * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	if pct > 90 {
		return color.RedString("[%s]", bar)
	} else if pct > 75 {
		return color.YellowString("[%s]", bar)
	}
	return color.GreenString("[%s]", bar)
}

func formatRate(kbps float64) string {
	if kbps >= 1024 {
		return fmt.Sprintf("%.1f MB", kbps/1024)
	}
	return fmt.Sprintf("%.1f KB", kbps)
}
