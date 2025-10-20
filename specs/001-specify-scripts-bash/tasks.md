---
description: "Task list for SpecMuxer tmux AI Orchestrator implementation"
---

# Tasks: SpecMuxer tmux AI Orchestrator

**Input**: Design documents from `/specs/001-specify-scripts-bash/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are MANDATORY. Write or extend tests before or alongside new code to maintain ≥95% coverage on touched packages. Integration tests must exercise real tmux sockets where feasible.

**Organization**: Tasks are grouped by user story to keep each slice independently deliverable.

## Format: `[ID] [P?] [Story] Description`
- `[P]`: Task can run in parallel (different files, no dependency). Omit when sequential.
- `[US#]`: Tag tasks that belong to a user story phase (US1, US2, US3, ...). Not used for Setup/Foundational/Polish phases.
- Always include explicit file paths.

## Phase 1: Setup (Shared Infrastructure)

- [ ] T001 Verify Go module toolchain pinning and tidy modules in `go.mod` / `go.sum`
- [ ] T002 Configure Make targets for lint/test (`go test ./... -race -count=1`, `golangci-lint`) in `Makefile`
- [ ] T003 Scaffold `.specmuxer/` default layout and permissions enforcement helpers in `pkg/runtime/storage`

---

## Phase 2: Foundational (Blocking Prerequisites)

- [ ] T004 Implement configuration loader with validation in `pkg/orchestrator/config/config.go`
- [ ] T005 Implement session domain structs & YAML serialization in `pkg/domain/session/session.go`
- [ ] T006 Implement YAML persistence driver (sessions/stats) in `pkg/runtime/storage/yaml_store.go`
- [ ] T007 Implement tmux client wrappers (ensure/new-window/attach with stdio binding) in `pkg/runtime/tmux/client.go`
- [ ] T008 Implement log manager with rotation/redaction in `pkg/telemetry/logs/manager.go`
- [ ] T009 Register adapter definitions & overrides in `pkg/adapters/registry.go`
- [ ] T010 Implement telemetry collector & status snapshot in `pkg/telemetry/stats/collector.go`
- [ ] T011 [P] Add unit tests for configuration, storage, tmux client, adapters, and telemetry packages (`pkg/**/*_test.go`)
- [ ] T012 [P] Set up integration fixture launching isolated tmux socket in `internal/integration/tmux_fixture_test.go`

---

## Phase 3: User Story 1 — Launch and Monitor Tool Sessions (Priority P1)

**Goal**: `run`, `status`, `top`, `attach`, tmux UI, and logging deliver real-time visibility.
**Independent Test**: Launch `specmuxer run codex -- --plan` in a clean workspace, inspect status/top/logs without resume/gc.

### Tests (write before implementation)
- [ ] T013 [US1] Add run/status lifecycle integration test in `internal/integration/run_status_test.go`
- [ ] T014 [P] [US1] Add unit tests for status JSON formatter & idle detection in `pkg/telemetry/stats/formatter_test.go`
- [ ] T015 [P] [US1] Add unit tests for tmux attach/session name helpers in `cmd/specmuxer/helpers_test.go`

### Implementation
- [ ] T016 [US1] Implement run orchestrator service launching tmux sessions in `pkg/orchestrator/run/service.go`
- [ ] T017 [US1] Implement CLI `run` command with env passthrough & active-session prompts in `cmd/specmuxer/run_cmd.go`
- [ ] T018 [US1] Implement session logging writer in `pkg/orchestrator/run/log_writer.go`
- [ ] T019 [P] [US1] Implement status command + JSON/wide output in `cmd/specmuxer/status_cmd.go`
- [ ] T020 [P] [US1] Implement top streaming UI in `pkg/telemetry/stats/top.go`
- [ ] T021 [US1] Implement attach command with liveness table & non-TTY guidance in `cmd/specmuxer/attach_cmd.go`
- [ ] T022 [P] [US1] Implement status/sidebar renderer with cmdux in `pkg/ui/statusbar/statusbar.go`
- [ ] T023 [P] [US1] Wire CLI root/runtime dependencies in `cmd/specmuxer/runtime.go`
- [ ] T024 [US1] Document run/status behavior (README, docs/recovery.md) reflecting terminal detection

---

## Phase 4: User Story 2 — Resume After Interruptions (Priority P2)

**Goal**: `resume` restores eligible sessions, respects user_killed flag, and replays hooks.
**Independent Test**: Kill tmux socket, run `resume --all`, inspect resume summary/log hooks.

### Tests
- [ ] T025 [US2] Add resume integration test covering interrupted sessions in `internal/integration/resume_test.go`
- [ ] T026 [P] [US2] Add unit tests for eligibility filtering & tmux liveness in `pkg/orchestrator/resume/filter_test.go`

### Implementation
- [ ] T027 [US2] Implement resume service (eligibility, adapter hooks) in `pkg/orchestrator/resume/service.go`
- [ ] T028 [US2] Implement CLI `resume` command (filters, summaries) in `cmd/specmuxer/resume_cmd.go`
- [ ] T029 [P] [US2] Extend storage to persist resume hooks & adapter state in `pkg/runtime/storage/yaml_store.go`
- [ ] T030 [P] [US2] Update telemetry collector to track resume metrics in `pkg/telemetry/stats/collector.go`
- [ ] T031 [US2] Update docs/recovery.md with resume workflow & troubleshooting guidance

---

## Phase 5: User Story 3 — Maintain and Inspect System Health (Priority P3)

**Goal**: `logs`, `gc`, `doctor`, `stop` provide operational insight and hygiene.
**Independent Test**: Produce logs, run `gc --dry-run`, follow logs, run `doctor`, confirm actionable guidance.

### Tests
- [ ] T032 [US3] Add integration tests for gc adoption & doctor failure diagnostics in `internal/integration/gc_doctor_test.go`
- [ ] T033 [P] [US3] Add unit tests for log tailing + redaction in `pkg/telemetry/logs/manager_test.go`

### Implementation
- [ ] T034 [US3] Implement `logs` command with follow & redaction in `cmd/specmuxer/logs_cmd.go`
- [ ] T035 [US3] Implement `gc` command with adopt/ignore and telemetry integration in `cmd/specmuxer/gc_cmd.go`
- [ ] T036 [US3] Implement `doctor` command checking tmux version, permissions, config in `cmd/specmuxer/doctor_cmd.go`
- [ ] T037 [P] [US3] Implement `stop` command to gracefully terminate sessions in `cmd/specmuxer/stop_cmd.go`
- [ ] T038 [P] [US3] Update statistics snapshot to track log volume & last GC in `pkg/telemetry/stats/collector.go`
- [ ] T039 [US3] Document operational runbooks (docs/recovery.md, docs/faq.md)

---

## Phase 6: Polish & Cross-Cutting Concerns

- [ ] T040 Add CLI completions & help text polish in `cmd/specmuxer/main.go`
- [ ] T041 Add performance verification script for status/top in `scripts/perf/status_benchmark.go`
- [ ] T042 Refresh README quickstart & changelog entries in `README.md` / `CHANGELOG.md`
- [ ] T043 Conduct security & dependency review doc in `docs/security.md`
- [ ] T044 Final `go test ./... -race -count=1` and lint verification via `Makefile`

---

## Dependencies & Execution Order

1. Setup (Phase 1) → Foundational (Phase 2) → User Story phases (US1 → US2 → US3) → Polish.
2. Within each user story, follow ordering: Tests → Core services → CLI commands → Docs.
3. Parallel `[P]` tasks may run once their prerequisites complete (e.g., UI renderer alongside status command after foundational telemetry ready).

## Parallel Execution Examples

- After completing Foundational tasks, run T014/T015 alongside T019/T020 in US1.
- In US2, T026 can run parallel to T027 once T025 ensures fixtures exist.
- In US3, T033 can execute while implementing T034, as they touch distinct files.

## Implementation Strategy

1. Deliver MVP by completing Setup + Foundational + User Story 1.
2. Iterate with resilience features in User Story 2.
3. Add operational tooling in User Story 3.
4. Finish with polish tasks (docs, performance scripts, final verification).
