# Research Log — SpecMuxer tmux AI Orchestrator

## Tasks Dispatched

- [x] Research reuse of adapter/session utilities from `ccmodel/cmd/execcmd` for SpecMuxer.
- [x] Research tmux integration testing approach for SpecMuxer’s Go CLI.
- [x] Identify best practices for automating tmux from Go CLIs.
- [x] Identify best practices for Cobra-driven multi-command CLIs with resumable workflows.
- [x] Identify patterns for filesystem-backed session metadata, log rotation, and redaction.

## Findings

- **Decision**: Re-implement SpecMuxer adapter/session management while borrowing design patterns (session records, tmux helpers) from `ccmodel/cmd/execcmd`; do not import ccmodel packages directly.  
  **Rationale**: SpecMuxer requires project-scoped `.specmuxer/` storage, configurable adapters, and stricter governance (permissions, runbook hooks) than ccmodel’s global sessions. Direct reuse would force cross-repo dependency and limit schema changes. Leveraging the MIT-licensed patterns ensures familiarity while keeping architecture SOLID.  
  **Alternatives considered**: (1) Vendor ccmodel’s `execcmd` package wholesale—rejected due to mismatched storage paths and tighter resume semantics; (2) Implement from scratch without reference—rejected because ccmodel offers proven tmux guardrails worth replicating conceptually.

- **Decision**: Use real tmux servers within integration tests by launching isolated sockets (`tmux -L specmuxer-test`) and teardown after assertions; supplement with unit tests that mock the tmux command runner.  
  **Rationale**: Resume and UI behaviors rely on tmux features (pipe-pane, session creation) that mocks can’t fully represent. Running tmux in CI is feasible on Linux/macOS without root, and we can gate tests when tmux missing. Unit tests cover formatting/serialization without tmux.  
  **Alternatives considered**: (1) Mock all tmux invocations—rejected because it risks regressions in attach/log piping; (2) Shell-scripted expect harness—rejected for Go-centric workflow and complexity; (3) Require manual testing—rejected due to Relentless Verification principle.

- **Decision**: Interact with tmux via explicit command wrappers that accept context (socket name, env) and capture stderr for doctor logging; ensure idempotent session creation and environment sync (mirroring `ensureTmuxSession`, `syncTmuxEnvironment` patterns).  
  **Rationale**: The reference project demonstrates reliable handling of tmux absent scenarios and environment propagation. Wrapping commands enables re-use across run/resume/gc flows and simplifies logging.  
  **Alternatives considered**: (1) Embed tmux as library through libtmux bindings—rejected due to portability and maintenance; (2) Use screen/other multiplexers—out of scope.

- **Decision**: Structure the CLI with Cobra root command (`specmuxer`) and subcommands matching required verbs; centralize shared flags/env injection; reuse cmdux (if compatible) for rich terminal UX while keeping adapter registry decoupled.  
  **Rationale**: Cobra provides mature command discovery, completions, and aligns with ccmodel precedent. cmdux offers status bars/sidebars already solving interactive demands.  
  **Alternatives considered**: (1) Implement bespoke CLI parser—rejected for time/safety; (2) Use urfave/cli—less compatible with existing tooling and completions.

- **Decision**: Store session metadata as YAML (vs JSON) with explicit schemas, enforce directory permissions on creation, apply log rotation via size+daily segments, and allow regex-based redaction using Go’s regex library.  
  **Rationale**: YAML matches spec requirement and supports comments/custom resume hooks. Applying rotation/permissions at write-time meets Operational Excellence. Redaction rules mirror patterns from existing internal tooling.  
  **Alternatives considered**: (1) Keep JSON for compatibility—rejected because spec mandates `.yml`; (2) Use embedded DB (boltdb)—overkill for single-host CLI; (3) Defer redaction—violates observability requirements.

- **Decision**: Detect non交互终端场景并在 run/attach 阶段优雅降级，提供 socket/session 提示，同时在 tmux 客户端内绑定 stdin/stdout/stderr 以匹配 `ccmodel` 的 attach 体验。  
  **Rationale**: 自动 attach 在 CI 或 API 调用场景下必须避免 `open terminal failed`，同时保留真实终端里的快捷体验；统一由 tmux 客户端封装可以保持 DRY。  
  **Alternatives considered**: (1) 完全禁用自动 attach——会降低交互体验； (2) 在 CLI 层手动 fork `tmux attach`——会复制逻辑并破坏层次，故统一在 tmux 客户端实现。

- **Decision**: Align codex adapter defaults with the modern Codex CLI (`codex` as entrypoint, `codex resume --last`, `codex --version` for health) and drop nonexistent stop/snapshot commands.  
  **Rationale**: The CLI no longer exposes `codex run --interactive` or `codex stop`; using them caused sessions to exit immediately. Matching the actual tool semantics keeps sessions alive and preserves health checks even in non-interactive environments.  
  **Alternatives considered**: (1) Retain legacy commands and expect downstream overrides—rejected because out-of-the-box UX would remain broken; (2) shell-wrap Codex through shims—unnecessary once definitions mirror the real CLI.

## Resolved Clarifications

- `NEEDS CLARIFICATION: reusable adapter utilities from ccmodel/execcmd?` → We will re-implement with inspiration; no direct dependency required.
- `NEEDS CLARIFICATION: approach for tmux-integration E2E (mock vs. real tmux fixture)` → Adopt real tmux socket integration tests with conditional skip fallback.
