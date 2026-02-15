package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/krisk248/tuner/internal/platform"
	"github.com/krisk248/tuner/internal/profile"
	"github.com/krisk248/tuner/internal/tune"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply tuning changes for a profile",
	RunE:  runApply,
}

var (
	applyProfile string
	applyAuto    bool
)

func init() {
	applyCmd.Flags().StringVar(&applyProfile, "profile", "", "profile to apply (server, desktop, laptop)")
	applyCmd.Flags().BoolVar(&applyAuto, "auto", false, "apply without confirmation")
	rootCmd.AddCommand(applyCmd)
}

func runApply(cmd *cobra.Command, args []string) error {
	platform.RequireRoot("apply")

	var p profile.Profile
	if applyProfile != "" {
		p = profile.ForType(profile.Type(applyProfile))
	} else {
		p = profile.AutoDetect()
	}

	bold := color.New(color.Bold)
	bold.Printf("Profile: %s\n", p.Type)
	if p.Type == profile.Laptop {
		fmt.Printf("Power state: %s\n", p.PowerState)
	}

	engine := tune.NewEngine(p)
	changes := engine.ComputeChanges()

	if len(changes) == 0 {
		fmt.Println("No changes needed. System is already optimally configured.")
		return nil
	}

	fmt.Printf("\n%d changes to apply:\n\n", len(changes))
	for _, c := range changes {
		fmt.Printf("  [%s] %s: %s → %s\n", c.Subsystem, c.Parameter, c.OldValue, c.NewValue)
	}

	if !applyAuto {
		fmt.Printf("\nProceed? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	fmt.Println()
	success, failed := engine.Apply(changes, applyAuto)

	fmt.Printf("\nApplied: %d succeeded, %d failed\n", success, failed)

	if failed > 0 {
		color.Yellow("Some changes failed.")
	}

	return nil
}
