package main

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/bagaking/specmuxer/pkg/runtime/tmux"
)

func init() {
	rootCmd.AddCommand(newAttachCommand())
}

func newAttachCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach [session]",
		Short: "Attach to a running tmux session",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deps, err := loadRuntime()
			if err != nil {
				return err
			}

			target := strings.TrimSpace(args[0])
			if target == "" {
				return nil
			}
			return deps.tmux.Attach(cmd.Context(), tmux.AttachOptions{
				Session: target,
				Socket:  deps.runSvc.SocketPath(),
			})
		},
	}
	return cmd
}
