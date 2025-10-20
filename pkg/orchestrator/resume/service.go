package resume

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bagaking/specmuxer/pkg/adapters"
	"github.com/bagaking/specmuxer/pkg/domain/session"
	"github.com/bagaking/specmuxer/pkg/orchestrator/config"
	"github.com/bagaking/specmuxer/pkg/orchestrator/run"
	"github.com/bagaking/specmuxer/pkg/runtime/storage"
	"github.com/bagaking/specmuxer/pkg/runtime/tmux"
	"github.com/bagaking/specmuxer/pkg/telemetry/logs"
	"github.com/bagaking/specmuxer/pkg/telemetry/stats"
)

// Store represents the persistence layer needed for resume operations.
type Store interface {
	SaveSession(record *session.SessionRecord) error
	LoadSession(id string) (*session.SessionRecord, error)
	ListSessions() ([]session.SessionRecord, error)
	SaveStats(stats *storage.Stats) error
}

// Service coordinates resume workflows across sessions.
type Service struct {
	cfg        *config.Config
	store      Store
	logWriter  *run.LogWriter
	registry   *adapters.Registry
	tmux       *tmux.Client
	clock      func() time.Time
	socketPath string
}

// ServiceConfig contains dependencies for constructing the service.
type ServiceConfig struct {
	Config     *config.Config
	Store      Store
	LogManager *logs.Manager
	Registry   *adapters.Registry
	Tmux       *tmux.Client
	Clock      func() time.Time
}

// Options control resume execution.
type Options struct {
	All        bool
	SessionIDs []string
	DryRun     bool
}

// Summary captures resume results.
type Summary struct {
	Attempted       int
	Succeeded       int
	Skipped         int
	SucceededIDs    []string
	SkippedSessions []SkippedSession
}

// SkippedSession captures a skipped resume attempt and its reason.
type SkippedSession struct {
	ID     string
	Reason string
}

// NewService constructs a resume Service.
func NewService(cfg ServiceConfig) *Service {
	clock := cfg.Clock
	if clock == nil {
		clock = time.Now
	}

	socket := filepath.Join(os.TempDir(), fmt.Sprintf("specmuxer-%s.sock", cfg.Config.ProjectID))

	return &Service{
		cfg:        cfg.Config,
		store:      cfg.Store,
		logWriter:  run.NewLogWriter(cfg.LogManager, clock),
		registry:   cfg.Registry,
		tmux:       cfg.Tmux,
		clock:      clock,
		socketPath: socket,
	}
}

// Resume attempts to relaunch eligible sessions based on options provided.
func (s *Service) Resume(ctx context.Context, opts Options) (Summary, error) {
	if s.cfg == nil {
		return Summary{}, errors.New("service not initialized")
	}

	records, err := s.store.ListSessions()
	if err != nil {
		return Summary{}, err
	}

	var targets []session.SessionRecord
	if opts.All {
		targets = records
	} else {
		idSet := make(map[string]struct{}, len(opts.SessionIDs))
		for _, id := range opts.SessionIDs {
			idSet[strings.TrimSpace(id)] = struct{}{}
		}
		for _, rec := range records {
			if _, ok := idSet[rec.ID]; ok {
				targets = append(targets, rec)
			}
		}
	}

	if len(targets) == 0 {
		return Summary{}, nil
	}

	var summary Summary
	for _, rec := range targets {
		summary.Attempted++

		def, err := s.registry.Resolve(rec.Tool, convertOverrides(s.cfg, rec.Tool))
		if err != nil {
			summary.Skipped++
			summary.SkippedSessions = append(summary.SkippedSessions, SkippedSession{ID: rec.ID, Reason: fmt.Sprintf("adapter resolve failed: %v", err)})
			continue
		}

		socket := rec.Tmux.Socket
		if socket == "" {
			socket = s.socketPath
		}

		safeName := run.BuildSessionName(s.cfg.ProjectID, rec.ID)
		sessionName := safeName

		alive := false
		seen := map[string]struct{}{}
		candidates := []string{}
		addCandidate := func(name string) {
			if name == "" {
				return
			}
			if _, ok := seen[name]; ok {
				return
			}
			seen[name] = struct{}{}
			candidates = append(candidates, name)
		}
		addCandidate(rec.Tmux.Session)
		addCandidate(strings.ReplaceAll(rec.Tmux.Session, ":", "_"))
		addCandidate(safeName)

		for _, candidate := range candidates {
			ok, err := s.tmux.HasSession(ctx, candidate, socket)
			if err != nil {
				alive = false
				continue
			}
			if ok {
				alive = true
				sessionName = candidate
				break
			}
		}

		if rec.Tmux.Session != "" && len(candidates) == 0 {
			alive = false
		}

		if eligible, reason := EvaluateEligibility(rec, def, alive); !eligible {
			summary.Skipped++
			summary.SkippedSessions = append(summary.SkippedSessions, SkippedSession{ID: rec.ID, Reason: reason})
			continue
		}

		if opts.DryRun {
			summary.SkippedSessions = append(summary.SkippedSessions, SkippedSession{ID: rec.ID, Reason: "dry-run"})
			continue
		}

		cmd := def.ResumeCommand()
		env := mergeEnv(cmd.Env, rec.Env)

		if err := s.tmux.EnsureSession(ctx, tmux.EnsureSessionOptions{
			Session:    sessionName,
			Socket:     socket,
			WindowName: cmd.Description,
			Command:    cmd.Exec,
			Env:        env,
			WorkingDir: cmd.WorkingDir,
		}); err != nil {
			summary.Skipped++
			summary.SkippedSessions = append(summary.SkippedSessions, SkippedSession{ID: rec.ID, Reason: err.Error()})
			continue
		}

		now := s.clock()
		rec.Status = session.StatusRunning
		rec.LastOutputAt = &now
		session.SetUserKilled(&rec, false)
		rec.Tmux.Socket = socket
		rec.Tmux.Session = sessionName

		if err := s.store.SaveSession(&rec); err != nil {
			summary.Skipped++
			summary.SkippedSessions = append(summary.SkippedSessions, SkippedSession{ID: rec.ID, Reason: fmt.Sprintf("persist session: %v", err)})
			continue
		}

		if s.logWriter != nil {
			_ = s.logWriter.Write(rec.ID, "session resumed")
		}

		summary.Succeeded++
		summary.SucceededIDs = append(summary.SucceededIDs, rec.ID)
	}

	if err := s.updateStats(summary); err != nil {
		return summary, err
	}

	return summary, nil
}

func (s *Service) updateStats(summary Summary) error {
	records, err := s.store.ListSessions()
	if err != nil {
		return err
	}
	collector := stats.NewCollector()
	snapshot := collector.BuildSnapshot(records)

	rate := stats.CalculateResumeSuccessRate(summary.Attempted, summary.Succeeded)
	return s.store.SaveStats(stats.ToStorageStats(snapshot, rate))
}

func convertOverrides(cfg *config.Config, adapter string) map[string]adapters.LifecycleCommand {
	if cfg == nil || cfg.AdapterOverrides == nil {
		return nil
	}
	ov, ok := cfg.AdapterOverrides[strings.ToLower(adapter)]
	if !ok {
		return nil
	}
	result := make(map[string]adapters.LifecycleCommand)
	if len(ov.Start) > 0 {
		result["start"] = adapters.LifecycleCommand{Exec: append([]string(nil), ov.Start...)}
	}
	if len(ov.Resume) > 0 {
		result["resume"] = adapters.LifecycleCommand{Exec: append([]string(nil), ov.Resume...)}
	}
	if len(ov.Health) > 0 {
		result["health"] = adapters.LifecycleCommand{Exec: append([]string(nil), ov.Health...)}
	}
	if len(ov.Stop) > 0 {
		result["stop"] = adapters.LifecycleCommand{Exec: append([]string(nil), ov.Stop...)}
	}
	if len(ov.ExtractState) > 0 {
		result["extract_state"] = adapters.LifecycleCommand{Exec: append([]string(nil), ov.ExtractState...)}
	}
	return result
}

func mergeEnv(base map[string]string, overrides map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range overrides {
		result[k] = v
	}
	return result
}
