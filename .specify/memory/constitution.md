<!--
Sync Impact Report
Version change: N/A → 1.0.0
Modified principles:
- Placeholder → I. Go Craftsmanship
- Placeholder → II. Relentless Verification
- Placeholder → III. Layered Scalability
- Placeholder → IV. DRY & SOLID Discipline
- Placeholder → V. Operational Excellence
Added sections:
- Engineering Standards
- Delivery Workflow
Removed sections:
- None
Templates requiring updates:
- ✅ .specify/templates/plan-template.md
- ✅ .specify/templates/spec-template.md (aligned; no change required)
- ✅ .specify/templates/tasks-template.md
- ✅ .specify/templates/checklist-template.md (aligned; no change required)
Follow-up TODOs:
- None
-->

# SpecMuxer Constitution

## Core Principles

### I. Go Craftsmanship
All production code MUST be written in idiomatic Go using modules with clear package
boundaries. Every change MUST pass `gofmt`, `go vet`, and `staticcheck`, and the
repository MUST compile via `go build ./...` on the pinned toolchain before merge.
Rationale: Strict adherence to Go conventions keeps the codebase approachable and
worthy of the "world's greatest" ambition.

### II. Relentless Verification
Teams MUST author automated tests before or alongside new behavior, enforce
`go test ./... -race -count=1`, and maintain ≥95% line coverage on new and changed
packages. CI MUST enforce `golangci-lint run` with the strict profile and block merges
on any lint, race, or coverage regression.
Rationale: Comprehensive verification guarantees uncompromising quality and protects
future velocity.

### III. Layered Scalability
The architecture MUST observe the Interface → Application → Domain → Infrastructure
layering: higher layers may depend only on the next layer down, and domain packages
MUST be free of framework or infrastructure imports. Cross-layer communication MUST
happen via well-defined interfaces and DTOs.
Rationale: Enforcing clean seams keeps the system composable, horizontally scalable,
and resilient to change.

### IV. DRY & SOLID Discipline
Reusable logic MUST move into shared packages once duplicated thrice, and every
component MUST satisfy SOLID: single responsibility, open to extension through small
interfaces, and dependencies injected via constructors. God objects and `interface{}`
catch-alls are prohibited without architecture council approval.
Rationale: DRY and SOLID practices prevent entropy, making the codebase predictable
and extensible.

### V. Operational Excellence
Every feature MUST declare performance budgets, structured logging, metrics, and
runbooks before release. Observability hooks MUST be exercised in automated tests, and
documentation MUST stay in lockstep with behavior to guarantee operability under load.
Rationale: Operational readiness is inseparable from greatness; the project must perform
flawlessly in production.

## Engineering Standards

- **Toolchain**: The repository MUST pin the active Go toolchain in `go.mod`; upgrades
  require an approved governance proposal that documents impact and rollout.
- **Lint & Static Analysis**: CI MUST run `golangci-lint run --config .golangci.yml`,
  `go vet ./...`, and organization-approved security linters on every change.
- **Testing Matrix**: Merge pipelines MUST execute
  `go test ./... -race -count=1 -coverprofile=coverage.out` and publish coverage
  artifacts; failures block merge.
- **Code Generation**: Generated artifacts MUST be reproducible via `go generate ./...`
  and committed with a checksum note in the PR description.
- **Dependencies**: External modules MUST include license classification in the change
  description; new dependencies require architecture council approval and
  SPDX-compatible licensing.

## Delivery Workflow

1. Draft specs and plans MUST satisfy the Constitution Check: documented layering impact,
   test strategy, observability additions, and dependency risk analysis.
2. Implementation MUST begin with failing tests, followed by minimal code to pass, then
   refactor to uphold DRY & SOLID.
3. Feature branches MUST maintain passing CI at all times; partial work merges behind
   feature flags only when they do not violate principles.
4. Release candidates MUST deliver updated runbooks, performance benchmarks, and
   compliance evidence for lint/test gating.

## Governance

- **Authority**: The architecture council stewards this constitution and reviews every
  principle for compliance each quarter.
- **Amendments**: Amendments require a written RFC referencing affected principles,
  council majority approval, and an explicit migration plan; upon acceptance, update the
  constitution and increment the semantic version according to impact.
- **Versioning Policy**: Adopt semantic versioning (MAJOR.MINOR.PATCH). MAJOR increments
  change or remove principles; MINOR adds principles or expands governance; PATCH refines
  wording without altering obligations.
- **Compliance Reviews**: PR reviewers MUST confirm Constitution adherence via the plan
  checklist; failing audits trigger remediation tasks prioritized as P0.

**Version**: 1.0.0 | **Ratified**: 2025-10-18 | **Last Amended**: 2025-10-18
