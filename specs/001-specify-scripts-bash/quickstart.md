# Quickstart — SpecMuxer tmux AI Orchestrator

Goal: Launch SpecMuxer, start an AI tool session, observe activity, and exercise recovery within five minutes.

## 1. Install Prerequisites (1 minute)
- Ensure Go ≥ 1.23 with toolchain 1.24.2: `go env GOTOOLDIR` should resolve.
- Install tmux ≥ 3.x: `tmux -V`.
- Clone the repository and run `make codex_init` (bootstraps .specify assets).

## 2. Build the CLI (1 minute)
```bash
go mod tidy
go build -o bin/specmuxer ./cmd/specmuxer
```

## 3. Initialize a Workspace (30 seconds)
```bash
mkdir -p ~/workspace/demo && cd ~/workspace/demo
specmuxer doctor   # creates .specmuxer/, validates tmux + permissions
```

Expected: `.specmuxer/` with `conf.yml`, `stats.yml`, and permission 0700.

## 4. Run an AI Tool Session (1 minute)
```bash
specmuxer run codex -- --plan
```
- SpecMuxer launches a tmux session named `specmuxer:p-<hash>-codex-...`.
- Status bar/side bar appear with last activity and shortcuts.
- Logs stream to `.specmuxer/logs/<session>.log`.

Monitor:
```bash
specmuxer status --json | jq '.projects[0].sessions[0]'
specmuxer logs <session> --follow
specmuxer top
```

## 5. Simulate Interruption & Resume (1 minute)
```bash
tmux kill-session -t specmuxer:p-*
specmuxer resume --all
```
- Resume summary shows which sessions restarted.
- Post-resume commands from `sessions/<session>.yml` run automatically.

## 6. Housekeeping & Diagnostics (1 minute)
```bash
specmuxer gc
specmuxer doctor
```
- `gc --adopt` cleans orphaned records when you're confident tmux sessions are gone.
- `doctor` validates tmux availability, workspace permissions, and configuration.

## 7. Documentation & Next Steps
- Explore `docs/concepts.md` for adapter architecture.
- Update `.specmuxer/conf.yml` to configure automatic resume or log redaction.
- Review `CHANGELOG.md` for release history and upcoming milestones.
