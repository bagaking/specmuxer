---
description: "Task list for SpecMuxer tmux AI Orchestrator implementation"
---

# Tasks: SpecMuxer tmux AI Orchestrator

**Input**: Design documents from `/specs/001-specify-scripts-bash/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are MANDATORY. Each story begins with automated tests that enforce coverage ≥95%, `go test ./... -race`, and lint gates.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions
- **Single project**: `cmd/`, `pkg/`, `internal/`, `docs/`, `scripts/`, `specs/`
- Paths shown below assume single project - adjust based on plan.md structure

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Initialize Go module with go1.23/toolchain go1.24.2 in go.mod
- [X] T002 Scaffold Cobra root command with persistent flags in cmd/specmuxer/root.go
- [X] T003 Add CLI entrypoint wiring root command in cmd/specmuxer/main.go
- [X] T004 Add lint/test/coverage make targets enforcing constitution gates in Makefile

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T005 Implement configuration loader with permission checks in pkg/orchestrator/config/config.go
- [X] T006 Model session domain structs with YAML tags in pkg/domain/session/session.go
- [X] T007 Implement YAML persistence driver for sessions/stats in pkg/runtime/storage/yaml_store.go
- [X] T008 Implement log rotation and redaction manager in pkg/telemetry/logs/manager.go
- [X] T009 Implement tmux client wrappers for session/window ops in pkg/runtime/tmux/client.go
- [X] T010 Implement adapter registry with default codex/claude profiles in pkg/adapters/registry.go
- [X] T011 Implement telemetry collector for status/top metrics in pkg/telemetry/stats/collector.go
- [X] T012 [P] Add unit tests for session serialization round-trip in pkg/domain/session/session_test.go
- [X] T013 [P] Add unit tests for tmux client error handling in pkg/runtime/tmux/client_test.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Launch and Monitor Tool Sessions (Priority: P1) 🎯 MVP

**Goal**: Developers can start AI tool sessions, pass env/args, and observe real-time activity via status/top/attach with tmux UI enhancements.

**Independent Test**: Launch `specmuxer run codex -- --plan` in a clean workspace, verify session creation, status/top reporting, attach behavior, and per-session logging without resume/gc features.

### Tests for User Story 1 (MANDATORY) ⚠️

**NOTE: Write these tests FIRST, ensure they FAIL before implementation. Do not begin implementation until coverage impact is understood.**

- [X] T014 [US1] Write integration test covering run→status lifecycle in internal/integration/run_status_test.go
- [X] T015 [P] [US1] Add unit tests for status JSON formatting in pkg/telemetry/stats/formatter_test.go

### Implementation for User Story 1

- [X] T016 [US1] Implement run orchestrator service launching tmux sessions in pkg/orchestrator/run/service.go
- [X] T017 [US1] Wire run Cobra command with `--` and `--env` handling in cmd/specmuxer/run.go
- [X] T018 [US1] Implement status command with table/json/wide outputs in cmd/specmuxer/status.go
- [X] T019 [US1] Implement top streaming loop with idle threshold checks in pkg/telemetry/stats/top.go
- [X] T020 [US1] Implement attach command with session selector in cmd/specmuxer/attach.go
- [X] T021 [US1] Implement per-session log writer with timestamps in pkg/orchestrator/run/log_writer.go
- [X] T022 [P] [US1] Implement status/sidebar UI renderer in pkg/ui/statusbar/statusbar.go

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Resume After Interruptions (Priority: P2)

**Goal**: Interrupted sessions can be resumed automatically or manually with adapter-specific context and user-killed safeguards.

**Independent Test**: Simulate tmux restart, run `specmuxer resume --all`, confirm eligible sessions resume with hooks while user_killed sessions remain stopped.

### Tests for User Story 2 (MANDATORY) ⚠️

**NOTE: Ensure new tests enforce layering boundaries and reuse existing fixtures where possible.**

- [X] T023 [US2] Write integration test for resume --all recovery in internal/integration/resume_test.go
- [X] T024 [P] [US2] Add unit tests for resume eligibility filtering in pkg/orchestrator/resume/filter_test.go

### Implementation for User Story 2

- [X] T025 [US2] Implement resume service applying adapter state/hooks in pkg/orchestrator/resume/service.go
- [X] T026 [US2] Wire resume Cobra command with `--all` and `--dry-run` in cmd/specmuxer/resume.go
- [X] T027 [US2] Persist user_killed flag and resume hooks in session YAML writer pkg/domain/session/session_yaml.go
- [X] T028 [US2] Implement adapter resume plumbing including extract_state in pkg/adapters/registry_resume.go
- [X] T029 [US2] Add configuration option for automatic resume in pkg/orchestrator/config/autoresume.go
- [X] T030 [P] [US2] Extend telemetry collector to record resume metrics in pkg/telemetry/stats/resume_metrics.go

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Maintain and Inspect System Health (Priority: P3)

**Goal**: Maintainers can inspect logs, run GC with adopt/ignore, and execute doctor checks to keep operations healthy.

**Independent Test**: Run sessions, rotate logs, execute `gc --dry-run`, follow logs, and run `doctor`; ensure actionable guidance, adoption prompts, and permission checks work without resume/run.

### Tests for User Story 3 (MANDATORY) ⚠️

**NOTE: Extend observability assertions to cover new metrics/logging introduced by this story.**

- [ ] T031 [US3] Write integration test covering logs/gc/doctor flow in internal/integration/logs_gc_test.go
- [ ] T032 [P] [US3] Add unit tests for doctor check matrix in internal/applier/doctor/checks_test.go

### Implementation for User Story 3

- [ ] T033 [US3] Implement logs command with redaction and follow cursor in cmd/specmuxer/logs.go
- [ ] T034 [US3] Implement log rotation worker bound to manager in pkg/telemetry/logs/rotation.go
- [ ] T035 [US3] Implement GC workflow detecting orphan sessions in internal/applier/gc/gc.go
- [ ] T036 [US3] Implement adopt/ignore prompt UI in pkg/ui/sidebar/gc_prompt.go
- [ ] T037 [US3] Implement doctor checks for tmux/PATH/permissions/config in internal/applier/doctor/checks.go
- [ ] T038 [US3] Publish stats snapshot including resume success rate in pkg/telemetry/stats/snapshot.go

**Checkpoint**: All user stories should now be independently functional

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T039 Update architecture overview with layering narrative in docs/concepts.md
- [ ] T040 [P] Generate adapters guide covering overrides in docs/adapters.md
- [ ] T041 [P] Document recovery walkthrough in docs/recovery.md
- [ ] T042 [P] Compile troubleshooting FAQ in docs/faq.md
- [ ] T043 Add changelog entry and version bump in CHANGELOG.md
- [ ] T044 Refresh README quickstart/CLI table in README.md
- [ ] T045 Create performance verification script for status/top in scripts/perf/status_benchmark.go
- [ ] T046 Validate quickstart instructions against live CLI in specs/001-specify-scripts-bash/quickstart.md

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Depends on completion of User Story 1 logging infrastructure for rehydration
- **User Story 3 (P3)**: Depends on User Story 1 logging + telemetry and User Story 2 session metadata

### Within Each User Story

- Tests MUST be written first, FAIL initially, and deliver ≥95% coverage before implementation proceeds
- Models before services
- Services before endpoints/commands
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, run/status UI (T022) and status formatter tests (T015) can proceed while core run logic is implemented
- Resume metrics extension (T030) can run parallel to resume CLI wiring (T026) once service stub exists
- Doctor unit tests (T032) can run parallel to GC workflow implementation (T035)

---

## Parallel Example: User Story 1

```bash
# Launch mandatory tests for User Story 1 together:
Task: "Write integration test covering run→status lifecycle in internal/integration/run_status_test.go"
Task: "Add unit tests for status JSON formatting in pkg/telemetry/stats/formatter_test.go"

# Launch all models/services for User Story 1 together:
Task: "Implement run orchestrator service launching tmux sessions in pkg/orchestrator/run/service.go"
Task: "Implement status/sidebar UI renderer in pkg/ui/statusbar/statusbar.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1
   - Developer B: User Story 2
   - Developer C: User Story 3
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Drive TDD: start with failing tests, then implement until `go test ./... -race -count=1` passes
- Commit after each task or logical group once lint and tests are green
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
