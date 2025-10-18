package integration

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/bagaking/specmuxer/pkg/adapters"
	"github.com/bagaking/specmuxer/pkg/orchestrator/config"
	"github.com/bagaking/specmuxer/pkg/orchestrator/run"
	"github.com/bagaking/specmuxer/pkg/runtime/storage"
	"github.com/bagaking/specmuxer/pkg/runtime/tmux"
	"github.com/bagaking/specmuxer/pkg/telemetry/logs"
	"github.com/bagaking/specmuxer/pkg/telemetry/stats"
)

func TestRunStatusLifecycle(t *testing.T) {
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux binary not available")
	}

	tempDir := t.TempDir()
	workspace := filepath.Join(tempDir, "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	cfg, err := config.Load(workspace, "")
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	cfg.AdapterOverrides = map[string]config.AdapterOverride{
		"codex": {
			Start: []string{"sh", "-lc", "printf 'specmuxer'"},
		},
	}

	store, err := storage.NewYAMLStore(cfg.Paths().SessionsDir, cfg.Paths().StatsPath)
	if err != nil {
		t.Fatalf("NewYAMLStore: %v", err)
	}

	logManager, err := logs.NewManager(cfg.Paths().LogsDir)
	if err != nil {
		t.Fatalf("logs.NewManager: %v", err)
	}

	registry := adapters.NewRegistry()
	tmuxClient := tmux.New()

	service := run.NewService(run.ServiceConfig{
		Config:     cfg,
		Store:      store,
		LogManager: logManager,
		Registry:   registry,
		Tmux:       tmuxClient,
		Clock:      func() time.Time { return time.Now() },
	})

	sessionID, err := service.Launch(context.Background(), run.LaunchRequest{
		Adapter:   "codex",
		HumanName: "integration",
		Env: map[string]string{
			"EXAMPLE": "1",
		},
	})
	if err != nil {
		t.Fatalf("Launch: %v", err)
	}
	if sessionID == "" {
		t.Fatalf("expected session ID")
	}

	records, err := store.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(records) == 0 {
		t.Fatalf("expected stored session record")
	}

	collector := stats.NewCollector()
	snapshot := collector.BuildSnapshot(records)
	if snapshot.Totals.Active+snapshot.Totals.Idle == 0 {
		t.Fatalf("expected active or idle sessions in totals")
	}

	jsonData, err := stats.FormatJSON(snapshot)
	if err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}

	var parsed stats.JSONStatus
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("decode formatted JSON: %v", err)
	}
	if len(parsed.Projects) == 0 {
		t.Fatalf("expected project summary")
	}
	if parsed.Projects[0].Sessions[0].ID != sessionID {
		t.Fatalf("expected session %s in formatted JSON", sessionID)
	}
}
