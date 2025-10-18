package stats

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/bagaking/specmuxer/pkg/domain/session"
)

// Streamer renders rolling status snapshots similar to `tmux top`.
type Streamer struct {
	collector *Collector
}

// NewStreamer constructs a Streamer using the provided collector.
func NewStreamer(collector *Collector) *Streamer {
	if collector == nil {
		collector = NewCollector()
	}
	return &Streamer{collector: collector}
}

// Stream continuously fetches session data, rendering it to the sink until the
// context is cancelled.
func (s *Streamer) Stream(ctx context.Context, interval time.Duration, fetch func() ([]session.SessionRecord, error), sink io.Writer) error {
	if interval <= 0 {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := s.renderOnce(fetch, sink); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (s *Streamer) renderOnce(fetch func() ([]session.SessionRecord, error), sink io.Writer) error {
	records, err := fetch()
	if err != nil {
		return err
	}
	snapshot := s.collector.BuildSnapshot(records)
	if _, err := fmt.Fprint(sink, FormatTable(snapshot, false)); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(sink, "---"); err != nil {
		return err
	}
	return nil
}
