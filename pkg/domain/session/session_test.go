package session

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestSessionRecordYAMLRoundTrip(t *testing.T) {
	createdAt := time.Date(2025, time.October, 18, 12, 0, 0, 0, time.UTC)
	lastOutput := createdAt.Add(2 * time.Minute)

	input := SessionRecord{
		ID:        "sess-123",
		ProjectID: "proj-abc",
		Tool:      "codex",
		HumanName: "Daily Driver",
		CreatedAt: createdAt,
		LastOutputAt: func() *time.Time {
			ts := lastOutput
			return &ts
		}(),
		Status: StatusIdle,
		UserKilled: func() *bool {
			b := false
			return &b
		}(),
		Tmux: TmuxMetadata{
			Session: "specmuxer:p-abc-codex",
			Window:  "1",
			Socket:  "/tmp/tmux-1234",
		},
		AdapterState: map[string]any{
			"cursor": 42,
		},
		Env: map[string]string{
			"OPENAI_API_KEY": "sk-abc",
		},
		LogFiles: []LogPointer{
			{
				Path:      "/tmp/logs/log-2025-10-18T12.log",
				SizeBytes: 2048,
				SegmentDate: Date{
					Time: createdAt,
				},
				Checksum: func() *string {
					val := "sha256:deadbeef"
					return &val
				}(),
			},
		},
		ResumeHooks: []ResumeCommand{
			{
				Type:         ResumeCommandPrompt,
				Value:        "继续工作",
				DelaySeconds: 5,
			},
		},
		Error: func() *string {
			msg := ""
			return &msg
		}(),
	}

	data, err := yaml.Marshal(&input)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var output SessionRecord
	if err := yaml.Unmarshal(data, &output); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if output.ID != input.ID {
		t.Fatalf("round-trip ID mismatch: got %q want %q", output.ID, input.ID)
	}

	if output.Tmux.Session != input.Tmux.Session {
		t.Fatalf("round-trip tmux.session mismatch: got %q want %q", output.Tmux.Session, input.Tmux.Session)
	}

	if output.Env["OPENAI_API_KEY"] != "sk-abc" {
		t.Fatalf("env round-trip failed: got %q", output.Env["OPENAI_API_KEY"])
	}

	if len(output.LogFiles) != 1 || output.LogFiles[0].Path != input.LogFiles[0].Path {
		t.Fatalf("log files round-trip mismatch: %#v", output.LogFiles)
	}

	if output.Status != StatusIdle {
		t.Fatalf("status mismatch: got %q want %q", output.Status, StatusIdle)
	}

	if output.ResumeHooks[0].DelaySeconds != 5 {
		t.Fatalf("resume delay mismatch: %d", output.ResumeHooks[0].DelaySeconds)
	}
}
