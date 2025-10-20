# Implementation Plan: SpecMuxer tmux AI Orchestrator

**Branch**: `001-specify-scripts-bash` | **Date**: 2025-10-19 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-specify-scripts-bash/spec.md`

## Summary

Deliver a CLI-first tmux orchestrator that can launch, resume, monitor, and recover AI coding tool sessions per project. The implementation builds on Go 1.23 with a pinned go1.24.2 toolchain, persists runtime state under `.specmuxer/`, streams telemetry, and hardens session UX for non交互终端 attach、active-session确认、以及 tmux 会话缺失后的恢复指引。

## Technical Context

**Language/Version**: Go 1.23 (toolchain go1.24.2)  
**Primary Dependencies**: tmux ≥3.x CLI, `github.com/spf13/cobra`, `github.com/bagaking/cmdux`, `github.com/tidwall/gjson`, Go stdlib  
**Storage**: Project-local filesystem `.specmuxer/` (YAML session + stats, rotated logs with redaction)  
**Testing**: `go test ./... -race -count=1`, golangci-lint strict profile, integration harness launching tmux sockets  
**Target Platform**: macOS & Linux hosts (best-effort WSL) with tmux ≥3.x  
**Project Type**: Single Go CLI binary under `cmd/specmuxer` with layered `pkg/` modules  
**Performance Goals**: `status` P95 ≤300 ms for ≤50 sessions; `resume --all` ≤10 s for ≤20 sessions; attach/launch flows must degrade gracefully when tmux panes vanish or terminal不可交互  
**Constraints**: Offline-capable, no root requirement, enforce `.specmuxer/` perms 0700/0600, ≥95% coverage on touched packages  
**Scale/Scope**: Tens of concurrent sessions per project (~50), multi-project isolation, extensible adapters (Codex, Claude, future tools)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Confirm the feature uses the pinned Go toolchain and defines required modules without introducing unapproved dependencies. (go.mod already pins go1.24.2; new changes stay within existing modules and vetted deps.)
- [x] Define the automated test plan that delivers ≥95% coverage, includes `go test ./... -race`, and identifies new lint rules or exceptions for council review. (Unit tests + tmux-backed integration suites remain required; no lint waivers requested.)
- [x] Map the change onto the Interface → Application → Domain → Infrastructure layers and articulate how boundaries stay intact. (`cmd/` commands → orchestrator services → domain models → runtime adapters; no upward imports.)
- [x] Capture DRY/SOLID considerations: reusable packages touched, responsibilities per component, and duplication slated for elimination. (Shared helpers centralize terminal detection/prompting; tmux interactions encapsulated in runtime client.)
- [x] Specify observability and performance deliverables: metrics, logs, runbooks, and budgets that will ship with the feature. (Telemetry collector persists stats, logs rotated & redacted, docs/recovery.md reflects attach/resume pathways and performance budgets.)

## Project Structure

### Documentation (this feature)

```
specs/001-specify-scripts-bash/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── tasks.md
```

### Source Code (repository root)

```
cmd/specmuxer/
├── main.go
├── runtime.go
├── run_cmd.go
├── attach_cmd.go
├── resume_cmd.go
├── status_cmd.go
├── top_cmd.go
├── helpers.go
└── helpers_test.go

pkg/
├── adapters/
├── domain/session/
├── orchestrator/
│   ├── run/
│   └── resume/
├── runtime/
│   ├── storage/
│   └── tmux/
├── telemetry/
│   ├── logs/
│   └── stats/
└── ui/
    ├── sidebar/
    └── statusbar/

internal/
└── integration/

.specmuxer/
├── conf.yml
├── sessions/
├── logs/
└── stats.yml
```

**Structure Decision**: Continue with a single Go CLI project: interface commands under `cmd/specmuxer`, application services in `pkg/orchestrator`, domain models in `pkg/domain/session`, infrastructure adapters in `pkg/runtime` & `pkg/telemetry`, and integration tests in `internal/integration`.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| _None_ |  |  |
