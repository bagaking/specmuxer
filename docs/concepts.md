# Architecture Overview

SpecMuxer follows a layered architecture to keep tmux orchestration, domain logic, and CLI presentation isolated:

- **Interface (cmd/specmuxer)** – Cobra-based commands (`run`, `status`, `top`, `resume`, `logs`, `gc`, `doctor`, `attach`) parse flags, enforce UX rules, and delegate to orchestrator services.
- **Application (pkg/orchestrator)** – `run.Service` launches sessions, materialises YAML records, and updates telemetry; `resume.Service` restores eligible sessions while honouring adapter overrides; GC/Doctor helpers live under `internal/applier` for cross-cutting maintenance.
- **Domain (pkg/domain)** – Session records, YAML helpers, and adapter definitions describe persisted state, resume hooks, and lifecycle commands with cloning helpers for overrides.
- **Infrastructure (pkg/runtime / pkg/telemetry)** – Thin wrappers around tmux (`pkg/runtime/tmux`), YAML storage (`pkg/runtime/storage`), log rotation/readers, and status collectors provide reusable building blocks. UI packages render status bars and GC prompts without coupling to Cobra.

Key data directories are resolved from `.specmuxer/` via `pkg/orchestrator/config`, which enforces permissions, project IDs, and resume mode (manual/automatic). Telemetry snapshots are collected via `pkg/telemetry/stats` and persisted through `storage.Stats` with resume success metrics.

```
cmd/               # CLI entrypoints
pkg/
  orchestrator/   # run + resume services, log writers
  runtime/        # tmux and YAML persistence
  telemetry/      # stats collectors, log manager/reader
  adapters/       # adapter registry + overrides
  ui/             # status bar + GC prompt helpers
internal/
  integration/    # end-to-end tests
  applier/        # GC and doctor workflows
```
