package resume

import (
	"strings"

	"github.com/bagaking/specmuxer/pkg/adapters"
	"github.com/bagaking/specmuxer/pkg/domain/session"
)

// Eligible reports whether a session should be considered for resume.
func Eligible(record session.SessionRecord, def adapters.Definition) bool {
	if record.UserKilled != nil && *record.UserKilled {
		return false
	}

	if record.Status == session.StatusRunning {
		return false
	}

	if def.Resume != nil && len(def.Resume.Exec) > 0 {
		return true
	}

	// Fallback to start command when resume command absent.
	return len(def.Start.Exec) > 0
}

func adapterKey(name string) string {
	return strings.ToLower(name)
}
