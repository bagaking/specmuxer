package run

import (
	"fmt"
	"time"

	"github.com/bagaking/specmuxer/pkg/telemetry/logs"
)

// LogWriter appends session lifecycle entries to the log manager.
type LogWriter struct {
	manager *logs.Manager
	clock   func() time.Time
}

// NewLogWriter constructs a writer using the supplied log manager.
func NewLogWriter(manager *logs.Manager, clock func() time.Time) *LogWriter {
	if clock == nil {
		clock = time.Now
	}
	return &LogWriter{
		manager: manager,
		clock:   clock,
	}
}

// Write appends a timestamped line to the session log.
func (w *LogWriter) Write(sessionID string, format string, args ...any) error {
	if w == nil || w.manager == nil {
		return fmt.Errorf("log writer not initialized")
	}
	line := fmt.Sprintf(format, args...)
	payload := fmt.Sprintf("[%s] %s", w.clock().Format(time.RFC3339), line)
	if _, err := w.manager.Append(sessionID, payload); err != nil {
		return err
	}
	return nil
}
