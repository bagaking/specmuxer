# Changelog

## 0.1.0 - 2025-10-18
- Bootstrap Go module with Cobra-based CLI (`run`, `status`, `top`, `attach`).
- Persist sessions in `.specmuxer/` with YAML storage, log rotation, and telemetry collectors.
- Add resume workflow (`specmuxer resume`) with adapter overrides and resume success tracking.
- Ship maintenance tooling: log tailing, GC dry-run/adopt, doctor diagnostics, and sidebar prompts.
- Document architecture, adapters, recovery steps, and FAQ; add install target and performance harness.
