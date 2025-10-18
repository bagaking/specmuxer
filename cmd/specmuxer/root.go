package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// GlobalOptions holds persistent CLI flags shared across all commands.
type GlobalOptions struct {
	Workspace string
	Config    string
	Verbose   bool
}

var globalOpts = &GlobalOptions{}

var rootCmd = &cobra.Command{
	Use:           "specmuxer",
	Short:         "SpecMuxer orchestrates tmux-based AI tool sessions",
	Long:          `SpecMuxer orchestrates tmux-backed AI coding tool sessions with durable storage and recovery.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		return prepareWorkspace()
	},
}

// Execute runs the root command and exits with a non-zero code on failure.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if rootCmd.SilenceErrors {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

// prepareWorkspace resolves workspace and config paths before command execution.
func prepareWorkspace() error {
	if globalOpts.Workspace == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("determine working directory: %w", err)
		}
		globalOpts.Workspace = cwd
	} else {
		abs, err := filepath.Abs(globalOpts.Workspace)
		if err != nil {
			return fmt.Errorf("resolve workspace %q: %w", globalOpts.Workspace, err)
		}
		globalOpts.Workspace = abs
	}

	if globalOpts.Config != "" {
		abs, err := filepath.Abs(globalOpts.Config)
		if err != nil {
			return fmt.Errorf("resolve config %q: %w", globalOpts.Config, err)
		}
		globalOpts.Config = abs
	}

	return nil
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&globalOpts.Workspace, "workspace", "w", "", "Workspace root containing .specmuxer/")
	rootCmd.PersistentFlags().StringVar(&globalOpts.Config, "config", "", "Override path to specmuxer configuration file")
	rootCmd.PersistentFlags().BoolVarP(&globalOpts.Verbose, "verbose", "v", false, "Enable verbose diagnostic logging")
}
