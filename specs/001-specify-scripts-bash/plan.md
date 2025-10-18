# Implementation Plan: SpecMuxer tmux AI Orchestrator

**Branch**: `001-specify-scripts-bash` | **Date**: 2025-10-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-specify-scripts-bash/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build “SpecMuxer,” a tmux-backed orchestrator that launches, resumes, observes, and maintains multiple AI coding tool sessions per project with resilient storage in `.specmuxer/`. The MVP must deliver the nine core CLI commands, durable session metadata, adapter-driven recovery, tmux UI augmentations, GC/doctor operations, and open-source quality assets while holding performance SLOs (status ≤300 ms P95, resume ≤10 s for 20 sessions).

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: Go 1.23 module with pinned toolchain go1.24.2 (align with ccmodel reference)  
**Primary Dependencies**: tmux CLI (>=3.x), github.com/spf13/cobra for CLI, github.com/bagaking/cmdux for command UX, github.com/tidwall/gjson for JSON handling (status JSON export)  
**Storage**: Local filesystem (.specmuxer/ conf.yml, stats.yml, sessions/*.yml, logs/*.log with rotation)  
**Testing**: go test ./... with race + coverage, integration harness launching ephemeral tmux sockets, tmux-dependent tests skipped if tmux unavailable  
**Target Platform**: Linux and macOS hosts with tmux ≥ 3.x; best-effort via WSL  
**Project Type**: Single CLI/binary project with internal packages (cmd/, pkg/)  
**Performance Goals**: status/top P95 <300 ms for ≤50 sessions; resume ≤10 s for ≤20 sessions, GC adoption guidance under 2 s  
**Constraints**: No root requirement, logs redacted via regex, permissions 0700/0600, offline capable  
**Scale/Scope**: Tens of concurrent sessions per project (baseline 50), multi-project isolation within same host

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Confirm the feature uses the pinned Go toolchain and defines required modules
      without introducing unapproved dependencies.  
      *Action*: Initialize go.mod with Go 1.23/toolchain go1.24.2, reuse vetted cobra/cmdux dependencies from ccmodel, submit dependency audit in PR.
- [x] Define the automated test plan that delivers ≥95% coverage, includes `go test ./... -race`,
      and identifies new lint rules or exceptions for council review.  
      *Action*: Configure Go test matrix (unit for storage/adapters, integration harness hitting tmux sandbox), enforce golangci-lint strict preset, add coverage enforcement script.
- [x] Map the change onto the Interface → Application → Domain → Infrastructure layers and
      articulate how boundaries stay intact.  
      *Action*: Interface (cmd) → Application (orchestrator services) → Domain (session/adapters) → Infrastructure (tmux, filesystem); adapters only depend downward via interfaces.
- [x] Capture DRY/SOLID considerations: reusable packages touched, responsibilities per component,
      and duplication slated for elimination.  
      *Action*: Centralize tmux + filesystem utilities under pkg/runtime, adapter registry per tool, ensure session manager single responsibility; reuse logging/CLI scaffolding instead of duplicating ccmodel logic.
- [x] Specify observability and performance deliverables: metrics, logs, runbooks, and budgets
      that will ship with the feature.  
      *Action*: Define stats.yml schema, CLI status metrics, log rotation monitoring, runbook covering resume, GC, doctor checks, and performance validation scripts.

## Project Structure

### Documentation (this feature)

```
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```
# [REMOVE IF UNUSED] Option 1: Single project (DEFAULT)
src/
├── models/
├── services/
├── cli/
└── lib/

tests/
├── contract/
├── integration/
└── unit/

# [REMOVE IF UNUSED] Option 2: Web application (when "frontend" + "backend" detected)
backend/
├── src/
│   ├── models/
│   ├── services/
│   └── api/
└── tests/

frontend/
├── src/
│   ├── components/
│   ├── pages/
│   └── services/
└── tests/

# [REMOVE IF UNUSED] Option 3: Mobile + API (when "iOS/Android" detected)
api/
└── [same as backend above]

ios/ or android/
└── [platform-specific structure: feature modules, UI flows, platform tests]
```

**Structure Decision**: Single CLI binary rooted under `cmd/specmuxer/` with supporting packages:

```
cmd/specmuxer/
pkg/
├── adapters/
├── orchestrator/
├── runtime/         # tmux + filesystem glue
├── telemetry/       # status/top aggregation
└── ui/              # tmux status/sidebar rendering
internal/
└── applier/         # GC, doctor, resume workflows
```

Tests mirror packages under `pkg/...` with integration cases in `internal/integration/`.

## Complexity Tracking

*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
