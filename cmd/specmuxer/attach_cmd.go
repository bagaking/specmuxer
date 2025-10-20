package main

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/bagaking/specmuxer/pkg/domain/session"
	"github.com/bagaking/specmuxer/pkg/runtime/tmux"
	"github.com/bagaking/specmuxer/pkg/telemetry/stats"
)

func init() {
	rootCmd.AddCommand(newAttachCommand())
}

func newAttachCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach [session]",
		Short: "Attach to a running tmux session",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deps, err := loadRuntime()
			if err != nil {
				return err
			}

			records, err := deps.runSvc.Sessions()
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			liveInfo := computeLiveness(ctx, deps.tmux, records, deps.runSvc.SocketPath())
			deps.collector.SetLiveness(livenessAsBool(liveInfo))
			defer deps.collector.SetLiveness(nil)

			if len(args) == 0 {
				return printSessionChoices(cmd, deps.collector, records, liveInfo)
			}

			target := strings.TrimSpace(args[0])
			if target == "" {
				return printSessionChoices(cmd, deps.collector, records, liveInfo)
			}

			record, ok := findAttachTarget(records, target)
			if !ok {
				fmt.Fprintf(cmd.ErrOrStderr(), "Session %s not found.\n\n", target)
				return printSessionChoices(cmd, deps.collector, records, liveInfo)
			}

			info := liveInfo[record.ID]
			sessionName := info.Session
			socket := deps.runSvc.SocketPath()
			if record.Tmux.Socket != "" {
				socket = record.Tmux.Socket
			}

			if !info.Alive {
				fmt.Fprintf(cmd.ErrOrStderr(), "tmux session %s (socket %s) not found. Try 'specmuxer resume %s' to recreate it.\n", sessionName, socket, record.ID)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Attaching to session %s (tmux: %s, socket: %s)\n", record.ID, sessionName, socket)
			if !hasInteractiveTerminal() {
				fmt.Fprintf(cmd.ErrOrStderr(), "Interactive terminal required to attach.\n")
				fmt.Fprintf(cmd.OutOrStdout(), "Use 'tmux -S %s attach-session -t %s' from a terminal.\n", socket, sessionName)
				return nil
			}
			return deps.tmux.Attach(ctx, tmux.AttachOptions{
				Session: sessionName,
				Socket:  socket,
			})
		},
	}
	return cmd
}

func printSessionChoices(cmd *cobra.Command, collector *stats.Collector, records []session.SessionRecord, live map[string]LivenessInfo) error {
	if len(records) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No sessions available to attach.")
		return nil
	}
	snapshot := collector.BuildSnapshot(records)
	fmt.Fprintln(cmd.OutOrStdout(), stats.FormatTable(snapshot, true))
	fmt.Fprintln(cmd.OutOrStdout())
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "SESSION ID\tTMUX SESSION\tSOCKET\tUPDATED")
	for _, rec := range records {
		info := live[rec.ID]
		last := "-"
		if rec.LastOutputAt != nil {
			last = rec.LastOutputAt.Format(time.RFC3339)
		}
		tmuxSession := info.Session
		if tmuxSession == "" {
			tmuxSession = rec.Tmux.Session
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", rec.ID, tmuxSession, rec.Tmux.Socket, last)
	}
	_ = w.Flush()
	return nil
}

func findAttachTarget(records []session.SessionRecord, target string) (session.SessionRecord, bool) {
	if rec, ok := findSession(records, target); ok {
		return rec, true
	}
	for _, rec := range records {
		if rec.Tmux.Session == target {
			return rec, true
		}
	}
	return session.SessionRecord{}, false
}
