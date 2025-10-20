package main

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/bagaking/specmuxer/pkg/domain/session"
	"github.com/bagaking/specmuxer/pkg/orchestrator/resume"
	"github.com/bagaking/specmuxer/pkg/telemetry/stats"
)

func init() {
	rootCmd.AddCommand(newResumeCommand())
}

func newResumeCommand() *cobra.Command {
	var (
		all    bool
		dryRun bool
	)

	cmd := &cobra.Command{
		Use:   "resume [sessionID]...",
		Short: "Resume interrupted sessions",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			deps, err := loadRuntime()
			if err != nil {
				return err
			}

			records, err := deps.store.ListSessions()
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			liveInfo := computeLiveness(ctx, deps.tmux, records, deps.runSvc.SocketPath())

			ids := normalizeIDs(args)
			if !all && len(ids) == 0 {
				return printResumeChoices(cmd, deps, records, liveInfo)
			}

			opts := resume.Options{
				All:        all,
				DryRun:     dryRun,
				SessionIDs: ids,
			}

			summary, err := deps.resumeSvc.Resume(cmd.Context(), opts)
			if err != nil {
				return err
			}

			if summary.Attempted == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No matching sessions to resume.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Resume attempted:%d succeeded:%d skipped:%d\n", summary.Attempted, summary.Succeeded, summary.Skipped)
			printResumeSummary(cmd, summary)
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Resume all eligible sessions")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate resume without launching sessions")
	return cmd
}

func normalizeIDs(ids []string) []string {
	var result []string
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func printResumeChoices(cmd *cobra.Command, deps *runtimeDeps, records []session.SessionRecord, liveInfo map[string]LivenessInfo) error {
	deps.collector.SetLiveness(livenessAsBool(liveInfo))
	defer deps.collector.SetLiveness(nil)

	collectorSnapshot := deps.collector.BuildSnapshot(records)
	fmt.Fprint(cmd.OutOrStdout(), stats.FormatTable(collectorSnapshot, true))
	fmt.Fprintln(cmd.OutOrStdout())

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "SESSION ID\tADAPTER\tSTATUS\tTMUX\tUSER KILLED\tLAST OUTPUT")
	eligible := 0
	for _, rec := range records {
		def, err := resolveDefinition(deps.config, deps.registry, rec.Tool)
		if err != nil {
			continue
		}
		info := liveInfo[rec.ID]
		alive := info.Alive
		if ok, _ := resume.EvaluateEligibility(rec, def, alive); !ok {
			continue
		}
		eligible++
		tmuxState := "present"
		if !alive {
			tmuxState = "missing"
		}
		last := "-"
		if rec.LastOutputAt != nil {
			last = rec.LastOutputAt.Format(time.RFC3339)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%t\t%s\n",
			rec.ID,
			rec.Tool,
			rec.Status,
			tmuxState,
			rec.UserKilled != nil && *rec.UserKilled,
			last,
		)
	}
	if eligible == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No sessions are eligible for resume.")
		return nil
	}
	_ = w.Flush()
	fmt.Fprintln(cmd.OutOrStdout(), "\nUse 'specmuxer resume <id>' to resume specific sessions, or '--all' to resume everything.")
	return nil
}

func printResumeSummary(cmd *cobra.Command, summary resume.Summary) {
	if len(summary.SucceededIDs) > 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "Resumed:")
		for _, id := range summary.SucceededIDs {
			fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", id)
		}
	}

	if len(summary.SkippedSessions) > 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "Skipped:")
		w := tabwriter.NewWriter(cmd.OutOrStderr(), 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "SESSION ID\tREASON")
		for _, skip := range summary.SkippedSessions {
			fmt.Fprintf(w, "%s\t%s\n", skip.ID, skip.Reason)
		}
		_ = w.Flush()
	}
}
