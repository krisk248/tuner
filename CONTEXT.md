# Tuner — AI Context Guide

This file provides context for AI assistants working on this codebase.

## What This Is

A Linux system tuner CLI written in Go. It detects hardware state by reading sysfs/procfs directly (no gopsutil), recommends profile-based optimizations, and applies them. Single static binary, no CGO.

## Architecture Overview

```
cmd/tuner/main.go → cli.Execute()
                      ├── diagnose  → detect.*  → output.Formatter
                      ├── suggest   → detect.*  + profile.Values → diff
                      ├── apply     → tune.Engine → sysfs.Write*
                      ├── save      → persist.WriteSysctl/WriteUdev
                      ├── reset     → persist.RestoreBackup
                      ├── fix-power → systemctl stop/start
                      ├── profile   → profile.AutoDetect
                      ├── benchmark → benchmark.Disk/Network
                      └── watch     → watch.Dashboard (ANSI loop)
```

## Key Separation

- `detect/` is **read-only**. It reads system state. Never writes.
- `tune/` is **write-only**. It applies changes via `sysfs.Write*`.
- `profile/` defines **target values**. Hardcoded Go structs, no config files.
- `suggest` compares current state (detect) against target (profile) and shows the diff.
- `apply` uses `tune.Engine` to write the diff.

## Profile System

Three profiles: `laptop`, `desktop`, `server`. Auto-detected via:
1. Battery present → laptop
2. DMI chassis type → server/laptop
3. No display + multi-user.target → server
4. Fallback → desktop

Each profile has a `Values` struct defining target parameters for CPU governor, EPP, swappiness, dirty ratios, TCP settings, I/O scheduler, etc.

### Profile-Specific Power Manager Logic

- **Laptop**: Expects TLP. If TLP active, skip CPU suggestions (TLP manages governor/EPP). Ignores tuned.
- **Desktop**: No power manager expected. Always shows CPU suggestions. Never mentions TLP.
- **Server**: Expects tuned. If tuned active, skip CPU suggestions. Ignores TLP. Never shows battery or Wi-Fi.

## Detection Subsystems (detect/)

| File | Struct | What It Reads |
|------|--------|---------------|
| `cpu.go` | `CPUInfo` | Governor, EPP, turbo, frequencies, core count |
| `memory.go` | `MemoryInfo` | Swappiness, dirty ratios, THP, zswap, meminfo |
| `storage.go` | `StorageInfo` | Block devices, schedulers, rotational, type |
| `network.go` | `NetworkInfo` | TCP params, interfaces, Wi-Fi (iw), offloads (ethtool) |
| `power.go` | `PowerInfo` | Battery, AC, TLP/tuned/PPD service state |
| `services.go` | `ServiceInfo` | Boot time, failed units, slow services |
| `kernel.go` | `KernelInfo` | Version, cmdline |
| `gpu.go` | `GPUInfo` | DRM devices |
| `server.go` | `ServerInfo` | file-max, somaxconn, conntrack, port range, irqbalance |

Each has a `Detect*()` function and a `*Section()` function that converts to `output.Section`.

## sysfs Package

Low-level I/O for `/sys` and `/proc` paths.

- `sysfs.Exists(path)` — always guard reads with this
- `sysfs.ReadString(path)` — single trimmed line
- `sysfs.ReadInt(path)` — parse int
- `sysfs.ReadInt64(path)` — parse int64
- `sysfs.ReadBracketedValue(path)` — parse `[value]` from kernel format
- `sysfs.ReadFields(path)` — space-separated fields
- `sysfs.WriteString(path, value)` — write to sysfs

Path constants are in `sysfs/paths.go`.

## Output System

`output.Section` contains `[]output.Field` (key, value, status).
`output.Formatter` renders to table, JSON, or markdown.
`output.DiagMode` (laptop/desktop/server/auto) controls profile-aware filtering in section builders.

## Common Patterns

### Adding a new sysfs parameter:
1. Add path constant to `sysfs/paths.go`
2. Add field to relevant detect struct
3. Read it in `Detect*()` with `sysfs.Exists()` guard
4. Display it in `*Section()`
5. Add target value to `profile/values.go` and profile files
6. Add suggestion in `cli/suggest.go`
7. Add apply logic in `tune/changes_*.go`

### Adding a new subsystem:
1. Create `detect/newsubsystem.go` with struct, `Detect*()`, `*Section()`
2. Wire into `cli/diagnose.go`
3. Add profile values if tunable
4. Add suggestions if applicable

## Dependencies

Minimal by design:
- `cobra` — CLI framework
- `fatih/color` — terminal colors
- `progressbar` — benchmark progress
- `speedtest-go` — network benchmark
- No gopsutil, no viper, no bubbletea

## Build

```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.Version=..." -o bin/tuner ./cmd/tuner
```

Version is injected via `-ldflags` from git tags.

## Testing

```
go test ./...
go vet ./...
golangci-lint run ./...
```

Profile detection tests are in `profile/profile_test.go`.
Output formatter tests are in `output/`.
sysfs tests are in `sysfs/`.
