package logs

// Manager coordinates log rotation and redaction policies.
type Manager struct{}

// NewManager constructs a Manager with defaults.
func NewManager() *Manager {
	return &Manager{}
}
