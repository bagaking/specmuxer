# Adapter Guide

SpecMuxer ships with `codex` and `claude` adapter definitions inside `pkg/adapters`. Each adapter exposes lifecycle commands:

| Adapter | Start Command | Resume Command | Health Check | Extract State | Supports Attach |
|---------|---------------|----------------|--------------|---------------|-----------------|
| codex   | `codex` | `codex resume --last` (fallbacks to start if override absent) | `codex --version` | _n/a_ | Yes |
| claude  | `claude workbench start` | `claude workbench resume` | `claude workbench status` | `claude workbench export` | Yes |

## Overriding Lifecycle Commands

Project-level overrides live in `specmuxer/conf.yml` under `adapter_overrides`. Example:

```yaml
adapter_overrides:
  codex:
    start: ["sh", "-lc", "CONDA_DEFAULT_ENV=spec env-codex"]
    resume: ["sh", "-lc", "env-codex resume"]
    timeout_seconds: 120
```

Overrides flow through `run.Service` and `resume.Service` via `adapters.Definition.WithOverrides`, ensuring your custom commands are deep-cloned and environment variables preserved.

## Adding a New Adapter

1. Register it via `adapters.NewRegistry()` (see `pkg/adapters/registry.go`).
2. Provide `Start` command at minimum; optional `Resume`, `Health`, `Stop`, `ExtractState` improve resilience.
3. Supply defaults for attach support and environment variables.
4. Add documentation/examples here and extend integration coverage.
