package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/bagaking/specmuxer/pkg/adapters"
	"github.com/bagaking/specmuxer/pkg/domain/session"
	"github.com/bagaking/specmuxer/pkg/orchestrator/config"
	"github.com/bagaking/specmuxer/pkg/orchestrator/run"
	"github.com/bagaking/specmuxer/pkg/runtime/tmux"
	"golang.org/x/term"
)

func findSession(records []session.SessionRecord, id string) (session.SessionRecord, bool) {
	for _, rec := range records {
		if rec.ID == id {
			return rec, true
		}
	}
	return session.SessionRecord{}, false
}

type LivenessInfo struct {
	Alive   bool
	Session string
}

func computeLiveness(ctx context.Context, client *tmux.Client, records []session.SessionRecord, defaultSocket string) map[string]LivenessInfo {
	result := make(map[string]LivenessInfo, len(records))
	if ctx == nil {
		ctx = context.Background()
	}
	probes := map[struct {
		socket  string
		session string
	}]bool{}
	for _, rec := range records {
		socket := rec.Tmux.Socket
		if socket == "" {
			socket = defaultSocket
		}
		info := LivenessInfo{Alive: false, Session: rec.Tmux.Session}
		if info.Session == "" {
			info.Session = run.BuildSessionName(rec.ProjectID, rec.ID)
		}
		seen := map[string]struct{}{}
		candidates := []string{}
		addCandidate := func(name string) {
			if name == "" {
				return
			}
			if _, ok := seen[name]; ok {
				return
			}
			seen[name] = struct{}{}
			candidates = append(candidates, name)
		}
		addCandidate(rec.Tmux.Session)
		if rec.Tmux.Session != "" {
			addCandidate(strings.ReplaceAll(rec.Tmux.Session, ":", "_"))
		}
		addCandidate(run.BuildSessionName(rec.ProjectID, rec.ID))

		for _, candidate := range candidates {
			key := struct {
				socket  string
				session string
			}{socket: socket, session: candidate}
			ok, seen := probes[key]
			if !seen {
				var err error
				ok, err = client.HasSession(ctx, candidate, socket)
				ok = err == nil && ok
				probes[key] = ok
			}
			if ok {
				info.Alive = true
				info.Session = candidate
				break
			}
		}
		result[rec.ID] = info
	}
	return result
}

func livenessAsBool(info map[string]LivenessInfo) map[string]bool {
	res := make(map[string]bool, len(info))
	for id, inf := range info {
		res[id] = inf.Alive
	}
	return res
}

func hasInteractiveTerminal() bool {
	return allTerminal(os.Stdin, os.Stdout, os.Stderr)
}

func allTerminal(files ...*os.File) bool {
	if len(files) == 0 {
		return false
	}
	for _, f := range files {
		if !isTerminalFile(f) {
			return false
		}
	}
	return true
}

func isTerminalFile(f *os.File) bool {
	if f == nil {
		return false
	}
	fd := int(f.Fd())
	return term.IsTerminal(fd)
}

func resolveDefinition(cfg *config.Config, registry *adapters.Registry, tool string) (adapters.Definition, error) {
	return registry.Resolve(tool, adapterOverrides(cfg, tool))
}

func adapterOverrides(cfg *config.Config, tool string) map[string]adapters.LifecycleCommand {
	if cfg == nil || cfg.AdapterOverrides == nil {
		return nil
	}
	ov, ok := cfg.AdapterOverrides[tool]
	if !ok {
		ov, ok = cfg.AdapterOverrides[strings.ToLower(tool)]
	}
	if !ok {
		return nil
	}
	result := make(map[string]adapters.LifecycleCommand)
	if len(ov.Start) > 0 {
		result["start"] = adapters.LifecycleCommand{Exec: append([]string(nil), ov.Start...), WorkingDir: ov.WorkingDir, Env: cloneEnv(ov.Env), Timeout: time.Duration(ov.TimeoutSec) * time.Second}
	}
	if len(ov.Resume) > 0 {
		result["resume"] = adapters.LifecycleCommand{Exec: append([]string(nil), ov.Resume...), WorkingDir: ov.WorkingDir, Env: cloneEnv(ov.Env), Timeout: time.Duration(ov.TimeoutSec) * time.Second}
	}
	if len(ov.Health) > 0 {
		result["health"] = adapters.LifecycleCommand{Exec: append([]string(nil), ov.Health...), WorkingDir: ov.WorkingDir, Env: cloneEnv(ov.Env), Timeout: time.Duration(ov.TimeoutSec) * time.Second}
	}
	if len(ov.Stop) > 0 {
		result["stop"] = adapters.LifecycleCommand{Exec: append([]string(nil), ov.Stop...), WorkingDir: ov.WorkingDir, Env: cloneEnv(ov.Env), Timeout: time.Duration(ov.TimeoutSec) * time.Second}
	}
	if len(ov.ExtractState) > 0 {
		result["extract_state"] = adapters.LifecycleCommand{Exec: append([]string(nil), ov.ExtractState...), WorkingDir: ov.WorkingDir, Env: cloneEnv(ov.Env), Timeout: time.Duration(ov.TimeoutSec) * time.Second}
	}
	return result
}

func cloneEnv(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func filterActiveSessions(records []session.SessionRecord) []session.SessionRecord {
	active := make([]session.SessionRecord, 0, len(records))
	for _, rec := range records {
		if rec.Status == session.StatusRunning || rec.Status == session.StatusIdle {
			active = append(active, rec)
		}
	}
	return active
}

func promptYesNo(in io.Reader, out io.Writer, prompt string) (bool, error) {
	if out == nil {
		out = os.Stdout
	}
	if _, err := fmt.Fprint(out, prompt); err != nil {
		return false, err
	}
	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	response := strings.ToLower(strings.TrimSpace(line))
	if response == "y" || response == "yes" {
		return true, nil
	}
	return false, nil
}
