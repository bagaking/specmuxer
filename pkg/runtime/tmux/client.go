package tmux

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// Runner executes tmux commands.
type Runner interface {
	Run(ctx context.Context, args []string, env map[string]string) (stdout string, stderr string, exitCode int, err error)
}

// ClientOption configures a Client.
type ClientOption func(*Client)

// Client exposes high-level tmux operations used by the orchestrator.
type Client struct {
	runner        Runner
	defaultSocket string
}

// New constructs a Client with optional configuration.
func New(opts ...ClientOption) *Client {
	c := &Client{
		runner: &ExecRunner{Bin: "tmux"},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithRunner overrides the command runner used by the client.
func WithRunner(r Runner) ClientOption {
	return func(c *Client) {
		if r != nil {
			c.runner = r
		}
	}
}

// WithDefaultSocket sets the socket used when an operation omits one.
func WithDefaultSocket(socket string) ClientOption {
	return func(c *Client) {
		c.defaultSocket = socket
	}
}

// EnsureSessionOptions controls how sessions are created.
type EnsureSessionOptions struct {
	Session    string
	Socket     string
	WindowName string
	Command    []string
	Env        map[string]string
	WorkingDir string
}

// EnsureSession creates the tmux session if it does not already exist.
func (c *Client) EnsureSession(ctx context.Context, opts EnsureSessionOptions) error {
	session := strings.TrimSpace(opts.Session)
	if session == "" {
		return errors.New("session name is required")
	}
	socket := c.resolveSocket(opts.Socket)

	args := c.baseArgs(socket, "has-session", "-t", session)
	_, stderr, exitCode, err := c.runner.Run(ctx, args, nil)
	if err != nil {
		return fmt.Errorf("tmux has-session: %w", err)
	}
	if exitCode == 0 {
		return nil
	}
	if exitCode != 1 {
		return fmt.Errorf("tmux has-session exit %d: %s", exitCode, strings.TrimSpace(stderr))
	}

	createArgs := c.baseArgs(socket, "new-session", "-d", "-s", session)
	if opts.WindowName != "" {
		createArgs = append(createArgs, "-n", opts.WindowName)
	}
	if opts.WorkingDir != "" {
		createArgs = append(createArgs, "-c", opts.WorkingDir)
	}
	if len(opts.Command) > 0 {
		createArgs = append(createArgs, "--")
		createArgs = append(createArgs, opts.Command...)
	}

	_, stderr, exitCode, err = c.runner.Run(ctx, createArgs, opts.Env)
	if err != nil {
		return fmt.Errorf("tmux new-session: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("tmux new-session exit %d: %s", exitCode, strings.TrimSpace(stderr))
	}
	return nil
}

// NewWindowOptions controls window creation.
type NewWindowOptions struct {
	Session    string
	Socket     string
	Name       string
	Command    []string
	Env        map[string]string
	WorkingDir string
}

// NewWindow creates a window within the target session.
func (c *Client) NewWindow(ctx context.Context, opts NewWindowOptions) error {
	session := strings.TrimSpace(opts.Session)
	if session == "" {
		return errors.New("session name is required")
	}
	socket := c.resolveSocket(opts.Socket)

	args := c.baseArgs(socket, "new-window", "-t", session)
	if opts.Name != "" {
		args = append(args, "-n", opts.Name)
	}
	if opts.WorkingDir != "" {
		args = append(args, "-c", opts.WorkingDir)
	}
	if len(opts.Command) > 0 {
		args = append(args, "--")
		args = append(args, opts.Command...)
	}

	_, stderr, exitCode, err := c.runner.Run(ctx, args, opts.Env)
	if err != nil {
		return fmt.Errorf("tmux new-window: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("tmux new-window exit %d: %s", exitCode, strings.TrimSpace(stderr))
	}
	return nil
}

// KillSessionOptions controls session termination.
type KillSessionOptions struct {
	Session string
	Socket  string
}

// KillSession terminates the specified tmux session.
func (c *Client) KillSession(ctx context.Context, opts KillSessionOptions) error {
	session := strings.TrimSpace(opts.Session)
	if session == "" {
		return errors.New("session name is required")
	}
	socket := c.resolveSocket(opts.Socket)

	args := c.baseArgs(socket, "kill-session", "-t", session)
	_, stderr, exitCode, err := c.runner.Run(ctx, args, nil)
	if err != nil {
		return fmt.Errorf("tmux kill-session: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("tmux kill-session exit %d: %s", exitCode, strings.TrimSpace(stderr))
	}
	return nil
}

// HasSession reports whether the given tmux session exists.
func (c *Client) HasSession(ctx context.Context, session, socket string) (bool, error) {
	session = strings.TrimSpace(session)
	if session == "" {
		return false, errors.New("session name is required")
	}
	socket = c.resolveSocket(socket)

	args := c.baseArgs(socket, "has-session", "-t", session)
	_, stderr, exitCode, err := c.runner.Run(ctx, args, nil)
	if err != nil {
		return false, fmt.Errorf("tmux has-session: %w", err)
	}
	if exitCode == 0 {
		return true, nil
	}
	if exitCode == 1 {
		return false, nil
	}
	return false, fmt.Errorf("tmux has-session exit %d: %s", exitCode, strings.TrimSpace(stderr))
}

// SendKeysOptions controls key dispatch.
type SendKeysOptions struct {
	Target  string
	Socket  string
	Keys    []string
	Literal bool
}

// SendKeys sends a sequence of keys to the target pane.
func (c *Client) SendKeys(ctx context.Context, opts SendKeysOptions) error {
	if len(opts.Keys) == 0 {
		return errors.New("at least one key is required")
	}
	socket := c.resolveSocket(opts.Socket)

	args := c.baseArgs(socket, "send-keys")
	if opts.Target != "" {
		args = append(args, "-t", opts.Target)
	}
	if opts.Literal {
		args = append(args, "-l")
	}
	args = append(args, opts.Keys...)

	_, stderr, exitCode, err := c.runner.Run(ctx, args, nil)
	if err != nil {
		return fmt.Errorf("tmux send-keys: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("tmux send-keys exit %d: %s", exitCode, strings.TrimSpace(stderr))
	}
	return nil
}

func (c *Client) baseArgs(socket string, rest ...string) []string {
	args := make([]string, 0, len(rest)+2)
	if socket != "" {
		args = append(args, "-S", socket)
	}
	args = append(args, rest...)
	return args
}

func (c *Client) resolveSocket(socket string) string {
	if socket != "" {
		return socket
	}
	return c.defaultSocket
}

// AttachOptions controls attach behaviour.
type AttachOptions struct {
	Session string
	Socket  string
}

// Attach connects the user to the specified tmux session.
func (c *Client) Attach(ctx context.Context, opts AttachOptions) error {
	session := strings.TrimSpace(opts.Session)
	if session == "" {
		return errors.New("session name is required")
	}
	socket := c.resolveSocket(opts.Socket)

	args := c.baseArgs(socket, "attach-session", "-t", session)
	_, stderr, exitCode, err := c.runner.Run(ctx, args, nil)
	if err != nil {
		return fmt.Errorf("tmux attach-session: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("tmux attach-session exit %d: %s", exitCode, strings.TrimSpace(stderr))
	}
	return nil
}

// ExecRunner executes tmux commands using os/exec.
type ExecRunner struct {
	Bin string
}

// Run executes the tmux command and captures stdout/stderr.
func (r *ExecRunner) Run(ctx context.Context, args []string, env map[string]string) (string, string, int, error) {
	bin := r.Bin
	if bin == "" {
		bin = "tmux"
	}
	cmd := exec.CommandContext(ctx, bin, args...)

	if env != nil {
		cmd.Env = mergeEnv(os.Environ(), env)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return "", "", -1, err
	}

	err := cmd.Wait()
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
			return stdout.String(), stderr.String(), exitCode, nil
		}
		return "", "", -1, err
	}

	return stdout.String(), stderr.String(), exitCode, nil
}

func mergeEnv(base []string, overrides map[string]string) []string {
	if len(overrides) == 0 {
		return base
	}
	envMap := make(map[string]string, len(base)+len(overrides))
	for _, kv := range base {
		if idx := strings.IndexByte(kv, '='); idx > 0 {
			envMap[kv[:idx]] = kv[idx+1:]
		}
	}
	for k, v := range overrides {
		envMap[k] = v
	}
	keys := make([]string, 0, len(envMap))
	for k := range envMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := make([]string, 0, len(envMap))
	for _, k := range keys {
		result = append(result, fmt.Sprintf("%s=%s", k, envMap[k]))
	}
	return result
}
