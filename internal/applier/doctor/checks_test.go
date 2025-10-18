package doctor

import (
	"os"
	"testing"
)

func TestRunAggregatesChecks(t *testing.T) {
	report := Run(Config{Workspace: os.TempDir()})
	if len(report.Checks) == 0 {
		t.Fatalf("expected checks to run")
	}
	if report.OverallStatus == StatusFail {
		t.Fatalf("expected overall status to be pass or warn")
	}
}

func TestWorkspaceCheckFailure(t *testing.T) {
	report := Run(Config{Workspace: ""})
	foundWarn := false
	for _, check := range report.Checks {
		if check.Name == "workspace" && check.Status == StatusWarn {
			foundWarn = true
		}
	}
	if !foundWarn {
		t.Fatalf("expected workspace warning when workspace unspecified")
	}
}
