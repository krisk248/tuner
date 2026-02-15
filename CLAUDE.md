# Tuner - Linux System Tuner CLI

## Project

Go CLI tool for Linux system diagnostics and tuning. Single static binary, no CGO.
Targets Fedora, Ubuntu, Arch, RHEL 8+ on amd64.

## Architecture

- `internal/sysfs/` - Low-level sysfs/procfs I/O with `Exists()` guards for kernel compat
- `internal/platform/` - Distro detection, kernel version, privilege checks
- `internal/detect/` - Read-only detection (8 subsystems). Shared across profiles
- `internal/profile/` - Hardcoded Go structs for tuning values. No YAML/config files
- `internal/tune/` - Write-side. Profile-dependent changes
- `internal/persist/` - sysctl.d, udev rules, backup.json
- `internal/output/` - Table/JSON/Markdown formatters
- `internal/benchmark/` - Disk I/O and network speedtest
- `internal/watch/` - Live ANSI terminal dashboard

## Key Rules

- Always guard sysfs reads with `sysfs.Exists()` - never crash on missing paths
- Detection is read-only, tuning is write-only. Don't mix them
- No gopsutil - read sysfs/procfs directly
- No viper/bubbletea - keep deps minimal

## Build

```
CGO_ENABLED=0 go build -o bin/tuner ./cmd/tuner
```

---

# Karpathy Guidelines

Behavioral guidelines to reduce common LLM coding mistakes.

## 1. Think Before Coding

- State assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.
- Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

Transform tasks into verifiable goals:
- "Add validation" -> "Write tests for invalid inputs, then make them pass"
- "Fix the bug" -> "Write a test that reproduces it, then make it pass"
- "Refactor X" -> "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] -> verify: [check]
2. [Step] -> verify: [check]
3. [Step] -> verify: [check]
```
