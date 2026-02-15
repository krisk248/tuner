package cli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/krisk248/tuner/internal/persist"
	"github.com/krisk248/tuner/internal/platform"
	"github.com/krisk248/tuner/internal/profile"
	"github.com/krisk248/tuner/internal/tune"
	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Persist tuning changes to survive reboots",
	Long:  "Writes sysctl.d and udev rules so tunings persist across reboots.",
	RunE:  runSave,
}

var saveProfile string

func init() {
	saveCmd.Flags().StringVar(&saveProfile, "profile", "", "profile to save (server, desktop, laptop)")
	rootCmd.AddCommand(saveCmd)
}

func runSave(cmd *cobra.Command, args []string) error {
	platform.RequireRoot("save")

	var p profile.Profile
	if saveProfile != "" {
		p = profile.ForType(profile.Type(saveProfile))
	} else {
		p = profile.AutoDetect()
	}

	fmt.Printf("Saving tuning for profile: %s\n", p.Type)

	// Backup current values before persisting
	engine := tune.NewEngine(p)
	changes := engine.ComputeChanges()
	backup := tune.Backup(changes)

	if err := persist.SaveBackup(string(p.Type), backup); err != nil {
		return fmt.Errorf("failed to save backup: %w", err)
	}
	fmt.Printf("  Backup saved to %s\n", persist.BackupFile)

	// Write sysctl drop-in
	if err := persist.WriteSysctl(p); err != nil {
		return fmt.Errorf("failed to write sysctl config: %w", err)
	}
	fmt.Printf("  Written %s\n", persist.SysctlPath())

	// Write udev rules
	if err := persist.WriteUdev(p); err != nil {
		return fmt.Errorf("failed to write udev rules: %w", err)
	}
	fmt.Printf("  Written %s\n", persist.UdevPath())

	// Reload
	if err := persist.ReloadSysctl(); err != nil {
		color.Yellow("Warning: failed to reload sysctl: %v", err)
	}
	if err := persist.ReloadUdev(); err != nil {
		color.Yellow("Warning: failed to reload udev: %v", err)
	}

	color.Green("Tuning persisted successfully.")
	return nil
}
