package persist

import (
	"os/exec"
)

// ReloadSysctl applies sysctl settings from the drop-in file.
func ReloadSysctl() error {
	return exec.Command("sysctl", "--system").Run()
}

// ReloadUdev triggers udev rules reload.
func ReloadUdev() error {
	if err := exec.Command("udevadm", "control", "--reload-rules").Run(); err != nil {
		return err
	}
	return exec.Command("udevadm", "trigger").Run()
}
