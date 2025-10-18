package doctor

import (
	"os"
	"os/exec"
)

// Status represents the outcome of a doctor check.
type Status string

const (
	StatusPass Status = "pass"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

// Check captures the result of a single verification step.
type Check struct {
	Name        string
	Status      Status
	Message     string
	Remediation string
}

// Report aggregates all doctor checks.
type Report struct {
	Checks        []Check
	OverallStatus Status
}

// Config allows callers to supply additional context for checks.
type Config struct {
	Workspace string
}

// Run executes the default doctor checks.
func Run(cfg Config) Report {
	checks := []Check{
		tmuxCheck(),
		workspaceCheck(cfg.Workspace),
	}
	return Report{
		Checks:        checks,
		OverallStatus: aggregateStatus(checks),
	}
}

func tmuxCheck() Check {
	if _, err := exec.LookPath("tmux"); err != nil {
		return Check{
			Name:        "tmux",
			Status:      StatusFail,
			Message:     "tmux binary not found in PATH",
			Remediation: "Install tmux >= 3.x and ensure it is available on PATH",
		}
	}
	return Check{
		Name:    "tmux",
		Status:  StatusPass,
		Message: "tmux binary reachable",
	}
}

func workspaceCheck(workspace string) Check {
	if workspace == "" {
		return Check{Name: "workspace", Status: StatusWarn, Message: "workspace path not provided"}
	}
	info, err := os.Stat(workspace)
	if err != nil {
		return Check{Name: "workspace", Status: StatusFail, Message: "workspace path not accessible", Remediation: "Verify directory exists and permissions are correct"}
	}
	if !info.IsDir() {
		return Check{Name: "workspace", Status: StatusFail, Message: "workspace path is not a directory"}
	}
	return Check{Name: "workspace", Status: StatusPass, Message: "workspace directory accessible"}
}

func aggregateStatus(checks []Check) Status {
	status := StatusPass
	for _, check := range checks {
		if check.Status == StatusFail {
			return StatusFail
		}
		if check.Status == StatusWarn {
			status = StatusWarn
		}
	}
	return status
}
