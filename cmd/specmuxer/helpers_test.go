package main

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bagaking/specmuxer/pkg/domain/session"
	"github.com/bagaking/specmuxer/pkg/runtime/tmux"
	"github.com/bagaking/specmuxer/pkg/telemetry/stats"
)

func TestAllTerminalNil(t *testing.T) {
	if allTerminal() {
		t.Fatal("expected empty call to allTerminal to return false")
	}
	if allTerminal(nil) {
		t.Fatal("expected nil file to be treated as non-terminal")
	}
}

func TestAllTerminalWithPipe(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	if allTerminal(r) {
		t.Fatal("expected pipe reader to be non-terminal")
	}
	if allTerminal(w) {
		t.Fatal("expected pipe writer to be non-terminal")
	}
}

func TestFilterActiveSessions(t *testing.T) {
	records := []session.SessionRecord{
		{ID: "running", Status: session.StatusRunning},
		{ID: "idle", Status: session.StatusIdle},
		{ID: "stopped", Status: session.StatusStopped},
	}
	active := filterActiveSessions(records)
	if len(active) != 2 {
		t.Fatalf("expected 2 active sessions, got %d", len(active))
	}
	if active[0].ID != "running" || active[1].ID != "idle" {
		t.Fatalf("unexpected active order: %+v", active)
	}
}

func TestPromptYesNo(t *testing.T) {
	in := strings.NewReader("y\n")
	var out bytes.Buffer
	ok, err := promptYesNo(in, &out, "? ")
	if err != nil {
		t.Fatalf("prompt: %v", err)
	}
	if !ok {
		t.Fatal("expected confirmation")
	}

	in2 := strings.NewReader("\n")
	out.Reset()
	ok, err = promptYesNo(in2, &out, "? ")
	if err != nil {
		t.Fatalf("prompt empty: %v", err)
	}
	if ok {
		t.Fatal("expected rejection on empty input")
	}
}

func TestFetchSessionsWithLivenessMarksMissingTmuxAsStopped(t *testing.T) {
	now := time.Date(2025, time.October, 18, 12, 0, 0, 0, time.UTC)
	records := []session.SessionRecord{
		{
			ID:           "sess-missing",
			ProjectID:    "proj",
			Tool:         "codex",
			Status:       session.StatusRunning,
			LastOutputAt: &now,
			Tmux: session.TmuxMetadata{
				Session: "specmuxer_proj_sess-missing",
				Socket:  "/tmp/specmuxer.sock",
			},
		},
	}
	collector := stats.NewCollector(stats.WithClock(func() time.Time { return now }))
	tmuxClient := tmux.New(tmux.WithRunner(&livenessRunnerStub{exitCode: 1}))
	fetch := fetchSessionsWithLiveness(context.Background(), &sessionListerStub{records: records}, collector, tmuxClient, "/tmp/default.sock")

	got, err := fetch()
	if err != nil {
		t.Fatalf("fetchSessionsWithLiveness() error = %v, want nil", err)
	}
	if len(got) != 1 || got[0].ID != "sess-missing" {
		t.Fatalf("fetchSessionsWithLiveness() = %#v, want sess-missing record", got)
	}

	snapshot := collector.BuildSnapshot(got)
	if snapshot.Totals.Stopped != 1 {
		t.Fatalf("BuildSnapshot() stopped total = %d, want 1", snapshot.Totals.Stopped)
	}
	if snapshot.Projects[0].Sessions[0].TmuxAlive {
		t.Fatalf("BuildSnapshot() tmux alive = true, want false")
	}
}

func TestComputeLivenessReusesTmuxProbeResults(t *testing.T) {
	records := []session.SessionRecord{
		{
			ID:        "sess-1",
			ProjectID: "proj",
			Tmux: session.TmuxMetadata{
				Session: "specmuxer_proj_shared",
				Socket:  "specmuxer.sock",
			},
		},
		{
			ID:        "sess-2",
			ProjectID: "proj",
			Tmux: session.TmuxMetadata{
				Session: "specmuxer_proj_shared",
				Socket:  "specmuxer.sock",
			},
		},
	}
	runner := &livenessCountingRunnerStub{exitCode: 0}
	tmuxClient := tmux.New(tmux.WithRunner(runner))

	got := computeLiveness(context.Background(), tmuxClient, records, "default.sock")

	if runner.calls != 1 {
		t.Fatalf("tmux probe calls = %d, want 1", runner.calls)
	}
	if !got["sess-1"].Alive || !got["sess-2"].Alive {
		t.Fatalf("expected both sessions alive, got %#v", got)
	}
}

func TestComputeLivenessDoesNotShareDistinctTmuxProbes(t *testing.T) {
	records := []session.SessionRecord{
		{
			ID:        "sess-1",
			ProjectID: "proj",
			Tmux: session.TmuxMetadata{
				Session: "specmuxer_proj_shared",
				Socket:  "specmuxer-a.sock",
			},
		},
		{
			ID:        "sess-2",
			ProjectID: "proj",
			Tmux: session.TmuxMetadata{
				Session: "specmuxer_proj_shared",
				Socket:  "specmuxer-b.sock",
			},
		},
		{
			ID:        "sess-3",
			ProjectID: "proj",
			Tmux: session.TmuxMetadata{
				Session: "specmuxer_proj_other",
				Socket:  "specmuxer-a.sock",
			},
		},
	}
	runner := &livenessCountingRunnerStub{exitCode: 0}
	tmuxClient := tmux.New(tmux.WithRunner(runner))

	got := computeLiveness(context.Background(), tmuxClient, records, "default.sock")

	if runner.calls != 3 {
		t.Fatalf("tmux probe calls = %d, want 3", runner.calls)
	}
	for _, rec := range records {
		if !got[rec.ID].Alive {
			t.Fatalf("expected %s alive, got %#v", rec.ID, got[rec.ID])
		}
	}
}

func TestComputeLivenessDoesNotCacheAcrossCalls(t *testing.T) {
	records := []session.SessionRecord{
		{
			ID:        "sess-1",
			ProjectID: "proj",
			Tmux: session.TmuxMetadata{
				Session: "specmuxer_proj_shared",
				Socket:  "specmuxer.sock",
			},
		},
	}
	runner := &livenessCountingRunnerStub{exitCode: 0}
	tmuxClient := tmux.New(tmux.WithRunner(runner))

	first := computeLiveness(context.Background(), tmuxClient, records, "default.sock")
	second := computeLiveness(context.Background(), tmuxClient, records, "default.sock")

	if runner.calls != 2 {
		t.Fatalf("tmux probe calls after two computeLiveness calls = %d, want 2", runner.calls)
	}
	if !first["sess-1"].Alive || !second["sess-1"].Alive {
		t.Fatalf("expected both calls to report alive, got first=%#v second=%#v", first, second)
	}
}

type sessionListerStub struct {
	records []session.SessionRecord
	err     error
}

func (s *sessionListerStub) ListSessions() ([]session.SessionRecord, error) {
	if s.err != nil {
		return nil, s.err
	}
	return append([]session.SessionRecord(nil), s.records...), nil
}

type livenessRunnerStub struct {
	exitCode int
	err      error
}

func (s *livenessRunnerStub) Run(_ context.Context, _ []string, _ map[string]string) (string, string, int, error) {
	if s.err != nil {
		return "", "", -1, s.err
	}
	return "", "", s.exitCode, nil
}

type livenessCountingRunnerStub struct {
	exitCode int
	calls    int
}

func (s *livenessCountingRunnerStub) Run(_ context.Context, _ []string, _ map[string]string) (string, string, int, error) {
	s.calls++
	return "", "", s.exitCode, nil
}
