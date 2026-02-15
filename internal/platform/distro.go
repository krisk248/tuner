package platform

import (
	"strings"

	"github.com/krisk248/tuner/internal/sysfs"
)

// Family represents the Linux distribution family.
type Family string

const (
	FamilyRHEL   Family = "rhel"
	FamilyDebian Family = "debian"
	FamilyArch   Family = "arch"
	FamilyUnknown Family = "unknown"
)

// Distro holds distribution information parsed from /etc/os-release.
type Distro struct {
	ID         string // e.g. "fedora", "ubuntu", "arch"
	Name       string // e.g. "Fedora Linux"
	Version    string // e.g. "41"
	VersionID  string // e.g. "41"
	PrettyName string // e.g. "Fedora Linux 41 (Workstation Edition)"
	Family     Family
}

// DetectDistro parses /etc/os-release to detect the Linux distribution.
func DetectDistro() Distro {
	d := Distro{Family: FamilyUnknown}

	lines, err := sysfs.ReadLines("/etc/os-release")
	if err != nil {
		return d
	}

	kv := make(map[string]string)
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		val := strings.Trim(parts[1], "\"")
		kv[key] = val
	}

	d.ID = kv["ID"]
	d.Name = kv["NAME"]
	d.Version = kv["VERSION"]
	d.VersionID = kv["VERSION_ID"]
	d.PrettyName = kv["PRETTY_NAME"]

	idLike := kv["ID_LIKE"]

	switch {
	case d.ID == "fedora" || d.ID == "rhel" || d.ID == "centos" || d.ID == "rocky" || d.ID == "almalinux" ||
		strings.Contains(idLike, "rhel") || strings.Contains(idLike, "fedora"):
		d.Family = FamilyRHEL
	case d.ID == "ubuntu" || d.ID == "debian" || d.ID == "linuxmint" || d.ID == "pop" ||
		strings.Contains(idLike, "debian") || strings.Contains(idLike, "ubuntu"):
		d.Family = FamilyDebian
	case d.ID == "arch" || d.ID == "manjaro" || d.ID == "endeavouros" ||
		strings.Contains(idLike, "arch"):
		d.Family = FamilyArch
	}

	return d
}

// PackageManager returns the package manager command for the distro family.
func (d Distro) PackageManager() string {
	switch d.Family {
	case FamilyRHEL:
		return "dnf"
	case FamilyDebian:
		return "apt"
	case FamilyArch:
		return "pacman"
	default:
		return ""
	}
}
