package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	gc "github.com/bagaking/specmuxer/internal/applier/gc"
	"github.com/bagaking/specmuxer/pkg/ui/sidebar"
)

func init() {
	rootCmd.AddCommand(newGCCommand())
}

func newGCCommand() *cobra.Command {
	var adopt bool

	cmd := &cobra.Command{
		Use:   "gc",
		Short: "Inspect or adopt orphaned sessions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			deps, err := loadRuntime()
			if err != nil {
				return err
			}

			service := gc.NewService(gc.ServiceConfig{Store: deps.store, Tmux: deps.tmux})
			ctx := cmd.Context()
			if ctx == nil || ctx == context.Background() {
				ctx = context.TODO()
			}

			var summary gc.Summary
			if adopt {
				summary, err = service.Adopt(ctx)
			} else {
				summary, err = service.DryRun(ctx)
			}
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Orphans detected: %v\n", summary.Orphans)
			if adopt {
				fmt.Fprintf(cmd.OutOrStdout(), "Removed: %v\n", summary.Removed)
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), sidebar.RenderGCPrompt(summary.Orphans))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&adopt, "adopt", false, "Adopt orphan records instead of dry-run")
	return cmd
}
