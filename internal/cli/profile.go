package cli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/krisk248/tuner/internal/profile"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Auto-detect machine profile (server/desktop/laptop)",
	RunE:  runProfile,
}

func init() {
	rootCmd.AddCommand(profileCmd)
}

func runProfile(cmd *cobra.Command, args []string) error {
	p := profile.AutoDetect()

	bold := color.New(color.Bold)
	bold.Printf("Detected profile: ")
	profileColor := color.New(color.Bold, color.FgGreen)
	profileColor.Println(p.Type)

	if p.Type == profile.Laptop {
		fmt.Printf("Power state: %s\n", p.PowerState)
	}

	return nil
}
