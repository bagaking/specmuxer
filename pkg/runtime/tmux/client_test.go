package tmux

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestEnsureSessionSkipsCreationWhenPresent(t *testing.T) {
	runner := &stubRunner{
		responses: []runResponse{
			{exitCode: 0},
		},
	}
	client := New(WithRunner(runner))

	err := client.EnsureSession(context.Background(), EnsureSessionOptions{
		Session: "specmuxer:p-1",
		Socket:  "/tmp/socket",
	})
	if err != nil {
		t.Fatalf("EnsureSession: %v", err)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("expected 1 tmux invocation, got %d", len(runner.calls))
	}
	assertArgsEqual(t, runner.calls[0].Args, []string{"-S", "/tmp/socket", "has-session", "-t", "specmuxer:p-1"})
}

func TestEnsureSessionCreatesWhenMissing(t *testing.T) {
	runner := &stubRunner{
		responses: []runResponse{
			{exitCode: 1},
			{exitCode: 0},
		},
	}
	client := New(WithRunner(runner))

	err := client.EnsureSession(context.Background(), EnsureSessionOptions{
		Session:    "specmuxer:p-1",
		Socket:     "/tmp/socket",
		WindowName: "main",
		Command:    []string{"bash", "-lc", "echo hi"},
		WorkingDir: "/workspace",
		Env: map[string]string{
			"FOO": "BAR",
		},
	})
	if err != nil {
		t.Fatalf("EnsureSession: %v", err)
	}
	if len(runner.calls) != 2 {
		t.Fatalf("expected 2 tmux invocations, got %d", len(runner.calls))
	}
	assertArgsEqual(t, runner.calls[0].Args, []string{"-S", "/tmp/socket", "has-session", "-t", "specmuxer:p-1"})
	assertArgsEqual(t, runner.calls[1].Args, []string{
		"-S", "/tmp/socket", "new-session", "-d", "-s", "specmuxer:p-1",
		"-n", "main", "-c", "/workspace", "--", "bash", "-lc", "echo hi",
	})
	if len(runner.calls[1].Env) != 1 || runner.calls[1].Env["FOO"] != "BAR" {
		t.Fatalf("expected env to be forwarded, got %#v", runner.calls[1].Env)
	}
}

func TestEnsureSessionFailsOnUnexpectedExitCode(t *testing.T) {
	runner := &stubRunner{
		responses: []runResponse{
			{exitCode: 2, stderr: "boom"},
		},
	}
	client := New(WithRunner(runner))

	err := client.EnsureSession(context.Background(), EnsureSessionOptions{
		Session: "specmuxer:p-1",
	})
	if err == nil || !contains(err.Error(), "has-session exit 2") {
		t.Fatalf("expected exit code error, got %v", err)
	}
}

func TestEnsureSessionPropagatesCreationError(t *testing.T) {
	runner := &stubRunner{
		responses: []runResponse{
			{exitCode: 1},
			{exitCode: 1, stderr: "cannot create"},
		},
	}
	client := New(WithRunner(runner))

	err := client.EnsureSession(context.Background(), EnsureSessionOptions{
		Session: "specmuxer:p-1",
	})
	if err == nil || !contains(err.Error(), "new-session exit 1") {
		t.Fatalf("expected create error, got %v", err)
	}
}

func TestEnsureSessionPropagatesRunnerError(t *testing.T) {
	expected := errors.New("tmux missing")
	runner := &stubRunner{
		responses: []runResponse{
			{err: expected},
		},
	}
	client := New(WithRunner(runner))

	err := client.EnsureSession(context.Background(), EnsureSessionOptions{
		Session: "specmuxer:p-1",
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected runner error propagation, got %v", err)
	}
}

func TestNewWindowUsesSessionAndName(t *testing.T) {
	runner := &stubRunner{
		responses: []runResponse{
			{exitCode: 0},
		},
	}
	client := New(WithRunner(runner))

	err := client.NewWindow(context.Background(), NewWindowOptions{
		Session: "specmuxer:p-1",
		Name:    "status",
		Command: []string{"top"},
	})
	if err != nil {
		t.Fatalf("NewWindow: %v", err)
	}
	assertArgsEqual(t, runner.calls[0].Args, []string{
		"new-window", "-t", "specmuxer:p-1", "-n", "status", "--", "top",
	})
}

func TestNewWindowPropagatesError(t *testing.T) {
	runner := &stubRunner{
		responses: []runResponse{
			{exitCode: 1, stderr: "failure"},
		},
	}
	client := New(WithRunner(runner))

	err := client.NewWindow(context.Background(), NewWindowOptions{
		Session: "specmuxer:p-1",
	})
	if err == nil || !contains(err.Error(), "new-window exit 1") {
		t.Fatalf("expected error, got %v", err)
	}
}

func TestSendKeysRequiresKeys(t *testing.T) {
	client := New()
	err := client.SendKeys(context.Background(), SendKeysOptions{})
	if err == nil || !contains(err.Error(), "at least one key") {
		t.Fatalf("expected missing keys error, got %v", err)
	}
}

func TestSendKeysLiteralFlag(t *testing.T) {
	runner := &stubRunner{
		responses: []runResponse{
			{exitCode: 0},
		},
	}
	client := New(WithRunner(runner))

	err := client.SendKeys(context.Background(), SendKeysOptions{
		Target:  "specmuxer:p-1.0",
		Keys:    []string{"export FOO=BAR", "Enter"},
		Literal: true,
	})
	if err != nil {
		t.Fatalf("SendKeys: %v", err)
	}

	assertArgsEqual(t, runner.calls[0].Args, []string{
		"send-keys", "-t", "specmuxer:p-1.0", "-l", "export FOO=BAR", "Enter",
	})
}

func TestKillSession(t *testing.T) {
	runner := &stubRunner{
		responses: []runResponse{
			{exitCode: 0},
		},
	}
	client := New(WithRunner(runner))

	err := client.KillSession(context.Background(), KillSessionOptions{
		Session: "specmuxer:p-1",
	})
	if err != nil {
		t.Fatalf("KillSession: %v", err)
	}
	assertArgsEqual(t, runner.calls[0].Args, []string{"kill-session", "-t", "specmuxer:p-1"})
}

func TestKillSessionError(t *testing.T) {
	runner := &stubRunner{
		responses: []runResponse{
			{exitCode: 1, stderr: "no such session"},
		},
	}
	client := New(WithRunner(runner))

	err := client.KillSession(context.Background(), KillSessionOptions{
		Session: "missing",
	})
	if err == nil || !contains(err.Error(), "kill-session exit 1") {
		t.Fatalf("expected kill error, got %v", err)
	}
}

type stubRunner struct {
	responses []runResponse
	calls     []runCall
}

type runResponse struct {
	stdout   string
	stderr   string
	exitCode int
	err      error
}

type runCall struct {
	Args []string
	Env  map[string]string
}

func (s *stubRunner) Run(_ context.Context, args []string, env map[string]string) (string, string, int, error) {
	if len(s.responses) == 0 {
		return "", "", 0, errors.New("unexpected call")
	}
	resp := s.responses[0]
	s.responses = s.responses[1:]
	if env == nil {
		env = map[string]string{}
	}
	s.calls = append(s.calls, runCall{
		Args: append([]string(nil), args...),
		Env:  cloneMap(env),
	})
	return resp.stdout, resp.stderr, resp.exitCode, resp.err
}

func cloneMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func assertArgsEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("unexpected args length: got %d want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("args mismatch at %d: got %q want %q (full=%v)", i, got[i], want[i], got)
		}
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
