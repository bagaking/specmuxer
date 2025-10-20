package resume

import (
	"strings"

	"github.com/bagaking/specmuxer/pkg/adapters"
	"github.com/bagaking/specmuxer/pkg/domain/session"
)

// Eligible reports whether a session should be considered for resume.
func Eligible(record session.SessionRecord, def adapters.Definition) bool {
	eligible, _ := EvaluateEligibility(record, def, true)
	return eligible
}

// EvaluateEligibility returns whether the session can be resumed and, when not, the reason.
func EvaluateEligibility(record session.SessionRecord, def adapters.Definition, tmuxAlive bool) (bool, string) {
	if record.UserKilled != nil && *record.UserKilled {
		return false, "user_killed flag set"
	}

	if tmuxAlive && record.Status == session.StatusRunning {
		return false, "session already running"
	}

	if def.Resume != nil && len(def.Resume.Exec) > 0 {
		return true, ""
	}

	if len(def.Start.Exec) > 0 {
		return true, ""
	}

	return false, "adapter missing resume/start command"
}

func adapterKey(name string) string {
	return strings.ToLower(name)
}
