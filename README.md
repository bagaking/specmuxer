# SpecMuxer

SpecMuxer orchestrates tmux-based AI tool sessions with durable YAML metadata, log redaction, and self-healing workflows.

## CLI Commands

| Command | Description |
|---------|-------------|
| `specmuxer run <adapter> [--env KEY=VALUE] [-- name] [-- adapter args] [--force]` | Launch a new adapter session inside tmux (auto-detaches when no interactive TTY is present; prompts before creating another when sessions already running) |
| `specmuxer status [--json] [--wide]` | Show session hierarchy and metrics |
| `specmuxer top` | Stream live updates similar to `top` |
| `specmuxer resume [--all] [--dry-run] [session id...]` | Resume eligible sessions |
| `specmuxer logs --session id [--tail N] [--follow]` | Tail redacted session logs |
| `specmuxer gc [--adopt]` | Inspect and prune orphaned sessions |
| `specmuxer doctor` | Run environment diagnostics |
| `specmuxer attach <session>` | Attach to an existing tmux session (requires interactive TTY) |

## Quickstart

1. Install dependencies (`go install ./cmd/specmuxer`).
2. Verify setup:
   ```bash
   specmuxer doctor
   ```
3. Launch a session:
   ```bash
   specmuxer run codex --name primary -- --plan
   specmuxer status --wide
   ```
   *When run from CI or any non-interactive shell, SpecMuxer skips automatic attach and prints the tmux session/socket so you can connect later from a real terminal. If other sessions are already running, add `--force` to bypass the confirmation prompt automatically.*
4. Resume after reboot:
   ```bash
   specmuxer resume --all
   ```
5. Maintain hygiene:
   ```bash
   specmuxer logs --session <id> --tail 20
   specmuxer gc
   ```

See `docs/` for architecture, adapter overrides, recovery walkthroughs, and troubleshooting tips.

## Maintainer Verification

Use the make targets as the local verification entrypoints:

```bash
make test
make lint
make coverage
```

`make lint` requires `golangci-lint` on `PATH`. `make coverage` runs the race-enabled test suite with an atomic coverage profile and reports the project threshold result; the current tree is below the configured threshold, so use it as a coverage gap report until coverage is raised or the threshold policy changes.
