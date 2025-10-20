# specmuxer Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-18

## Active Technologies
- Go 1.23 module with pinned toolchain go1.24.2 (align with ccmodel reference) + mux CLI (>=3.x), github.com/spf13/cobra for CLI, github.com/bagaking/cmdux for command UX, github.com/tidwall/gjson for JSON handling (status JSON export) (001-specify-scripts-bash)
- Go 1.23 (toolchain go1.24.2) + mux ≥3.x CLI, `github.com/spf13/cobra`, `github.com/bagaking/cmdux`, `github.com/tidwall/gjson`, Go stdlib (001-specify-scripts-bash)
- Project-local filesystem `.specmuxer/` (YAML session + stats, rotated logs with redaction) (001-specify-scripts-bash)

## Project Structure
```
src/
tests/
```

## Commands
# Add commands for Go 1.23 module with pinned toolchain go1.24.2 (align with ccmodel reference)

## Code Style
Go 1.23 module with pinned toolchain go1.24.2 (align with ccmodel reference): Follow standard conventions

## Recent Changes
- 001-specify-scripts-bash: Added Go 1.23 (toolchain go1.24.2) + mux ≥3.x CLI, `github.com/spf13/cobra`, `github.com/bagaking/cmdux`, `github.com/tidwall/gjson`, Go stdlib
- 001-specify-scripts-bash: Added Go 1.23 module with pinned toolchain go1.24.2 (align with ccmodel reference) + mux CLI (>=3.x), github.com/spf13/cobra for CLI, github.com/bagaking/cmdux for command UX, github.com/tidwall/gjson for JSON handling (status JSON export)

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
