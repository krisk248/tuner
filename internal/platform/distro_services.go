package platform

// ServiceNames maps generic service names to distro-specific unit names.
type ServiceNames struct {
	Firewall  string
	SSH       string
	Cron      string
	NTP       string
	NetworkMgr string
}

// GetServiceNames returns distro-specific systemd service names.
func GetServiceNames(family Family) ServiceNames {
	switch family {
	case FamilyRHEL:
		return ServiceNames{
			Firewall:   "firewalld",
			SSH:        "sshd",
			Cron:       "crond",
			NTP:        "chronyd",
			NetworkMgr: "NetworkManager",
		}
	case FamilyDebian:
		return ServiceNames{
			Firewall:   "ufw",
			SSH:        "ssh",
			Cron:       "cron",
			NTP:        "systemd-timesyncd",
			NetworkMgr: "NetworkManager",
		}
	case FamilyArch:
		return ServiceNames{
			Firewall:   "firewalld",
			SSH:        "sshd",
			Cron:       "cronie",
			NTP:        "systemd-timesyncd",
			NetworkMgr: "NetworkManager",
		}
	default:
		return ServiceNames{
			Firewall:   "firewalld",
			SSH:        "sshd",
			Cron:       "crond",
			NTP:        "chronyd",
			NetworkMgr: "NetworkManager",
		}
	}
}
