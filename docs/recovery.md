# Recovery Walkthrough

## After Host Reboot

1. Re-enter the workspace (`cd /path/to/project`).
2. Run `specmuxer status --json` to inspect persisted sessions under `.specmuxer/`. The tabular output also includes a `TMUX` indicator—`missing` tells you the YAML record exists but tmux panes need to be recreated.
3. Execute `specmuxer resume --all`:
   - Eligible sessions (`user_killed=false`) start via adapter-specific resume commands.
   - Skipped sessions are reported with reasons (user killed, missing adapter, etc.).
4. Verify health with `specmuxer top` and inspect logs (`specmuxer logs --session <id> --tail 40`). If you trigger `specmuxer attach` from a non-interactive shell (CI, background service), it now prints the tmux socket + session name instead of failing so you can attach later from a terminal. Likewise, `specmuxer run` warns when sessions are already active; add `--force` when driving recovery workflows non-interactively.

## Resuming Specific Sessions

```
specmuxer resume sess-123 sess-456
```

Useful when only a subset should be restored or when dry-running (`--dry-run`) to preview actions.

## Handling Orphaned Sessions

1. Run `specmuxer gc` to list orphaned tmux sessions and stray YAML records.
2. Review the prompt rendered by the GC sidebar helper; adopt entries with `specmuxer gc --adopt` to prune records.
3. If orphans represent unmanaged tmux panes you wish to keep, ignore them and rerun GC when ready.

## Doctor Checks

```
specmuxer doctor
```

Outputs tmux availability, workspace accessibility, and remediation guidance. Use this before filing a support ticket to ensure prerequisites are satisfied.
