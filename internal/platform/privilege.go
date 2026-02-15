package platform

import (
	"fmt"
	"os"
)

// IsRoot returns true if the current process is running as root.
func IsRoot() bool {
	return os.Geteuid() == 0
}

// RequireRoot exits with an error message if not running as root.
func RequireRoot(command string) {
	if !IsRoot() {
		fmt.Fprintf(os.Stderr, "Error: '%s' requires root privileges. Run with sudo.\n", command)
		os.Exit(1)
	}
}
