package cli

import (
	"github.com/krisk248/tuner/internal/watch"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Live system monitoring dashboard",
	RunE:  runWatch,
}

var (
	watchCPU     bool
	watchMemory  bool
	watchNetwork bool
)

func init() {
	watchCmd.Flags().BoolVar(&watchCPU, "cpu", false, "show CPU stats")
	watchCmd.Flags().BoolVar(&watchMemory, "memory", false, "show memory stats")
	watchCmd.Flags().BoolVar(&watchNetwork, "network", false, "show network stats")
	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, args []string) error {
	cfg := watch.Config{
		CPU:     watchCPU,
		Memory:  watchMemory,
		Network: watchNetwork,
	}
	return watch.Run(cfg)
}
