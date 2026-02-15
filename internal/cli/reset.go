package cli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/krisk248/tuner/internal/persist"
	"github.com/krisk248/tuner/internal/platform"
	"github.com/krisk248/tuner/internal/sysfs"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Revert tuning changes from backup",
	Long:  "Restores original system values from backup and removes persisted configs.",
	RunE:  runReset,
}

func init() {
	rootCmd.AddCommand(resetCmd)
}

func runReset(cmd *cobra.Command, args []string) error {
	platform.RequireRoot("reset")

	if !persist.BackupExists() {
		return fmt.Errorf("no backup found at %s. Nothing to reset", persist.BackupFile)
	}

	backup, err := persist.LoadBackup()
	if err != nil {
		return fmt.Errorf("failed to load backup: %w", err)
	}

	fmt.Printf("Restoring values from backup (profile: %s, saved: %s)\n\n", backup.Profile, backup.Timestamp)

	// Restore original values
	restored := 0
	failed := 0
	for path, value := range backup.Values {
		if !sysfs.Exists(path) {
			continue
		}
		fmt.Printf("  %s → %s ... ", path, value)
		if err := sysfs.WriteString(path, value); err != nil {
			color.Red("FAILED (%v)", err)
			failed++
		} else {
			color.Green("OK")
			restored++
		}
	}

	// Remove persisted configs
	fmt.Println()
	if err := persist.RemoveSysctl(); err != nil {
		color.Yellow("Warning: failed to remove sysctl config: %v", err)
	} else {
		fmt.Printf("  Removed %s\n", persist.SysctlPath())
	}

	if err := persist.RemoveUdev(); err != nil {
		color.Yellow("Warning: failed to remove udev rules: %v", err)
	} else {
		fmt.Printf("  Removed %s\n", persist.UdevPath())
	}

	// Reload
	persist.ReloadSysctl()
	persist.ReloadUdev()

	// Remove backup
	if err := persist.RemoveBackup(); err != nil {
		color.Yellow("Warning: failed to remove backup: %v", err)
	}

	fmt.Printf("\nRestored %d values, %d failed\n", restored, failed)
	color.Green("Reset complete.")
	return nil
}
