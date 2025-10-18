# SpecMuxer

SpecMuxer orchestrates tmux-based AI tool sessions with durable YAML metadata, log redaction, and self-healing workflows.

## CLI Commands

| Command | Description |
|---------|-------------|
| `specmuxer run <adapter> [--env KEY=VALUE] [-- name] [-- adapter args]` | Launch a new adapter session inside tmux |
| `specmuxer status [--json] [--wide]` | Show session hierarchy and metrics |
| `specmuxer top` | Stream live updates similar to `top` |
| `specmuxer resume [--all|--session id] [--dry-run]` | Resume eligible sessions |
| `specmuxer logs --session id [--tail N] [--follow]` | Tail redacted session logs |
| `specmuxer gc [--adopt]` | Inspect and prune orphaned sessions |
| `specmuxer doctor` | Run environment diagnostics |
| `specmuxer attach <session>` | Attach to an existing tmux session |

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
