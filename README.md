# Tuner

A fast, zero-dependency Linux system diagnostic and tuning CLI. Single static binary, no CGO. Reads directly from sysfs/procfs for accurate hardware state, and applies profile-based optimizations.

Targets **Fedora, Ubuntu, Arch, RHEL 8+** on amd64.

## Install

```bash
# Build from source
git clone https://github.com/krisk248/tuner.git
cd tuner
make build

# Binary is at bin/tuner
sudo cp bin/tuner /usr/local/bin/
```

## Quick Start

```bash
# See what profile your system is
tuner profile

# Full system diagnostic
tuner diagnose

# Diagnostic filtered for a specific profile
tuner diagnose --profile server

# See what tuner would change
tuner suggest

# Apply changes (requires root)
sudo tuner apply

# Persist changes across reboots
sudo tuner save

# Revert everything
sudo tuner reset
```

## Commands

| Command | Description | Root |
|---------|-------------|------|
| `diagnose` | Detect hardware/software state across all subsystems | No |
| `suggest` | Show recommended tuning changes for your profile | No |
| `apply` | Apply tuning changes interactively | Yes |
| `save` | Persist changes to sysctl.d/udev (survives reboots) | Yes |
| `reset` | Revert all changes from backup | Yes |
| `fix-power` | Fix power manager conflicts (laptop only) | Yes |
| `profile` | Show auto-detected machine profile | No |
| `benchmark` | Run disk I/O and network speed tests | No |
| `watch` | Live terminal dashboard for system metrics | No |

## Profiles

Tuner auto-detects your machine type and tailors everything accordingly:

**Laptop** — Battery-aware tuning with AC/battery power states. Expects TLP for power management. Prioritizes battery life on battery, responsiveness on AC. Shows Wi-Fi signal quality, battery health.

**Desktop** — Always-plugged-in tuning. No power manager expected. Optimizes for responsiveness and throughput. Full CPU suggestions always shown.

**Server** — Throughput-focused tuning. Expects tuned with `throughput-performance` profile. Shows server-specific parameters (somaxconn, file-max, conntrack, port range, IRQ balance). Hides Wi-Fi and battery sections.

Override auto-detection with `--profile`:

```bash
tuner diagnose --profile server
tuner suggest --profile laptop
```

## Subsystems

- **CPU** — Governor, EPP, turbo boost, frequency scaling
- **Memory** — Swappiness, dirty ratios, THP, zswap
- **Storage** — I/O scheduler per device type (NVMe/SSD/HDD), read-ahead
- **Network** — TCP congestion, fast open, buffer sizes, NIC offloads, Wi-Fi quality
- **Power** — Battery health, TLP/tuned/PPD status, AC detection
- **Services** — Boot time analysis, failed units, slow services
- **Kernel** — Version, command line parameters
- **GPU** — DRM device detection
- **Server** — File descriptors, somaxconn, conntrack, port range, IRQ balance

## Output Formats

```bash
tuner diagnose                    # table (default)
tuner diagnose -f json            # JSON
tuner diagnose -f markdown        # Markdown
tuner diagnose --no-color         # no ANSI colors
```

## Subsystem Filters

Show only specific subsystems:

```bash
tuner diagnose --cpu --memory
tuner diagnose --network
tuner diagnose --storage --gpu
```

## Build

```bash
make build          # Build for linux/amd64
make test           # Run tests
make lint           # Run golangci-lint
make release        # Build release with goreleaser
```

The binary is statically linked with no CGO:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/tuner ./cmd/tuner
```

## Release

Tagged releases use [GoReleaser](https://goreleaser.com/):

```bash
git tag v0.1.0
git push origin v0.1.0
make release
```

See `.goreleaser.yaml` for configuration.

## Project Structure

```
cmd/tuner/          Entry point
internal/
  cli/              Cobra commands (diagnose, suggest, apply, etc.)
  detect/           Read-only hardware detection (8 subsystems)
  profile/          Hardcoded tuning profiles (laptop, desktop, server)
  tune/             Write-side tuning engine
  persist/          sysctl.d, udev rules, backup.json
  output/           Table/JSON/Markdown formatters
  sysfs/            Low-level sysfs/procfs I/O
  platform/         Distro detection, privilege checks
  benchmark/        Disk I/O and network speed tests
  watch/            Live ANSI terminal dashboard
```

## Design Principles

- **No gopsutil** — reads sysfs/procfs directly for accuracy and zero deps
- **No viper/bubbletea** — minimal dependency tree
- **Detection is read-only** — never writes during diagnose/suggest
- **Guard everything** — all sysfs reads use `Exists()` checks, never crash on missing paths
- **Profile-aware** — output and suggestions adapt to laptop/desktop/server context

## License

MIT

---

Architected by **Kannan** and built with ❤️ with CC
