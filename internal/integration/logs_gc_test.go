package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	doctor "github.com/bagaking/specmuxer/internal/applier/doctor"
	gc "github.com/bagaking/specmuxer/internal/applier/gc"
	"github.com/bagaking/specmuxer/pkg/adapters"
	"github.com/bagaking/specmuxer/pkg/domain/session"
	"github.com/bagaking/specmuxer/pkg/orchestrator/config"
	"github.com/bagaking/specmuxer/pkg/orchestrator/run"
	"github.com/bagaking/specmuxer/pkg/runtime/storage"
	"github.com/bagaking/specmuxer/pkg/runtime/tmux"
	"github.com/bagaking/specmuxer/pkg/telemetry/logs"
)

func TestLogsGcDoctorFlow(t *testing.T) {
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
			Start: []string{"sh", "-lc", "sleep 1"},
		},
	}

	store, err := storage.NewYAMLStore(cfg.Paths().SessionsDir, cfg.Paths().StatsPath)
	if err != nil {
		t.Fatalf("NewYAMLStore: %v", err)
	}

	logManager, err := logs.NewManager(cfg.Paths().LogsDir, logs.WithRedactionRules([]string{"sk-[A-Za-z0-9]+"}))
	if err != nil {
		t.Fatalf("logs.NewManager: %v", err)
	}

	registry := adapters.NewRegistry()
	tmuxClient := tmux.New()

	runSvc := run.NewService(run.ServiceConfig{
		Config:     cfg,
		Store:      store,
		LogManager: logManager,
		Registry:   registry,
		Tmux:       tmuxClient,
	})

	ctx := context.Background()
	sessionID, err := runSvc.Launch(ctx, run.LaunchRequest{Adapter: "codex", HumanName: "logs"})
	if err != nil {
		t.Fatalf("Launch: %v", err)
	}

	writer := run.NewLogWriter(logManager, time.Now)
	if err := writer.Write(sessionID, "token sk-secret"); err != nil {
		t.Fatalf("writer.Write: %v", err)
	}

	reader := logs.NewReader(cfg.Paths().LogsDir, "sk-[A-Za-z0-9]+")
	lines, err := reader.LastLines(sessionID, 1)
	if err != nil {
		t.Fatalf("LastLines: %v", err)
	}
	if len(lines) != 1 || lines[0] == "" || lines[0] == "token sk-secret" {
		t.Fatalf("expected redacted log line, got %#v", lines)
	}

	record, err := store.LoadSession(sessionID)
	if err != nil {
		t.Fatalf("LoadSession: %v", err)
	}
	record.Status = session.StatusStopped
	record.Tmux.Session = record.Tmux.Session + "-gc"
	if err := store.SaveSession(record); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}

	gcService := gc.NewService(gc.ServiceConfig{
		Store: store,
		Tmux:  tmuxClient,
	})

	gcSummary, err := gcService.DryRun(ctx)
	if err != nil {
		t.Fatalf("gc dry run: %v", err)
	}
	if len(gcSummary.Orphans) == 0 {
		t.Fatalf("expected orphan sessions detected")
	}

	report := doctor.Run(doctor.Config{})
	if report.OverallStatus == doctor.StatusFail {
		t.Fatalf("expected doctor report to pass")
	}
}
