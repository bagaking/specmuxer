package logs

import (
	"context"
	"time"
)

// Rotator periodically prunes old log segments using Manager.Prune.
type Rotator struct {
	manager *Manager
}

// NewRotator constructs a Rotator.
func NewRotator(manager *Manager) *Rotator {
	return &Rotator{manager: manager}
}

// Run executes the pruning loop until the context is cancelled.
func (r *Rotator) Run(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		interval = time.Hour
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := r.manager.Prune(); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
