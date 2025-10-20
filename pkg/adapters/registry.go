package adapters

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Registry maintains adapter definitions available to SpecMuxer.
type Registry struct {
	definitions map[string]Definition
}

// Definition describes lifecycle commands for an adapter.
type Definition struct {
	Name           string
	Start          LifecycleCommand
	Resume         *LifecycleCommand
	Health         *LifecycleCommand
	Stop           *LifecycleCommand
	ExtractState   *LifecycleCommand
	SupportsAttach bool
	Overrides      map[string]LifecycleCommand
}

// LifecycleCommand represents a shell command template.
type LifecycleCommand struct {
	Exec        []string
	WorkingDir  string
	Env         map[string]string
	Timeout     time.Duration
	OnFailure   FailurePolicy
	Description string
}

// FailurePolicy dictates how the orchestrator should react when a command fails.
type FailurePolicy string

const (
	// FailurePolicyAbort instructs the orchestrator to stop the workflow on error.
	FailurePolicyAbort FailurePolicy = "abort"
	// FailurePolicyRetry instructs the orchestrator to attempt the command again.
	FailurePolicyRetry FailurePolicy = "retry"
	// FailurePolicyFallback instructs the orchestrator to attempt a secondary strategy.
	FailurePolicyFallback FailurePolicy = "fallback"
)

// NewRegistry constructs a registry pre-populated with built-in adapters.
func NewRegistry() *Registry {
	r := &Registry{
		definitions: map[string]Definition{},
	}
	for _, def := range builtinAdapters() {
		r.Register(def)
	}
	return r
}

// Register inserts or replaces an adapter definition.
func (r *Registry) Register(def Definition) {
	if r.definitions == nil {
		r.definitions = map[string]Definition{}
	}
	key := strings.ToLower(def.Name)
	r.definitions[key] = def.Clone()
}

// Resolve returns the adapter definition with optional overrides applied.
func (r *Registry) Resolve(name string, overrides map[string]LifecycleCommand) (Definition, error) {
	if r == nil {
		return Definition{}, fmt.Errorf("adapter registry not initialized")
	}
	def, ok := r.definitions[strings.ToLower(name)]
	if !ok {
		return Definition{}, fmt.Errorf("adapter %q not found", name)
	}
	return def.WithOverrides(overrides), nil
}

// List returns all adapter definitions sorted by name.
func (r *Registry) List() []Definition {
	if r == nil {
		return nil
	}
	names := make([]string, 0, len(r.definitions))
	for name := range r.definitions {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([]Definition, 0, len(names))
	for _, name := range names {
		result = append(result, r.definitions[name].Clone())
	}
	return result
}

// Clone returns a deep copy of the adapter definition.
func (d Definition) Clone() Definition {
	def := Definition{
		Name:           d.Name,
		Start:          d.Start.Clone(),
		SupportsAttach: d.SupportsAttach,
	}
	if d.Resume != nil {
		resume := d.Resume.Clone()
		def.Resume = &resume
	}
	if d.Health != nil {
		health := d.Health.Clone()
		def.Health = &health
	}
	if d.Stop != nil {
		stop := d.Stop.Clone()
		def.Stop = &stop
	}
	if d.ExtractState != nil {
		extract := d.ExtractState.Clone()
		def.ExtractState = &extract
	}
	if len(d.Overrides) > 0 {
		def.Overrides = make(map[string]LifecycleCommand, len(d.Overrides))
		for k, v := range d.Overrides {
			cmd := v.Clone()
			def.Overrides[k] = cmd
		}
	}
	return def
}

// WithOverrides returns a clone with project-level overrides merged in.
func (d Definition) WithOverrides(overrides map[string]LifecycleCommand) Definition {
	if len(overrides) == 0 {
		return d.Clone()
	}
	merged := d.Clone()
	for key, cmd := range overrides {
		switch strings.ToLower(key) {
		case "start":
			merged.Start = cmd.Clone()
		case "resume":
			resume := cmd.Clone()
			merged.Resume = &resume
		case "health":
			health := cmd.Clone()
			merged.Health = &health
		case "stop":
			stop := cmd.Clone()
			merged.Stop = &stop
		case "extract_state", "extract-state":
			extract := cmd.Clone()
			merged.ExtractState = &extract
		default:
			if merged.Overrides == nil {
				merged.Overrides = make(map[string]LifecycleCommand)
			}
			clone := cmd.Clone()
			merged.Overrides[key] = clone
		}
	}
	return merged
}

// Clone returns a copy of the lifecycle command.
func (c LifecycleCommand) Clone() LifecycleCommand {
	cmd := LifecycleCommand{
		Exec:        append([]string(nil), c.Exec...),
		WorkingDir:  c.WorkingDir,
		Timeout:     c.Timeout,
		OnFailure:   c.OnFailure,
		Description: c.Description,
	}
	if len(c.Env) > 0 {
		cmd.Env = make(map[string]string, len(c.Env))
		for k, v := range c.Env {
			cmd.Env[k] = v
		}
	}
	return cmd
}

func builtinAdapters() []Definition {
	return []Definition{
		{
			Name: "codex",
			Start: LifecycleCommand{
				Exec:        []string{"codex"},
				Env:         map[string]string{"TERM": "xterm-256color"},
				OnFailure:   FailurePolicyRetry,
				Description: "Launch Codex interactive CLI session",
			},
			Resume: &LifecycleCommand{
				Exec:        []string{"codex", "resume", "--last"},
				OnFailure:   FailurePolicyFallback,
				Description: "Resume Codex session from stored state",
			},
			Health: &LifecycleCommand{
				Exec:        []string{"codex", "--version"},
				Description: "Verify Codex CLI availability",
			},
			// Codex CLI does not expose a standalone stop/snapshot command; rely on tmux lifecycle.
			SupportsAttach: true,
		},
		{
			Name: "claude",
			Start: LifecycleCommand{
				Exec:        []string{"claude", "workbench", "start"},
				Env:         map[string]string{"LANG": "en_US.UTF-8"},
				OnFailure:   FailurePolicyRetry,
				Description: "Launch Claude workbench session",
			},
			Resume: &LifecycleCommand{
				Exec:        []string{"claude", "workbench", "resume"},
				Description: "Resume Claude session with stored context",
			},
			Health: &LifecycleCommand{
				Exec:        []string{"claude", "workbench", "status"},
				Description: "Health check for Claude workbench",
			},
			Stop: &LifecycleCommand{
				Exec:        []string{"claude", "workbench", "stop"},
				Description: "Gracefully stop Claude workbench session",
			},
			ExtractState: &LifecycleCommand{
				Exec:        []string{"claude", "workbench", "export"},
				Description: "Export Claude session transcript",
			},
			SupportsAttach: true,
		},
	}
}
