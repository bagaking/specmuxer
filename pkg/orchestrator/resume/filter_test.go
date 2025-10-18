package resume

import (
	"testing"
	"time"

	"github.com/bagaking/specmuxer/pkg/adapters"
	"github.com/bagaking/specmuxer/pkg/domain/session"
)

func TestEligibleForResume(t *testing.T) {
	falsePtr := func() *bool { v := false; return &v }
	truePtr := func() *bool { v := true; return &v }

	def := adapters.Definition{
		Name:   "codex",
		Start:  adapters.LifecycleCommand{Exec: []string{"sh", "-c", "echo"}},
		Resume: &adapters.LifecycleCommand{Exec: []string{"resume"}},
	}

	cases := []struct {
		name     string
		record   session.SessionRecord
		expected bool
	}{
		{
			name: "running session is already eligible",
			record: session.SessionRecord{
				Status:     session.StatusStopped,
				UserKilled: falsePtr(),
			},
			expected: true,
		},
		{
			name: "user killed session skipped",
			record: session.SessionRecord{
				Status:     session.StatusStopped,
				UserKilled: truePtr(),
			},
			expected: false,
		},
		{
			name:     "resume command missing uses start fallback",
			record:   session.SessionRecord{Status: session.StatusFailed, UserKilled: falsePtr()},
			expected: true,
		},
		{
			name: "idle threshold not exceeded still eligible",
			record: session.SessionRecord{
				Status:       session.StatusStopped,
				LastOutputAt: ptrTime(time.Now().Add(-10 * time.Minute)),
				UserKilled:   falsePtr(),
			},
			expected: true,
		},
	}

	for _, tc := range cases {
		if got := Eligible(tc.record, def); got != tc.expected {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.expected, got)
		}
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
