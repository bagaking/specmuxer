package run

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/bagaking/specmuxer/pkg/adapters"
	"github.com/bagaking/specmuxer/pkg/domain/session"
	"github.com/bagaking/specmuxer/pkg/orchestrator/config"
	"github.com/bagaking/specmuxer/pkg/runtime/storage"
	"github.com/bagaking/specmuxer/pkg/runtime/tmux"
	"github.com/bagaking/specmuxer/pkg/telemetry/logs"
	"github.com/bagaking/specmuxer/pkg/telemetry/stats"
)

// Store abstracts YAMLStore persistence used by the service.
type Store interface {
	SaveSession(record *session.SessionRecord) error
	LoadSession(id string) (*session.SessionRecord, error)
	DeleteSession(id string) error
	ListSessions() ([]session.SessionRecord, error)
	SaveStats(stats *storage.Stats) error
}

// Service coordinates session lifecycle operations.
type Service struct {
	cfg        *config.Config
	store      Store
	logWriter  *LogWriter
	registry   *adapters.Registry
	tmux       *tmux.Client
	clock      func() time.Time
	socketPath string
}

// BuildSessionName returns the tmux session name used for a SpecMuxer session.
func BuildSessionName(projectID, sessionID string) string {
	return fmt.Sprintf("specmuxer_%s_%s", projectID, sessionID)
}

// ServiceConfig bundles dependencies for Service.
type ServiceConfig struct {
	Config     *config.Config
	Store      Store
	LogManager *logs.Manager
	Registry   *adapters.Registry
	Tmux       *tmux.Client
	Clock      func() time.Time
	SocketPath string
}

// NewService constructs a Service.
func NewService(cfg ServiceConfig) *Service {
	clock := cfg.Clock
	if clock == nil {
		clock = time.Now
	}
	socket := cfg.SocketPath
	if socket == "" && cfg.Config != nil {
		socket = filepath.Join(os.TempDir(), fmt.Sprintf("specmuxer-%s.sock", cfg.Config.ProjectID))
	}

	return &Service{
		cfg:        cfg.Config,
		store:      cfg.Store,
		logWriter:  NewLogWriter(cfg.LogManager, clock),
		registry:   cfg.Registry,
		tmux:       cfg.Tmux,
		clock:      clock,
		socketPath: socket,
	}
}

// LaunchRequest describes a session launch invocation.
type LaunchRequest struct {
	Adapter   string
	HumanName string
	Env       map[string]string
	Args      []string
}

// Launch starts a new session using the configured adapter.
func (s *Service) Launch(ctx context.Context, req LaunchRequest) (string, error) {
	if s.cfg == nil {
		return "", errors.New("service not initialized")
	}
	if strings.TrimSpace(req.Adapter) == "" {
		return "", errors.New("adapter is required")
	}

	overrides := convertOverrides(s.cfg, req.Adapter)
	def, err := s.registry.Resolve(req.Adapter, overrides)
	if err != nil {
		return "", err
	}

	if len(def.Start.Exec) == 0 {
		return "", fmt.Errorf("adapter %q missing start command", req.Adapter)
	}

	sessionID := generateSessionID()
	now := s.clock()

	sessionName := BuildSessionName(s.cfg.ProjectID, sessionID)
	command := append([]string(nil), def.Start.Exec...)
	if len(req.Args) > 0 {
		command = append(command, req.Args...)
	}
	env := mergeEnv(def.Start.Env, req.Env)

	if err := s.tmux.EnsureSession(ctx, tmux.EnsureSessionOptions{
		Session:    sessionName,
		Socket:     s.socketPath,
		WindowName: def.Start.Description,
		Command:    command,
		Env:        env,
		WorkingDir: def.Start.WorkingDir,
	}); err != nil {
		return "", err
	}

	record := session.SessionRecord{
		ID:           sessionID,
		ProjectID:    s.cfg.ProjectID,
		Tool:         strings.ToLower(req.Adapter),
		HumanName:    req.HumanName,
		CreatedAt:    now,
		LastOutputAt: &now,
		Status:       session.StatusRunning,
		Tmux: session.TmuxMetadata{
			Session: sessionName,
			Socket:  s.socketPath,
		},
		Env: env,
	}
	session.SetUserKilled(&record, false)

	if err := s.store.SaveSession(&record); err != nil {
		return "", fmt.Errorf("save session: %w", err)
	}

	if s.logWriter != nil {
		_ = s.logWriter.Write(sessionID, "session created using adapter %s", req.Adapter)
	}

	if err := s.refreshStats(now); err != nil {
		return "", err
	}

	return sessionID, nil
}

// Sessions returns the persisted session records.
func (s *Service) Sessions() ([]session.SessionRecord, error) {
	return s.store.ListSessions()
}

// SocketPath returns the tmux socket path used by the service.
func (s *Service) SocketPath() string {
	return s.socketPath
}

func (s *Service) refreshStats(now time.Time) error {
	records, err := s.store.ListSessions()
	if err != nil {
		return fmt.Errorf("list sessions: %w", err)
	}

	collector := stats.NewCollector(stats.WithClock(func() time.Time { return now }))
	snapshot := collector.BuildSnapshot(records)

	payload := &storage.Stats{
		CollectedAt:    snapshot.CollectedAt,
		ActiveSessions: snapshot.Totals.Active,
		IdleSessions:   snapshot.Totals.Idle,
	}
	return s.store.SaveStats(payload)
}

func convertOverrides(cfg *config.Config, adapter string) map[string]adapters.LifecycleCommand {
	if cfg == nil || cfg.AdapterOverrides == nil {
		return nil
	}
	src, ok := cfg.AdapterOverrides[strings.ToLower(adapter)]
	if !ok {
		return nil
	}

	result := make(map[string]adapters.LifecycleCommand)

	if len(src.Start) > 0 {
		result["start"] = adapters.LifecycleCommand{
			Exec:       append([]string(nil), src.Start...),
			WorkingDir: src.WorkingDir,
			Env:        cloneMap(src.Env),
			Timeout:    time.Duration(src.TimeoutSec) * time.Second,
		}
	}
	if len(src.Resume) > 0 {
		result["resume"] = adapters.LifecycleCommand{
			Exec:       append([]string(nil), src.Resume...),
			WorkingDir: src.WorkingDir,
			Env:        cloneMap(src.Env),
			Timeout:    time.Duration(src.TimeoutSec) * time.Second,
		}
	}
	if len(src.Health) > 0 {
		result["health"] = adapters.LifecycleCommand{
			Exec:       append([]string(nil), src.Health...),
			WorkingDir: src.WorkingDir,
			Env:        cloneMap(src.Env),
			Timeout:    time.Duration(src.TimeoutSec) * time.Second,
		}
	}
	if len(src.Stop) > 0 {
		result["stop"] = adapters.LifecycleCommand{
			Exec:       append([]string(nil), src.Stop...),
			WorkingDir: src.WorkingDir,
			Env:        cloneMap(src.Env),
			Timeout:    time.Duration(src.TimeoutSec) * time.Second,
		}
	}
	if len(src.ExtractState) > 0 {
		result["extract_state"] = adapters.LifecycleCommand{
			Exec:       append([]string(nil), src.ExtractState...),
			WorkingDir: src.WorkingDir,
			Env:        cloneMap(src.Env),
			Timeout:    time.Duration(src.TimeoutSec) * time.Second,
		}
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

func cloneMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func generateSessionID() string {
	return uuid.NewString()
}
