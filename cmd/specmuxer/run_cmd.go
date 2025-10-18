package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bagaking/specmuxer/pkg/orchestrator/run"
	"github.com/bagaking/specmuxer/pkg/ui/statusbar"
)

func init() {
	rootCmd.AddCommand(newRunCommand())
}

func newRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [adapter] [-- adapter-args]",
		Short: "Launch an AI tool session",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			args := cmd.Flags().Args()
			if len(args) == 0 {
				return errors.New("adapter argument required")
			}

			adapterArgs := args
			var passthrough []string
			if dash := cmd.ArgsLenAtDash(); dash >= 0 {
				adapterArgs = args[:dash]
				passthrough = args[dash+1:]
			}

			adapter := adapterArgs[0]

			name, _ := cmd.Flags().GetString("name")
			envPairs, _ := cmd.Flags().GetStringArray("env")
			envMap, err := parseEnvPairs(envPairs)
			if err != nil {
				return err
			}

			deps, err := loadRuntime()
			if err != nil {
				return err
			}

			req := run.LaunchRequest{
				Adapter:   adapter,
				HumanName: name,
				Env:       envMap,
				Args:      passthrough,
			}

			ctx := cmd.Context()
			if ctx == context.Background() {
				ctx = context.TODO()
			}

			sessionID, err := deps.runSvc.Launch(ctx, req)
			if err != nil {
				return err
			}

			records, err := deps.runSvc.Sessions()
			if err != nil {
				return err
			}
			snapshot := deps.collector.BuildSnapshot(records)
			renderer := statusbar.NewRenderer()

			fmt.Fprintf(cmd.OutOrStdout(), "Session %s launched\n", sessionID)
			fmt.Fprintln(cmd.OutOrStdout(), renderer.Render(snapshot))
			return nil
		},
	}
	cmd.Flags().StringArrayP("env", "e", nil, "Set environment variable (KEY=VALUE)")
	cmd.Flags().String("name", "", "Optional human-friendly session name")
	return cmd
}

func parseEnvPairs(pairs []string) (map[string]string, error) {
	if len(pairs) == 0 {
		return map[string]string{}, nil
	}
	result := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		segments := strings.SplitN(pair, "=", 2)
		if len(segments) != 2 {
			return nil, fmt.Errorf("invalid env pair %q (expected KEY=VALUE)", pair)
		}
		result[segments[0]] = segments[1]
	}
	return result, nil
}
