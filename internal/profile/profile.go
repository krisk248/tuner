package profile

import (
	"os"
	"os/exec"
	"strings"

	"github.com/krisk248/tuner/internal/sysfs"
)

// Type represents the machine profile type.
type Type string

const (
	Server  Type = "server"
	Desktop Type = "desktop"
	Laptop  Type = "laptop"
)

// PowerState represents AC vs battery for laptops.
type PowerState string

const (
	OnAC      PowerState = "ac"
	OnBattery PowerState = "battery"
)

// Profile holds the detected or selected profile with tuning values.
type Profile struct {
	Type       Type
	PowerState PowerState
	Values     Values
}

// AutoDetect determines the machine profile.
func AutoDetect() Profile {
	p := Profile{PowerState: OnAC}

	// 1. Check for battery -> laptop
	if hasBattery() {
		p.Type = Laptop
		p.PowerState = detectPowerState()
		p.Values = LaptopValues(p.PowerState)
		return p
	}

	// 2. Check chassis type via DMI
	if chassisType := readChassisType(); chassisType > 0 {
		switch {
		case isServerChassis(chassisType):
			p.Type = Server
			p.Values = ServerValues()
			return p
		case isLaptopChassis(chassisType):
			p.Type = Laptop
			p.PowerState = detectPowerState()
			p.Values = LaptopValues(p.PowerState)
			return p
		}
	}

	// 3. No display + multi-user target -> server
	if !hasDisplay() && isMultiUserTarget() {
		p.Type = Server
		p.Values = ServerValues()
		return p
	}

	// 4. Fallback -> desktop
	p.Type = Desktop
	p.Values = DesktopValues()
	return p
}

// ForType returns a profile for an explicitly chosen type.
func ForType(t Type) Profile {
	p := Profile{Type: t, PowerState: OnAC}
	switch t {
	case Server:
		p.Values = ServerValues()
	case Laptop:
		p.PowerState = detectPowerState()
		p.Values = LaptopValues(p.PowerState)
	default:
		p.Values = DesktopValues()
	}
	return p
}

func hasBattery() bool {
	entries, err := os.ReadDir(sysfs.PowerSupplyBase)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "BAT") {
			return true
		}
	}
	return false
}

func detectPowerState() PowerState {
	entries, err := os.ReadDir(sysfs.PowerSupplyBase)
	if err != nil {
		return OnAC
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "AC") || strings.HasPrefix(e.Name(), "ADP") {
			path := sysfs.PowerSupplyBase + "/" + e.Name() + "/online"
			if v, err := sysfs.ReadInt(path); err == nil && v == 0 {
				return OnBattery
			}
		}
	}
	return OnAC
}

func readChassisType() int {
	v, err := sysfs.ReadInt(sysfs.ChassisType)
	if err != nil {
		return 0
	}
	return v
}

func isServerChassis(t int) bool {
	// 17=Main Server Chassis, 23=Rack Mount Chassis,
	// 25=Multi-system Chassis, 28=Sealed-case PC, 29=Multi-system Chassis
	return t == 17 || t == 23 || t == 25 || t == 28 || t == 29
}

func isLaptopChassis(t int) bool {
	// 8=Portable, 9=Laptop, 10=Notebook, 14=Sub Notebook,
	// 31=Convertible, 32=Detachable
	return t == 8 || t == 9 || t == 10 || t == 14 || t == 31 || t == 32
}

func hasDisplay() bool {
	return os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
}

func isMultiUserTarget() bool {
	out, err := exec.Command("systemctl", "get-default").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "multi-user.target"
}
