package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bagaking/specmuxer/pkg/orchestrator/run"
	"github.com/bagaking/specmuxer/pkg/runtime/tmux"
	"github.com/bagaking/specmuxer/pkg/ui/statusbar"
)

func init() {
	rootCmd.AddCommand(newRunCommand())
}

func newRunCommand() *cobra.Command {
	var detach bool
	var force bool
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

			existing, err := deps.runSvc.Sessions()
			if err != nil {
				return err
			}

			active := filterActiveSessions(existing)
			if len(active) > 0 && !force {
				out := cmd.OutOrStdout()
				if !hasInteractiveTerminal() {
					fmt.Fprintln(out, "Active sessions detected; skipping launch in non-interactive mode.")
					for _, rec := range active {
						fmt.Fprintf(out, "  - %s (%s, status=%s)\n", rec.ID, rec.Tool, rec.Status)
					}
					fmt.Fprintln(out, "Re-run with --force to create another session without confirmation.")
					return nil
				}

				fmt.Fprintln(out, "Active sessions detected:")
				for _, rec := range active {
					fmt.Fprintf(out, "  - %s (%s, status=%s)\n", rec.ID, rec.Tool, rec.Status)
				}
				proceed, err := promptYesNo(cmd.InOrStdin(), out, "Create another session? [y/N]: ")
				if err != nil {
					return err
				}
				if !proceed {
					fmt.Fprintln(out, "Aborted. No new session launched.")
					return nil
				}
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
			deps.collector.SetLiveness(nil)
			snapshot := deps.collector.BuildSnapshot(records)
			renderer := statusbar.NewRenderer()

			fmt.Fprintf(cmd.OutOrStdout(), "Session %s launched\n", sessionID)
			fmt.Fprintln(cmd.OutOrStdout(), renderer.Render(snapshot))

			if !detach {
				record, ok := findSession(records, sessionID)
				if ok {
					if !hasInteractiveTerminal() {
						fmt.Fprintf(cmd.OutOrStdout(), "Skipping tmux attach for session %s: no interactive terminal detected.\n", sessionID)
						fmt.Fprintf(cmd.OutOrStdout(), "  tmux session: %s\n  socket: %s\n", record.Tmux.Session, record.Tmux.Socket)
						fmt.Fprintf(cmd.OutOrStdout(), "Run 'specmuxer attach %s' from a terminal to connect.\n", sessionID)
						return nil
					}
					if err := deps.tmux.Attach(ctx, tmux.AttachOptions{Session: record.Tmux.Session, Socket: record.Tmux.Socket}); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
	cmd.Flags().StringArrayP("env", "e", nil, "Set environment variable (KEY=VALUE)")
	cmd.Flags().String("name", "", "Optional human-friendly session name")
	cmd.Flags().BoolVar(&detach, "detach", false, "Run without attaching to the tmux session")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation when sessions are already running")
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
