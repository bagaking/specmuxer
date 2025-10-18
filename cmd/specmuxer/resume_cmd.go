package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bagaking/specmuxer/pkg/orchestrator/resume"
)

func init() {
	rootCmd.AddCommand(newResumeCommand())
}

func newResumeCommand() *cobra.Command {
	var (
		all    bool
		dryRun bool
	)
	var sessions []string

	cmd := &cobra.Command{
		Use:   "resume",
		Short: "Resume interrupted sessions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			deps, err := loadRuntime()
			if err != nil {
				return err
			}

			opts := resume.Options{
				All:        all,
				DryRun:     dryRun,
				SessionIDs: normalizeIDs(sessions),
			}

			if !opts.All && len(opts.SessionIDs) == 0 {
				return fmt.Errorf("specify --all or session IDs")
			}

			summary, err := deps.resumeSvc.Resume(cmd.Context(), opts)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Resume attempted:%d succeeded:%d skipped:%d\n", summary.Attempted, summary.Succeeded, summary.Skipped)
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Resume all eligible sessions")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate resume without launching sessions")
	cmd.Flags().StringSliceVar(&sessions, "session", nil, "Specific session IDs to resume")
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
