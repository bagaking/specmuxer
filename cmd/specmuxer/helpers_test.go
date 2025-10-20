package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/bagaking/specmuxer/pkg/domain/session"
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
