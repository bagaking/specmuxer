package main

import (
	"context"
	"errors"
	"time"

	"github.com/spf13/cobra"

	"github.com/bagaking/specmuxer/pkg/domain/session"
	"github.com/bagaking/specmuxer/pkg/telemetry/stats"
)

func init() {
	rootCmd.AddCommand(newTopCommand())
}

func newTopCommand() *cobra.Command {
	var interval time.Duration
	cmd := &cobra.Command{
		Use:   "top",
		Short: "Stream live status snapshots",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			deps, err := loadRuntime()
			if err != nil {
				return err
			}

			streamer := stats.NewStreamer(deps.collector)
			deps.collector.SetLiveness(nil)
			ctx := cmd.Context()
			if ctx == nil || ctx == context.Background() {
				ctx = context.TODO()
			}

			fetch := func() ([]session.SessionRecord, error) {
				return deps.store.ListSessions()
			}

			err = streamer.Stream(ctx, interval, fetch, cmd.OutOrStdout())
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		},
	}
	cmd.Flags().DurationVarP(&interval, "interval", "i", time.Second, "Refresh interval")
	return cmd
}
