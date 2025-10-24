package stats

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/bagaking/specmuxer/pkg/domain/session"
)

func TestFormatJSONSerializesSnapshot(t *testing.T) {
	now := time.Date(2025, time.October, 18, 10, 0, 0, 0, time.UTC)
	last := now.Add(-30 * time.Second)

	snapshot := Snapshot{
		CollectedAt: now,
		Totals: Totals{
			Active:  1,
			Idle:    2,
			Stopped: 1,
			Failed:  0,
		},
		Projects: []ProjectSummary{
			{
				ProjectID: "proj-1",
				RootPath:  "/workspace",
				Sessions: []SessionSummary{
					{
						ID:           "sess-1",
						ProjectID:    "proj-1",
						Tool:         "codex",
						HumanName:    "Primary",
						Status:       session.StatusRunning,
						LastOutputAt: &last,
						TmuxAlive:    true,
					},
				},
			},
		},
		Top: []SessionSummary{
			{
				ID:        "sess-1",
				ProjectID: "proj-1",
				Tool:      "codex",
				TmuxAlive: true,
			},
		},
	}

	data, err := FormatJSON(snapshot)
	if err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}

	var parsed JSONStatus
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if parsed.Totals.Active != 1 || parsed.Totals.Idle != 2 {
		t.Fatalf("unexpected totals: %#v", parsed.Totals)
	}
	if len(parsed.Projects) != 1 || parsed.Projects[0].ProjectID != "proj-1" {
		t.Fatalf("unexpected projects: %#v", parsed.Projects)
	}
	if parsed.Projects[0].RootPath != "/workspace" {
		t.Fatalf("expected root path, got %#v", parsed.Projects[0].RootPath)
	}
	if len(parsed.Projects[0].Sessions) != 1 || parsed.Projects[0].Sessions[0].ID != "sess-1" {
		t.Fatalf("unexpected sessions: %#v", parsed.Projects[0].Sessions)
	}
	if parsed.Projects[0].Sessions[0].LastOutputAt == "" {
		t.Fatalf("expected last output timestamp")
	}
	if !parsed.Projects[0].Sessions[0].TmuxAlive {
		t.Fatalf("expected tmuxAlive true")
	}
}

func TestFormatTableWideIncludesRootSessionIDAndTmuxState(t *testing.T) {
	last := time.Date(2025, time.October, 18, 9, 30, 0, 0, time.UTC)
	snapshot := Snapshot{
		CollectedAt: time.Date(2025, time.October, 18, 10, 0, 0, 0, time.UTC),
		Totals: Totals{
			Active:  1,
			Stopped: 1,
		},
		Projects: []ProjectSummary{
			{
				ProjectID: "proj-1",
				RootPath:  "workspace",
				Sessions: []SessionSummary{
					{
						ID:           "sess-1",
						Tool:         "codex",
						HumanName:    "Primary",
						Status:       session.StatusRunning,
						LastOutputAt: &last,
						TmuxAlive:    false,
					},
					{
						ID:        "sess-2",
						Tool:      "claude",
						Status:    session.StatusStopped,
						TmuxAlive: true,
					},
				},
			},
		},
	}

	table := FormatTable(snapshot, true)
	for _, want := range []string{
		"Collected:  2025-10-18T10:00:00Z",
		"Totals:     active 1  idle 0  stopped 1  failed 0",
		"proj-1 (workspace)",
		"Primary (sess-1)",
		"2025-10-18T09:30:00Z",
		"missing",
		"sess-2",
		"present",
	} {
		if !strings.Contains(table, want) {
			t.Fatalf("expected table to contain %q, got:\n%s", want, table)
		}
	}
}
