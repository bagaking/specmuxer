package tmux

import "context"

// Client describes operations supported by underlying tmux interactions.
type Client interface {
	EnsureSession(ctx context.Context, name string, opts EnsureSessionOptions) error
}

// EnsureSessionOptions captures session creation details.
type EnsureSessionOptions struct {
	Socket  string
	Command []string
	Env     map[string]string
}
