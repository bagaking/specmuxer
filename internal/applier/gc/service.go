package gc

import (
	"context"
	"fmt"

	"github.com/bagaking/specmuxer/pkg/domain/session"
	"github.com/bagaking/specmuxer/pkg/runtime/tmux"
)

// Store abstraction mirrors YAMLStore subset for GC operations.
type Store interface {
	ListSessions() ([]session.SessionRecord, error)
	DeleteSession(id string) error
}

// Service identifies orphaned sessions and optionally adopts or prunes them.
type Service struct {
	store Store
	tmux  *tmux.Client
}

// ServiceConfig bundles dependencies.
type ServiceConfig struct {
	Store Store
	Tmux  *tmux.Client
}

// Summary reports GC findings.
type Summary struct {
	Attempted int
	Orphans   []string
	Removed   []string
}

// NewService constructs a GC service.
func NewService(cfg ServiceConfig) *Service {
	return &Service{
		store: cfg.Store,
		tmux:  cfg.Tmux,
	}
}

// DryRun detects orphaned sessions without modifying state.
func (s *Service) DryRun(ctx context.Context) (Summary, error) {
	records, err := s.store.ListSessions()
	if err != nil {
		return Summary{}, err
	}

	var summary Summary
	for _, rec := range records {
		summary.Attempted++
		exists, err := s.tmux.HasSession(ctx, rec.Tmux.Session, rec.Tmux.Socket)
		if err != nil {
			return summary, fmt.Errorf("check tmux session %s: %w", rec.Tmux.Session, err)
		}
		if !exists {
			summary.Orphans = append(summary.Orphans, rec.ID)
		}
	}
	return summary, nil
}

// Adopt removes orphaned session records.
func (s *Service) Adopt(ctx context.Context) (Summary, error) {
	summary, err := s.DryRun(ctx)
	if err != nil {
		return summary, err
	}
	for _, id := range summary.Orphans {
		if err := s.store.DeleteSession(id); err == nil {
			summary.Removed = append(summary.Removed, id)
		}
	}
	return summary, nil
}
