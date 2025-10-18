package main

import (
	"fmt"

	"github.com/spf13/cobra"

	doctor "github.com/bagaking/specmuxer/internal/applier/doctor"
)

func init() {
	rootCmd.AddCommand(newDoctorCommand())
}

func newDoctorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run environment diagnostics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			deps, err := loadRuntime()
			if err != nil {
				return err
			}
			report := doctor.Run(doctor.Config{Workspace: deps.config.WorkspaceRoot})
			for _, check := range report.Checks {
				fmt.Fprintf(cmd.OutOrStdout(), "%s	%s	%s\n", check.Name, check.Status, check.Message)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Overall: %s\n", report.OverallStatus)
			return nil
		},
	}
	return cmd
}
