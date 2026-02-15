package cli

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	verbose   bool
	noColor   bool
	outFormat string
)

// SetVersion sets the version string from main.
func SetVersion(v string) {
	version = v
}

var rootCmd = &cobra.Command{
	Use:   "tuner",
	Short: "Linux system diagnostic and tuning tool",
	Long:  "Tuner detects hardware/software state, recommends optimizations, and applies them.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor {
			color.NoColor = true
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().StringVarP(&outFormat, "format", "f", "table", "output format (table, json, markdown)")
	rootCmd.Version = version
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
