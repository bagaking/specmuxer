# Data Model — SpecMuxer tmux AI Orchestrator

## Overview
SpecMuxer persists all operational state within a project-local `.specmuxer/` directory. Core entities capture projects, sessions, adapters, and telemetry while remaining decoupled from specific AI tools.

## Entities

### ProjectWorkspace
- **project_id** (`string`, hash of absolute path) — stable identifier across restarts.
- **root_path** (`string`) — absolute workspace path.
- **config_path** (`string`, default `.specmuxer/conf.yml`) — configuration file location.
- **stats_path** (`string`, default `.specmuxer/stats.yml`) — aggregated metrics store.
- **log_dir** (`string`, default `.specmuxer/logs/`) — parent directory for rotated logs.
- **session_dir** (`string`, default `.specmuxer/sessions/`) — location of session YAML records.
- **resume_mode** (`enum`, `manual` | `automatic`) — default resume strategy; manual by default.
- **redaction_rules** (`[]regex`) — ordered list of redaction expressions applied to logs.

### SessionRecord
- **id** (`string`, UUID) — unique identifier for session.
- **project_id** (`string`) — foreign key to `ProjectWorkspace`.
- **tool** (`string`) — adapter name (e.g., `codex`, `claude`).
- **human_name** (`string`) — display label / CLI `--name`.
- **created_at** (`timestamp`) — session creation time.
- **last_output_at** (`timestamp`) — most recent stdout/stderr line timestamp.
- **status** (`enum`, `running` | `idle` | `stopped` | `failed`) — runtime status.
- **user_killed** (`bool`) — opt-out flag for automatic resume.
- **tmux** (`object`)  
  - **session** (`string`) — tmux session identifier (`specmuxer:p-<hash>-…`).  
  - **window** (`string`) — tmux window/pane reference.  
  - **socket** (`string`) — tmux socket path (per-project).  
- **adapter_state** (`map[string]any`) — adapter-specific serialized payload for resume/extract_state.
- **env** (`map[string]string`) — environment overrides applied during run/resume.
- **log_files** (`[]LogPointer`) — active log segments.
- **resume_hooks** (`[]ResumeCommand`) — ordered commands/messages to emit post-resume.
- **error** (`string`) — last failure reason (empty if healthy).

### ResumeCommand
- **type** (`enum`, `prompt` | `shell`) — whether to send a textual prompt or execute a shell command.
- **value** (`string`) — literal text or shell snippet.
- **delay_seconds** (`int`, optional) — wait before dispatching command to tool.

### LogPointer
- **path** (`string`) — absolute path to log file.
- **size_bytes** (`int`) — current file size.
- **segment_date** (`date`) — rotation date anchor.
- **checksum** (`string`, optional) — checksum for integrity auditing.

### AdapterDefinition
- **name** (`string`) — unique adapter key.
- **start** (`LifecycleCommand`) — command template to launch tool.
- **resume** (`LifecycleCommand`) — optional resume template; absence implies resume unsupported.
- **health** (`LifecycleCommand`) — probe to assess liveness (command or script).
- **stop** (`LifecycleCommand`) — graceful shutdown instruction.
- **extract_state** (`LifecycleCommand`) — optional archival step.
- **supports_attach** (`bool`) — indicates attach semantics availability.
- **overrides** (`map[string]LifecycleCommand`) — project-scoped overrides keyed by command name.

### LifecycleCommand
- **exec** (`[]string`) — command + arguments template.
- **working_dir** (`string`, optional) — default working directory.
- **env** (`map[string]string`) — additional environment variables.
- **timeout_seconds** (`int`, optional) — max runtime.
- **on_failure** (`enum`, `retry` | `abort` | `fallback`) — failure policy for orchestrator.

### StatisticsSnapshot
- **collected_at** (`timestamp`) — when snapshot saved.
- **active_sessions** (`int`) — count of running sessions.
- **idle_sessions** (`int`) — count of idle sessions (above idle threshold).
- **last_gc** (`timestamp`) — most recent GC execution.
- **resume_success_rate** (`float`) — success ratio for resume attempts.
- **top_latency_p95_ms** (`int`) — rolling latency metric for status/top.
- **log_volume_bytes** (`int`) — total log storage footprint.

### ObservabilityRecord
- **session_id** (`string`) — associated session.
- **activity_events** (`[]ActivityEvent`) — append-only timeline of key events (run, resume, stop, error).
- **performance_metrics** (`map[string]float64`) — per-session stats such as resume duration, command latency.

### ActivityEvent
- **timestamp** (`timestamp`)
- **event_type** (`enum`, `run`, `resume`, `stop`, `gc`, `doctor`, `error`, `log_rotated`)
- **details** (`map[string]string`) — contextual metadata (e.g., command, adapter result).

## Relationships

- `ProjectWorkspace` 1 — * contains many `SessionRecord`, `LogPointer`, and `StatisticsSnapshot`.
- `SessionRecord` 1 — * references many `ResumeCommand` and `LogPointer`.
- `AdapterDefinition` is registered globally, with optional project-local overrides stored alongside project configuration.
- `StatisticsSnapshot` summarizes metrics derived from `SessionRecord` and `ObservabilityRecord` data.

## State Transitions

```
      ┌──────────────┐
      │   running    │
      └─────┬────────┘
            │ inactivity ≥ idle threshold
            ▼
      ┌──────────────┐
      │     idle     │
      └─────┬────────┘
            │ resume invoked / new output
            ▼
      ┌──────────────┐
      │   running    │
      └─────┬────────┘
            │ graceful stop or user_killed=true
            ▼
      ┌──────────────┐
      │   stopped    │
      └─────┬────────┘
            │ unexpected failure (user_killed=false)
            ▼
      ┌──────────────┐
      │    failed    │
      └─────┬────────┘
            │ resume successful
            ▼
      ┌──────────────┐
      │   running    │
      └──────────────┘
```

## Validation Rules

- `.specmuxer/` directory must be created with `0700`; logs must be `0600`.
- YAML files must conform to schema; invalid edits trigger descriptive validation errors (doctor command).
- Idle detection uses configurable threshold (default 120 s) stored in configuration.
- Redaction rules must compile; invalid regex entries prevent log streaming until corrected.
- Resume is only attempted when `user_killed` is false and adapter supports resume.
- Log rotation ensures individual segments ≤ configured max size (default 200 MB); GC prunes segments beyond retention window.
