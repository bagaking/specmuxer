# Feature Specification: SpecMuxer tmux AI Orchestrator

**Feature Branch**: `001-specify-scripts-bash`  
**Created**: 2025-10-18  
**Status**: Draft  
**Input**: User description: "这个项目的目标是基于 tmux 实现一个管理多个 ai 变成工具并发执行的过程, 其功能包含了"

## User Scenarios & Testing *(mandatory)*

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.
  
  Assign priorities (P1, P2, P3, etc.) to each story, where P1 is the most critical.
  Think of each story as a standalone slice of functionality that can be:
  - Developed independently
  - Tested independently
  - Deployed independently
  - Demonstrated to users independently
-->

### User Story 1 - Launch and Monitor Tool Sessions (Priority: P1)

As a developer, I can start AI coding tools through `specmuxer run`, pass through custom arguments and environment variables, and immediately observe session activity from `status`, `top`, or the tmux-enhanced UI so I know which tools are busy.

**Why this priority**: Running tooling on demand with instant visibility is the core value proposition.

**Independent Test**: Launch a new session in a fresh workspace, verify command output, check status/top summaries, and inspect project-scoped logs without invoking recovery or housekeeping.

**Acceptance Scenarios**:

1. **Given** a workspace without active sessions, **When** the user runs `specmuxer run codex -- --plan`, **Then** a tmux session is created with status and logs recorded under `.specmuxer/`.
2. **Given** multiple sessions in flight, **When** the user calls `status --json`, **Then** the response lists each session with alive/idle determination based on the last output timestamp and serializes the project/session hierarchy.

---

### User Story 2 - Resume After Interruptions (Priority: P2)

As a developer returning after a crash or reboot, I can use `resume` to restore eligible sessions and their tool-specific context so that automated assistants continue where they left off without manual tmux surgery.

**Why this priority**: Resilience is a stated minimal goal and enables uninterrupted delivery.

**Independent Test**: Terminate the machine or tmux server while sessions run, restart the workspace, execute `resume --all`, and confirm only non user-killed sessions restart with tool-specific resume semantics.

**Acceptance Scenarios**:

1. **Given** a session interrupted by host reboot with `user_killed=false`, **When** the user executes `resume --all`, **Then** the system restarts the session and triggers any post-resume commands defined in its YAML.
2. **Given** a session marked `user_killed=true`, **When** the user runs `resume --all`, **Then** the session remains stopped and is reported as intentionally terminated.

---

### User Story 3 - Maintain and Inspect System Health (Priority: P3)

As a maintainer, I can use `logs`, `gc`, and `doctor` to diagnose issues, prune orphaned resources, and confirm the toolchain meets requirements so ongoing operations remain predictable and auditable.

**Why this priority**: Operational hygiene protects long-lived automation and compliance standards.

**Independent Test**: Run multiple sessions, generate logs, invoke `gc --dry-run`, follow logs, and execute `doctor`; confirm actionable guidance, adoption prompts, and permissions checks are produced without touching resume/run.

**Acceptance Scenarios**:

1. **Given** lingering tmux sessions not managed by SpecMuxer, **When** the user runs `gc`, **Then** the command offers adopt/ignore options and removes stale registry entries as configured.
2. **Given** a misconfigured environment (e.g., unsupported tmux version), **When** the user runs `doctor`, **Then** the command highlights the failing check, names the impacted project/session, and suggests the concrete remediation.

### Edge Cases

- Session YAML is edited with invalid resume hooks → system must reject the change with diagnostics without corrupting the stored record.
- Log rotation reaches the 200 MB limit mid-session → rotation should preserve continuity without losing buffered data or blocking the running tool.
- A session adapter lacks resume support → `resume` reports the limitation and skips only that tool while continuing others.
- User manually kills a tmux pane but not the session → health detection treats the pane as inactive yet keeps the session eligible for resume unless `user_killed=true`.

## Requirements *(mandatory)*

<!--
  ACTION REQUIRED: The content in this section represents placeholders.
  Fill them out with the right functional requirements.
-->

### Functional Requirements

- **FR-001**: Provide CLI commands `run`, `resume`, `attach`, `status`, `top`, `stop`, `logs`, `gc`, and `doctor` with arguments exactly as specified, including passthrough of `--` and support for JSON and wide formatting where called out.
- **FR-002**: Persist all configuration, state, statistics, session metadata, and logs within a per-project `.specmuxer/` directory containing `conf.yml`, `stats.yml`, `sessions/*.yml`, and `logs/*.log`, honoring default permissions (0700 for directory, 0600 for logs) and configurable log rotation (default 200 MB daily).
- **FR-003**: Generate stable project identifiers by hashing absolute paths and name tmux sessions using the pattern `specmuxer_<hash>_<session-id>` (or equivalent sanitized form) to avoid collisions and ensure compatibility with `tmux` naming rules.
- **FR-004**: Record session activity with timestamped log lines stored per session; `logs <session> --follow` must stream in real time and respect redaction rules defined via regular expressions.
- **FR-005**: Deliver resume semantics where non user-killed sessions interrupted unexpectedly are eligible for automatic or manual recovery, while `user_killed=true` sessions are excluded until explicitly re-run.
- **FR-006**: Allow users to toggle between manual resume (default) and automatic startup integration, and honor session-level YAML directives for post-resume commands such as `"请继续"` or `"/goahead"`.
- **FR-007**: Detect actual tool liveness for `status` and `top` views using recent output timestamps (default idle threshold 120 s, configurable to 90/180 s) and present hierarchical project→session→pane data with optional JSON output, including tmux availability indicators for each session.
- **FR-008**: Provide an adapter abstraction describing `start`, `resume`, `health`, `stop`, and `extract_state` behaviors per tool, allow project-scoped overrides, and ship defaults for popular AI assistants (codex, claude, etc.).
- **FR-009**: Accept `--env KEY=VALUE` flags to propagate environment variables into launched or resumed tool processes and maintain tmux UI enhancements (status bar with last activity, sidebar shortcuts) compatible with mouse/trackpad interactions.
- **FR-010**: Support Linux and macOS with tmux ≥ 3.x without requiring root access, and document best-effort guidance for Windows via WSL while explicitly stating lack of native Windows support.
- **FR-011**: Enforce performance SLOs: `status` on ≤50 sessions completes with P95 under 300 ms; a `resume --all` over ≤20 sessions finishes within 10 s barring external cold starts.
- **FR-012**: Implement `gc` to prune orphaned registry entries, stale tmux resources, and offer adopt/ignore choices for unmanaged tmux sessions before removal.
- **FR-013**: Ensure `doctor` inspects tmux version, PATH entries, permissions, and configuration validity, reporting next-step guidance referencing the affected project/session/adapter.
- **FR-014**: Publish user-facing documentation (README “5-minute start”, docs covering concepts, adapters, recovery, FAQ), Semantic Versioning with changelog, MIT license, and issue/PR templates as part of the deliverable.
- **FR-015**: Provide configuration for log redaction patterns, retain structured statistics in `stats.yml`, and expose metrics through CLI outputs without requiring external databases.
- **FR-016**: Detect non-interactive shells and gracefully skip tmux attach attempts, surfacing socket/session details and guidance rather than propagating tmux errors.
- **FR-017**: Before launching additional sessions while others remain active, prompt for confirmation (defaulting to abort when input is non-interactive) with an explicit `--force` override for automation.

### Key Entities *(include if feature involves data)*

- **Project Workspace**: Represents a root directory under orchestration, storing `.specmuxer/` assets, project_id hash, and configuration defaults.
- **Session Record**: YAML entry capturing tool name, adapter, creation time, resume eligibility, env overrides, health data, and user annotations.
- **Adapter Definition**: Declarative profile describing lifecycle commands and health checks for a specific AI tool, supporting inheritance/override.
- **Log File**: Time-stamped textual record associated with a session, subject to rotation, redaction, and follow streaming.
- **Statistics Snapshot**: Aggregated metrics stored in `stats.yml`, including counts of active/idle sessions, last activity timestamps, and GC history.

## Assumptions

- Manual resume is the default startup mode; users opting into automatic startup will configure platform-specific services outside this feature.
- AI tools referenced (e.g., codex, claude) expose command-line interfaces capable of resume semantics when provided stored identifiers.
- Users operate within local workstations or servers where tmux 3.x is available and accessible without elevated privileges.

## Success Criteria *(mandatory)*

<!--
  ACTION REQUIRED: Define measurable success criteria.
  These must be technology-agnostic and measurable.
-->

### Measurable Outcomes

- **SC-001**: A developer launching a new session can confirm creation, activity status, and log availability within 5 seconds of running `specmuxer run`.
- **SC-002**: `status --json` returns hierarchical activity data for 50 or fewer sessions within 300 ms (P95) during acceptance testing.
- **SC-003**: A simulated reboot scenario demonstrates that ≥95% of eligible sessions resume successfully with preserved context using a scripted `resume --all`.
- **SC-004**: `gc --dry-run` correctly identifies 100% of unmanaged tmux sessions in test fixtures and reports adopt/ignore actions without deleting managed sessions.
- **SC-005**: Usability testing with three target users reports that doctor guidance makes next steps clear without needing implementation knowledge (4/5 satisfaction rating or better).
