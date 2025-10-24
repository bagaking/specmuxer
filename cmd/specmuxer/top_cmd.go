package main

import (
	"context"
	"errors"
	"time"

	"github.com/spf13/cobra"

	"github.com/bagaking/specmuxer/pkg/domain/session"
	"github.com/bagaking/specmuxer/pkg/runtime/tmux"
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

			ctx := cmd.Context()
			if ctx == nil || ctx == context.Background() {
				ctx = context.TODO()
			}

			streamer := stats.NewStreamer(deps.collector)
			fetch := fetchSessionsWithLiveness(ctx, deps.store, deps.collector, deps.tmux, deps.runSvc.SocketPath())

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

type sessionLister interface {
	ListSessions() ([]session.SessionRecord, error)
}

func fetchSessionsWithLiveness(ctx context.Context, lister sessionLister, collector *stats.Collector, tmuxClient *tmux.Client, defaultSocket string) func() ([]session.SessionRecord, error) {
	return func() ([]session.SessionRecord, error) {
		records, err := lister.ListSessions()
		if err != nil {
			collector.SetLiveness(nil)
			return nil, err
		}
		liveInfo := computeLiveness(ctx, tmuxClient, records, defaultSocket)
		collector.SetLiveness(livenessAsBool(liveInfo))
		return records, nil
	}
}
