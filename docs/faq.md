# Troubleshooting FAQ

**Q: `specmuxer run` fails because tmux cannot connect to the socket.**
- Ensure `tmux >= 3.x` is installed and accessible on `PATH`.
- Delete stale sockets under `.specmuxer/*.sock` and retry.

**Q: Sessions resume unexpectedly after restart.**
- Check `resume_mode` in `.specmuxer/conf.yml`. Set to `manual` for opt-in behaviour or `automatic` to enable auto-resume.

**Q: Logs contain sensitive tokens.**
- Add regexes to `redaction_rules` in `conf.yml` (e.g. `sk-[A-Za-z0-9]+`).
- Use `specmuxer logs --tail` to verify redaction output.

**Q: GC reports orphans that I want to keep.**
- Orphans indicate SpecMuxer cannot see a running tmux session. If the pane is managed manually, ignore the entry. Adopt only when you want the YAML record removed.

**Q: Doctor warns about workspace permissions.**
- Ensure the workspace (and `.specmuxer/`) is owned by your user with 0700 permissions as enforced by `config.Load`.
