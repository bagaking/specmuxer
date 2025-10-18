package stats

import (
	"testing"
	"time"

	"github.com/bagaking/specmuxer/pkg/domain/session"
)

func TestBuildSnapshotAggregatesTotals(t *testing.T) {
	now := time.Date(2025, time.October, 18, 12, 0, 0, 0, time.UTC)
	idle := now.Add(-3 * time.Minute)
	last := now.Add(-30 * time.Second)

	trueVal := true

	records := []session.SessionRecord{
		{
			ID:           "sess-running",
			ProjectID:    "proj-1",
			Tool:         "codex",
			HumanName:    "Run",
			Status:       session.StatusRunning,
			LastOutputAt: &last,
		},
		{
			ID:           "sess-idle",
			ProjectID:    "proj-1",
			Tool:         "claude",
			HumanName:    "Idle",
			Status:       session.StatusRunning,
			LastOutputAt: &idle,
		},
		{
			ID:        "sess-stopped",
			ProjectID: "proj-2",
			Status:    session.StatusStopped,
		},
		{
			ID:         "sess-failed",
			ProjectID:  "proj-2",
			Status:     session.StatusFailed,
			UserKilled: &trueVal,
		},
	}

	collector := NewCollector(
		WithClock(func() time.Time { return now }),
		WithIdleThreshold(2*time.Minute),
	)

	snapshot := collector.BuildSnapshot(records)
	if !snapshot.CollectedAt.Equal(now) {
		t.Fatalf("expected collected time %v, got %v", now, snapshot.CollectedAt)
	}
	if snapshot.Totals.Active != 1 {
		t.Fatalf("expected 1 active, got %d", snapshot.Totals.Active)
	}
	if snapshot.Totals.Idle != 1 {
		t.Fatalf("expected 1 idle, got %d", snapshot.Totals.Idle)
	}
	if snapshot.Totals.Stopped != 1 {
		t.Fatalf("expected 1 stopped, got %d", snapshot.Totals.Stopped)
	}
	if snapshot.Totals.Failed != 1 {
		t.Fatalf("expected 1 failed, got %d", snapshot.Totals.Failed)
	}
	if len(snapshot.Projects) != 2 {
		t.Fatalf("expected 2 project summaries, got %d", len(snapshot.Projects))
	}
}

func TestTopOrderingRespectsLimit(t *testing.T) {
	now := time.Date(2025, time.October, 18, 12, 0, 0, 0, time.UTC)
	records := []session.SessionRecord{
		{ID: "a", ProjectID: "proj", LastOutputAt: ptrTime(now.Add(-10 * time.Second))},
		{ID: "b", ProjectID: "proj", LastOutputAt: ptrTime(now.Add(-2 * time.Second))},
		{ID: "c", ProjectID: "proj", LastOutputAt: ptrTime(now.Add(-30 * time.Second))},
	}

	collector := NewCollector(
		WithClock(func() time.Time { return now }),
		WithTopLimit(2),
	)

	top := collector.Top(records, 0)
	if len(top) != 2 {
		t.Fatalf("expected top limit 2, got %d", len(top))
	}
	if top[0].ID != "b" || top[1].ID != "a" {
		t.Fatalf("unexpected ordering: %#v", top)
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
