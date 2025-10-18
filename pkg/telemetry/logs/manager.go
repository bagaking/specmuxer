package logs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
)

const (
	defaultMaxBytes      int64 = 200 * 1024 * 1024 // 200 MB
	defaultRetention           = 30 * 24 * time.Hour
	defaultDirPerm             = 0o700
	defaultFilePerm            = 0o600
	redactionReplacement       = "[REDACTED]"
)

// Option configures a Manager.
type Option func(*Manager) error

// Manager coordinates log rotation and redaction policies.
type Manager struct {
	rootDir   string
	maxBytes  int64
	retention time.Duration
	redactors []*regexp.Regexp
	now       func() time.Time
}

// NewManager constructs a Manager with defaults.
func NewManager(rootDir string, opts ...Option) (*Manager, error) {
	if rootDir == "" {
		return nil, errors.New("rootDir is required")
	}
	if err := os.MkdirAll(rootDir, defaultDirPerm); err != nil {
		return nil, fmt.Errorf("create log directory %q: %w", rootDir, err)
	}

	m := &Manager{
		rootDir:   rootDir,
		maxBytes:  defaultMaxBytes,
		retention: defaultRetention,
		redactors: []*regexp.Regexp{},
		now:       time.Now,
	}

	for _, opt := range opts {
		if err := opt(m); err != nil {
			return nil, err
		}
	}

	return m, nil
}

// WithMaxBytes overrides the rotation size threshold.
func WithMaxBytes(maxBytes int64) Option {
	return func(m *Manager) error {
		if maxBytes <= 0 {
			return fmt.Errorf("maxBytes must be positive, got %d", maxBytes)
		}
		m.maxBytes = maxBytes
		return nil
	}
}

// WithRetention overrides the log retention window.
func WithRetention(window time.Duration) Option {
	return func(m *Manager) error {
		if window <= 0 {
			return fmt.Errorf("retention window must be positive, got %s", window)
		}
		m.retention = window
		return nil
	}
}

// WithRedactionRules configures regex replacements to scrub sensitive data.
func WithRedactionRules(patterns []string) Option {
	return func(m *Manager) error {
		m.redactors = m.redactors[:0]
		for _, pattern := range patterns {
			if pattern == "" {
				continue
			}
			re, err := regexp.Compile(pattern)
			if err != nil {
				return fmt.Errorf("compile redaction rule %q: %w", pattern, err)
			}
			m.redactors = append(m.redactors, re)
		}
		return nil
	}
}

// WithClock injects a custom clock (primarily for testing).
func WithClock(now func() time.Time) Option {
	return func(m *Manager) error {
		if now == nil {
			return errors.New("clock function cannot be nil")
		}
		m.now = now
		return nil
	}
}

// Append writes the supplied line to the session log, applying redaction and rotation.
// It returns the path of the log file that received the entry.
func (m *Manager) Append(sessionID string, line string) (string, error) {
	if sessionID == "" {
		return "", errors.New("sessionID is required")
	}
	cleanID := sanitizeSegment(sessionID)
	if cleanID == "" {
		cleanID = "session"
	}

	now := m.now()
	line = m.applyRedaction(line)
	if !strings.HasSuffix(line, "\n") {
		line += "\n"
	}

	path, err := m.resolveSegmentPath(cleanID, now, int64(len(line)))
	if err != nil {
		return "", err
	}

	file, created, err := openAppend(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if created {
		if err := os.Chmod(path, defaultFilePerm); err != nil {
			return "", fmt.Errorf("chmod log %q: %w", path, err)
		}
	}

	if _, err := file.WriteString(line); err != nil {
		return "", fmt.Errorf("write log %q: %w", path, err)
	}

	return path, nil
}

// Prune removes log files older than the configured retention window.
func (m *Manager) Prune() error {
	if m.retention <= 0 {
		return nil
	}

	cutoff := m.now().Add(-m.retention)
	entries, err := os.ReadDir(m.rootDir)
	if err != nil {
		return fmt.Errorf("read log directory: %w", err)
	}

	var joined error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(m.rootDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			joined = errors.Join(joined, fmt.Errorf("stat %q: %w", path, err))
			continue
		}
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(path); err != nil {
				joined = errors.Join(joined, fmt.Errorf("remove %q: %w", path, err))
			}
		}
	}
	return joined
}

func (m *Manager) resolveSegmentPath(cleanID string, ts time.Time, incoming int64) (string, error) {
	datePart := ts.Format("20060102")
	base := fmt.Sprintf("%s-%s", cleanID, datePart)
	path := filepath.Join(m.rootDir, base+".log")

	info, err := os.Stat(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return path, nil
	case err != nil:
		return "", fmt.Errorf("stat log %q: %w", path, err)
	default:
		if m.maxBytes > 0 && info.Size()+incoming > m.maxBytes {
			// Derive a unique suffix using the current time; ensure deterministic ordering.
			suffix := fmt.Sprintf("-%d", ts.UnixNano())
			path = filepath.Join(m.rootDir, base+suffix+".log")
		}
		return path, nil
	}
}

func (m *Manager) applyRedaction(line string) string {
	if len(m.redactors) == 0 {
		return line
	}
	redacted := line
	for _, re := range m.redactors {
		redacted = re.ReplaceAllString(redacted, redactionReplacement)
	}
	return redacted
}

func sanitizeSegment(sessionID string) string {
	var b strings.Builder
	for _, r := range sessionID {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	return strings.Trim(b.String(), "_")
}

func openAppend(path string) (*os.File, bool, error) {
	_, err := os.Stat(path)
	created := errors.Is(err, os.ErrNotExist)
	file, openErr := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, defaultFilePerm)
	if openErr != nil {
		return nil, created, fmt.Errorf("open log %q: %w", path, openErr)
	}
	return file, created, nil
}
