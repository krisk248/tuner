package tune

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/krisk248/tuner/internal/detect"
	"github.com/krisk248/tuner/internal/profile"
)

// Change represents a single tuning change.
type Change struct {
	Subsystem   string
	Parameter   string
	OldValue    string
	NewValue    string
	Path        string // sysfs/procfs path (for backup)
	ApplyFunc   func() error
}

// Result holds the outcome of applying a change.
type Result struct {
	Change Change
	Error  error
}

// Engine computes and applies tuning changes.
type Engine struct {
	Profile profile.Profile
	DryRun  bool
}

// NewEngine creates a tuning engine for the given profile.
func NewEngine(p profile.Profile) *Engine {
	return &Engine{Profile: p}
}

// ComputeChanges determines all changes needed to reach the target profile.
func (e *Engine) ComputeChanges() []Change {
	var changes []Change

	changes = append(changes, computeCPUChanges(e.Profile.Values)...)
	changes = append(changes, computeMemoryChanges(e.Profile.Values)...)
	changes = append(changes, computeStorageChanges(e.Profile.Values)...)
	changes = append(changes, computeNetworkChanges(e.Profile.Values)...)

	// Skip power tuning if TLP is active
	powerInfo := detect.DetectPower()
	if e.Profile.Values.SkipIfTLP && powerInfo.TLPActive {
		yellow := color.New(color.FgYellow)
		yellow.Println("Warning: TLP is active. Skipping power-related tuning.")
	}

	return changes
}

// Apply executes all computed changes.
func (e *Engine) Apply(changes []Change, auto bool) []Result {
	var results []Result

	for _, c := range changes {
		if !auto {
			fmt.Printf("  %s: %s → %s ... ", c.Parameter, c.OldValue, c.NewValue)
		}

		err := c.ApplyFunc()
		r := Result{Change: c, Error: err}
		results = append(results, r)

		if !auto {
			if err != nil {
				color.Red("FAILED (%v)", err)
			} else {
				color.Green("OK")
			}
		}
	}

	return results
}

// Backup returns a map of path -> original value for all changes.
func Backup(changes []Change) map[string]string {
	backup := make(map[string]string)
	for _, c := range changes {
		if c.Path != "" {
			backup[c.Path] = c.OldValue
		}
	}
	return backup
}
