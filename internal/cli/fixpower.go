package cli

import (
	"fmt"
	"os/exec"

	"github.com/fatih/color"
	"github.com/krisk248/tuner/internal/detect"
	"github.com/krisk248/tuner/internal/platform"
	"github.com/krisk248/tuner/internal/profile"
	"github.com/spf13/cobra"
)

var fixPowerCmd = &cobra.Command{
	Use:   "fix-power",
	Short: "Fix power management (stop tuned, start TLP) for laptops",
	RunE:  runFixPower,
}

func init() {
	rootCmd.AddCommand(fixPowerCmd)
}

func runFixPower(cmd *cobra.Command, args []string) error {
	// Refuse on server profile
	p := profile.AutoDetect()
	if p.Type == profile.Server {
		return fmt.Errorf("fix-power is for laptops/desktops. Servers should use tuned instead")
	}

	platform.RequireRoot("fix-power")

	power := detect.DetectPower()
	bold := color.New(color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	acted := false

	// Stop and disable tuned if active
	if power.Tuned.Active {
		bold.Println("Stopping tuned...")
		if err := exec.Command("systemctl", "stop", "tuned").Run(); err != nil {
			return fmt.Errorf("failed to stop tuned: %w", err)
		}
		if err := exec.Command("systemctl", "disable", "tuned").Run(); err != nil {
			return fmt.Errorf("failed to disable tuned: %w", err)
		}
		green.Println("  tuned stopped and disabled")
		acted = true
	}

	// Start TLP if enabled but not active
	if power.TLP.Enabled && !power.TLP.Active {
		bold.Println("Starting TLP...")
		if err := exec.Command("systemctl", "start", "tlp").Run(); err != nil {
			return fmt.Errorf("failed to start tlp: %w", err)
		}
		green.Println("  TLP started")
		acted = true
	}

	// TLP not installed at all
	if !power.TLP.Installed {
		yellow.Println("TLP not found. Install with:")
		fmt.Println("  sudo dnf install tlp    # Fedora/RHEL")
		fmt.Println("  sudo apt install tlp    # Ubuntu/Debian")
		fmt.Println("  sudo pacman -S tlp      # Arch")
		return nil
	}

	if !acted {
		fmt.Println("Power management is already correctly configured.")
	} else {
		fmt.Println()
		green.Println("Power management fixed.")
	}

	return nil
}
