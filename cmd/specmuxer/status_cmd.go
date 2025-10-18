package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bagaking/specmuxer/pkg/telemetry/stats"
)

func init() {
	rootCmd.AddCommand(newStatusCommand())
}

func newStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show session status summary",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			deps, err := loadRuntime()
			if err != nil {
				return err
			}

			records, err := deps.store.ListSessions()
			if err != nil {
				return err
			}
			snapshot := deps.collector.BuildSnapshot(records)

			jsonOutput, _ := cmd.Flags().GetBool("json")
			wide, _ := cmd.Flags().GetBool("wide")

			if jsonOutput {
				data, err := stats.FormatJSON(snapshot)
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			fmt.Fprint(cmd.OutOrStdout(), stats.FormatTable(snapshot, wide))
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "Output JSON instead of table")
	cmd.Flags().Bool("wide", false, "Include session IDs alongside names")
	return cmd
}
