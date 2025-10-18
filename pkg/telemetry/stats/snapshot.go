package stats

import "github.com/bagaking/specmuxer/pkg/runtime/storage"

// ToStorageStats converts a telemetry snapshot to the storage Stats format.
func ToStorageStats(snapshot Snapshot, resumeRate float64) *storage.Stats {
	return &storage.Stats{
		CollectedAt:       snapshot.CollectedAt,
		ActiveSessions:    snapshot.Totals.Active,
		IdleSessions:      snapshot.Totals.Idle,
		ResumeSuccessRate: resumeRate,
	}
}
