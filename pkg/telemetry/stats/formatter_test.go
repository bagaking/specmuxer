package stats

import (
	"encoding/json"
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
				Sessions: []SessionSummary{
					{
						ID:           "sess-1",
						ProjectID:    "proj-1",
						Tool:         "codex",
						HumanName:    "Primary",
						Status:       session.StatusRunning,
						LastOutputAt: &last,
					},
				},
			},
		},
		Top: []SessionSummary{
			{
				ID:        "sess-1",
				ProjectID: "proj-1",
				Tool:      "codex",
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
	if len(parsed.Projects[0].Sessions) != 1 || parsed.Projects[0].Sessions[0].ID != "sess-1" {
		t.Fatalf("unexpected sessions: %#v", parsed.Projects[0].Sessions)
	}
	if parsed.Projects[0].Sessions[0].LastOutputAt == "" {
		t.Fatalf("expected last output timestamp")
	}
}
