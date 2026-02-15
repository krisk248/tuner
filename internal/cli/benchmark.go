package cli

import (
	"fmt"

	"github.com/krisk248/tuner/internal/benchmark"
	"github.com/spf13/cobra"
)

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Run system benchmarks",
	RunE:  runBenchmark,
}

var (
	benchDisk    bool
	benchNetwork bool
)

func init() {
	benchmarkCmd.Flags().BoolVar(&benchDisk, "disk", false, "run disk benchmark")
	benchmarkCmd.Flags().BoolVar(&benchNetwork, "network", false, "run network benchmark")
	rootCmd.AddCommand(benchmarkCmd)
}

func runBenchmark(cmd *cobra.Command, args []string) error {
	// Default to both if none specified
	if !benchDisk && !benchNetwork {
		benchDisk = true
		benchNetwork = true
	}

	if benchDisk {
		result, err := benchmark.RunDiskBenchmark()
		if err != nil {
			return fmt.Errorf("disk benchmark failed: %w", err)
		}
		fmt.Println()
		benchmark.PrintDiskResult(result)
		fmt.Println()
	}

	if benchNetwork {
		result, err := benchmark.RunNetworkBenchmark()
		if err != nil {
			return fmt.Errorf("network benchmark failed: %w", err)
		}
		fmt.Println()
		benchmark.PrintNetworkResult(result)
	}

	return nil
}
