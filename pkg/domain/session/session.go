package session

import (
	"fmt"
	"time"
)

// SessionStatus represents lifecycle states persisted for a session record.
type SessionStatus string

const (
	StatusRunning SessionStatus = "running"
	StatusIdle    SessionStatus = "idle"
	StatusStopped SessionStatus = "stopped"
	StatusFailed  SessionStatus = "failed"
)

// ResumeCommandType indicates how a post-resume action should be applied.
type ResumeCommandType string

const (
	ResumeCommandPrompt ResumeCommandType = "prompt"
	ResumeCommandShell  ResumeCommandType = "shell"
)

// SessionRecord models the persisted session YAML document.
type SessionRecord struct {
	ID           string                 `yaml:"id"`
	ProjectID    string                 `yaml:"project_id"`
	Tool         string                 `yaml:"tool"`
	HumanName    string                 `yaml:"human_name,omitempty"`
	CreatedAt    time.Time              `yaml:"created_at"`
	LastOutputAt *time.Time             `yaml:"last_output_at,omitempty"`
	Status       SessionStatus          `yaml:"status"`
	UserKilled   *bool                  `yaml:"user_killed,omitempty"`
	Tmux         TmuxMetadata           `yaml:"tmux"`
	AdapterState map[string]any         `yaml:"adapter_state,omitempty"`
	Env          map[string]string      `yaml:"env,omitempty"`
	LogFiles     []LogPointer           `yaml:"log_files,omitempty"`
	ResumeHooks  []ResumeCommand        `yaml:"resume_hooks,omitempty"`
	Error        *string                `yaml:"error,omitempty"`
	Tags         map[string]string      `yaml:"tags,omitempty"`
	Metadata     map[string]interface{} `yaml:"metadata,omitempty"`
}

// TmuxMetadata stores tmux association for a session.
type TmuxMetadata struct {
	Session string `yaml:"session"`
	Window  string `yaml:"window,omitempty"`
	Socket  string `yaml:"socket,omitempty"`
}

// ResumeCommand describes post-resume actions.
type ResumeCommand struct {
	Type         ResumeCommandType `yaml:"type"`
	Value        string            `yaml:"value"`
	DelaySeconds int               `yaml:"delay_seconds,omitempty"`
}

// LogPointer represents a log segment associated with the session.
type LogPointer struct {
	Path        string  `yaml:"path"`
	SizeBytes   int64   `yaml:"size_bytes"`
	SegmentDate Date    `yaml:"segment_date"`
	Checksum    *string `yaml:"checksum,omitempty"`
}

// Date encodes day-level granularity in YAML.
type Date struct {
	time.Time
}

// MarshalYAML renders the date as YYYY-MM-DD.
func (d Date) MarshalYAML() (any, error) {
	if d.IsZero() {
		return "", nil
	}
	return d.Format(time.DateOnly), nil
}

// UnmarshalYAML parses the YYYY-MM-DD value into Date.
func (d *Date) UnmarshalYAML(value func(any) error) error {
	var raw string
	if err := value(&raw); err != nil {
		return fmt.Errorf("parse date: %w", err)
	}
	if raw == "" {
		d.Time = time.Time{}
		return nil
	}
	ts, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		return fmt.Errorf("parse date %q: %w", raw, err)
	}
	d.Time = ts
	return nil
}
