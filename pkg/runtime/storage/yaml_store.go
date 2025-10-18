package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/bagaking/specmuxer/pkg/domain/session"
)

const filePerm = 0o600

// ErrNotFound indicates the requested record does not exist.
var ErrNotFound = errors.New("not found")

// YAMLStore persists sessions and statistics using YAML documents.
type YAMLStore struct {
	sessionsDir string
	statsPath   string
}

// NewYAMLStore constructs a store using the provided directories.
func NewYAMLStore(sessionsDir, statsPath string) (*YAMLStore, error) {
	if sessionsDir == "" || statsPath == "" {
		return nil, errors.New("sessionsDir and statsPath are required")
	}
	if err := ensureDirExists(sessionsDir); err != nil {
		return nil, err
	}
	if err := ensureFileExists(statsPath); err != nil {
		return nil, err
	}
	return &YAMLStore{
		sessionsDir: sessionsDir,
		statsPath:   statsPath,
	}, nil
}

// SaveSession writes the session record to disk.
func (s *YAMLStore) SaveSession(record *session.SessionRecord) error {
	if record == nil {
		return errors.New("session record is nil")
	}
	if record.ID == "" {
		return errors.New("session id is required")
	}

	path := filepath.Join(s.sessionsDir, record.ID+".yml")
	payload, err := yaml.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal session %s: %w", record.ID, err)
	}
	if err := os.WriteFile(path, payload, filePerm); err != nil {
		return fmt.Errorf("write session %s: %w", record.ID, err)
	}
	return nil
}

// LoadSession reads a session record from disk.
func (s *YAMLStore) LoadSession(id string) (*session.SessionRecord, error) {
	if id == "" {
		return nil, errors.New("session id is required")
	}
	path := filepath.Join(s.sessionsDir, id+".yml")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read session %s: %w", id, err)
	}
	var record session.SessionRecord
	if err := yaml.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("parse session %s: %w", id, err)
	}
	return &record, nil
}

// DeleteSession removes a session record.
func (s *YAMLStore) DeleteSession(id string) error {
	if id == "" {
		return errors.New("session id is required")
	}
	path := filepath.Join(s.sessionsDir, id+".yml")
	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrNotFound
		}
		return fmt.Errorf("delete session %s: %w", id, err)
	}
	return nil
}

// ListSessions returns all session records sorted by creation time descending.
func (s *YAMLStore) ListSessions() ([]session.SessionRecord, error) {
	entries, err := os.ReadDir(s.sessionsDir)
	if err != nil {
		return nil, fmt.Errorf("read sessions dir: %w", err)
	}
	var records []session.SessionRecord
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".yml")
		record, err := s.LoadSession(id)
		if err != nil {
			return nil, err
		}
		records = append(records, *record)
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})
	return records, nil
}

// SaveStats writes aggregate statistics.
func (s *YAMLStore) SaveStats(stats *Stats) error {
	if stats == nil {
		return errors.New("stats payload is nil")
	}
	payload, err := yaml.Marshal(stats)
	if err != nil {
		return fmt.Errorf("marshal stats: %w", err)
	}
	if err := os.WriteFile(s.statsPath, payload, filePerm); err != nil {
		return fmt.Errorf("write stats: %w", err)
	}
	return nil
}

// LoadStats retrieves aggregate statistics. Returns zero-value Stats when file empty.
func (s *YAMLStore) LoadStats() (*Stats, error) {
	data, err := os.ReadFile(s.statsPath)
	if errors.Is(err, os.ErrNotExist) {
		return &Stats{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read stats: %w", err)
	}
	if len(data) == 0 {
		return &Stats{}, nil
	}
	var stats Stats
	if err := yaml.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("parse stats: %w", err)
	}
	return &stats, nil
}

// Stats represents aggregated telemetry persisted in stats.yml.
type Stats struct {
	CollectedAt       time.Time  `yaml:"collected_at"`
	ActiveSessions    int        `yaml:"active_sessions"`
	IdleSessions      int        `yaml:"idle_sessions"`
	LastGC            *time.Time `yaml:"last_gc,omitempty"`
	ResumeSuccessRate float64    `yaml:"resume_success_rate"`
	TopLatencyP95Ms   int        `yaml:"top_latency_p95_ms"`
	LogVolumeBytes    int64      `yaml:"log_volume_bytes"`
}

func ensureDirExists(path string) error {
	info, err := os.Stat(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return fmt.Errorf("sessions dir %q missing", path)
	case err != nil:
		return fmt.Errorf("stat dir %q: %w", path, err)
	case !info.IsDir():
		return fmt.Errorf("%q is not a directory", path)
	}
	return nil
}

func ensureFileExists(path string) error {
	info, err := os.Stat(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		if err := os.WriteFile(path, []byte{}, filePerm); err != nil {
			return fmt.Errorf("create file %q: %w", path, err)
		}
	case err != nil:
		return fmt.Errorf("stat file %q: %w", path, err)
	default:
		if info.Mode().Perm() != filePerm {
			if err := os.Chmod(path, filePerm); err != nil {
				return fmt.Errorf("chmod file %q: %w", path, err)
			}
		}
	}
	return nil
}
