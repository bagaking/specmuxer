package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/bagaking/specmuxer/pkg/telemetry/logs"
)

func init() {
	rootCmd.AddCommand(newLogsCommand())
}

func newLogsCommand() *cobra.Command {
	var (
		sessionID string
		tail      int
		follow    bool
	)

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Print session logs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if sessionID == "" {
				return fmt.Errorf("--session is required")
			}
			deps, err := loadRuntime()
			if err != nil {
				return err
			}

			reader := logs.NewReader(deps.config.Paths().LogsDir, deps.config.RedactionRules...)
			ctx := cmd.Context()
			if !follow {
				return printLogSnapshot(cmd, reader, sessionID, tail)
			}
			return followLogs(ctx, cmd, reader, sessionID, tail)
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "Session ID to tail")
	cmd.Flags().IntVar(&tail, "tail", 20, "Number of lines to show")
	cmd.Flags().BoolVar(&follow, "follow", false, "Stream logs until interrupted")
	return cmd
}

func printLogSnapshot(cmd *cobra.Command, reader *logs.Reader, sessionID string, limit int) error {
	lines, err := reader.LastLines(sessionID, limit)
	if err != nil {
		return err
	}
	for _, line := range lines {
		fmt.Fprintln(cmd.OutOrStdout(), line)
	}
	return nil
}

func followLogs(ctx context.Context, cmd *cobra.Command, reader *logs.Reader, sessionID string, limit int) error {
	seen := 0
	for {
		lines, err := reader.LastLines(sessionID, limit)
		if err != nil {
			return err
		}
		if len(lines) > seen {
			for _, line := range lines[seen:] {
				fmt.Fprintln(cmd.OutOrStdout(), line)
			}
			seen = len(lines)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}
