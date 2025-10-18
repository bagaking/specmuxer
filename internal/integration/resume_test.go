package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/bagaking/specmuxer/pkg/adapters"
	"github.com/bagaking/specmuxer/pkg/domain/session"
	"github.com/bagaking/specmuxer/pkg/orchestrator/config"
	"github.com/bagaking/specmuxer/pkg/orchestrator/resume"
	"github.com/bagaking/specmuxer/pkg/orchestrator/run"
	"github.com/bagaking/specmuxer/pkg/runtime/storage"
	"github.com/bagaking/specmuxer/pkg/runtime/tmux"
	"github.com/bagaking/specmuxer/pkg/telemetry/logs"
)

func TestResumeAll(t *testing.T) {
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
			Start:  []string{"sh", "-lc", "sleep 1"},
			Resume: []string{"sh", "-lc", "sleep 1"},
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

	runSvc := run.NewService(run.ServiceConfig{
		Config:     cfg,
		Store:      store,
		LogManager: logManager,
		Registry:   registry,
		Tmux:       tmuxClient,
	})

	ctx := context.Background()
	launchID, err := runSvc.Launch(ctx, run.LaunchRequest{Adapter: "codex", HumanName: "primary"})
	if err != nil {
		t.Fatalf("Launch: %v", err)
	}

	record, err := store.LoadSession(launchID)
	if err != nil {
		t.Fatalf("LoadSession: %v", err)
	}
	record.Status = session.StatusStopped
	record.LastOutputAt = ptrTime(time.Now().Add(-5 * time.Minute))
	f := false
	record.UserKilled = &f
	record.Tmux.Session = record.Tmux.Session + "-resume"
	if err := store.SaveSession(record); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}

	resumeSvc := resume.NewService(resume.ServiceConfig{
		Config:     cfg,
		Store:      store,
		LogManager: logManager,
		Registry:   registry,
		Tmux:       tmuxClient,
	})

	summary, err := resumeSvc.Resume(ctx, resume.Options{All: true})
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if summary.Succeeded != 1 {
		t.Fatalf("expected 1 resume success, got %+v", summary)
	}

	updated, err := store.LoadSession(launchID)
	if err != nil {
		t.Fatalf("LoadSession: %v", err)
	}
	if updated.Status != session.StatusRunning {
		t.Fatalf("expected status running, got %s", updated.Status)
	}

	// Mark as user_killed and ensure resume skips.
	trueVal := true
	updated.Status = session.StatusStopped
	updated.UserKilled = &trueVal
	if err := store.SaveSession(updated); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}

	summary, err = resumeSvc.Resume(ctx, resume.Options{All: true})
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if summary.Succeeded != 0 || summary.Skipped == 0 {
		t.Fatalf("expected skipped resume for user_killed: %+v", summary)
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
