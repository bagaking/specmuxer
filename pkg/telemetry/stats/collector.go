package stats

import (
	"sort"
	"time"

	"github.com/bagaking/specmuxer/pkg/domain/session"
)

const defaultIdleThreshold = 2 * time.Minute

// Collector aggregates telemetry for status/top commands.
type Collector struct {
	idleThreshold time.Duration
	now           func() time.Time
	topLimit      int
	projectPaths  map[string]string
	liveness      map[string]bool
}

// Option configures a Collector.
type Option func(*Collector)

// WithIdleThreshold overrides the inactivity duration that marks a session idle.
func WithIdleThreshold(threshold time.Duration) Option {
	return func(c *Collector) {
		if threshold > 0 {
			c.idleThreshold = threshold
		}
	}
}

// WithClock injects a custom time source.
func WithClock(clock func() time.Time) Option {
	return func(c *Collector) {
		if clock != nil {
			c.now = clock
		}
	}
}

// WithTopLimit sets the number of sessions tracked in the Top slice.
func WithTopLimit(limit int) Option {
	return func(c *Collector) {
		if limit > 0 {
			c.topLimit = limit
		}
	}
}

// WithProjectPaths supplies workspace paths keyed by project ID.
func WithProjectPaths(paths map[string]string) Option {
	return func(c *Collector) {
		if paths == nil {
			return
		}
		c.projectPaths = make(map[string]string, len(paths))
		for k, v := range paths {
			c.projectPaths[k] = v
		}
	}
}

// SetLiveness injects tmux liveness information keyed by session ID.
func (c *Collector) SetLiveness(live map[string]bool) {
	if live == nil {
		c.liveness = map[string]bool{}
		return
	}
	c.liveness = make(map[string]bool, len(live))
	for k, v := range live {
		c.liveness[k] = v
	}
}

// NewCollector constructs a Collector with optional configuration.
func NewCollector(opts ...Option) *Collector {
	c := &Collector{
		idleThreshold: defaultIdleThreshold,
		now:           time.Now,
		topLimit:      10,
		projectPaths:  map[string]string{},
		liveness:      map[string]bool{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Snapshot represents aggregated telemetry at a point in time.
type Snapshot struct {
	CollectedAt time.Time
	Totals      Totals
	Projects    []ProjectSummary
	Top         []SessionSummary
}

// Totals captures high-level counts across all sessions.
type Totals struct {
	Active  int
	Idle    int
	Stopped int
	Failed  int
}

// ProjectSummary groups sessions by project.
type ProjectSummary struct {
	ProjectID string
	RootPath  string
	Sessions  []SessionSummary
}

// SessionSummary provides per-session telemetry.
type SessionSummary struct {
	ID           string
	ProjectID    string
	Tool         string
	HumanName    string
	Status       session.SessionStatus
	LastOutputAt *time.Time
	Idle         bool
	UserKilled   bool
	TmuxAlive    bool
}

// BuildSnapshot aggregates telemetry for the provided sessions.
func (c *Collector) BuildSnapshot(records []session.SessionRecord) Snapshot {
	now := c.now()
	projectBuckets := make(map[string][]SessionSummary)
	var (
		totalActive  int
		totalIdle    int
		totalStopped int
		totalFailed  int
		topSessions  []SessionSummary
	)

	for _, record := range records {
		summary := SessionSummary{
			ID:           record.ID,
			ProjectID:    record.ProjectID,
			Tool:         record.Tool,
			HumanName:    record.HumanName,
			Status:       record.Status,
			LastOutputAt: record.LastOutputAt,
			UserKilled:   record.UserKilled != nil && *record.UserKilled,
		}

		summary.Idle = c.isIdle(now, record)
		summary.TmuxAlive = c.isAlive(record.ID)

		statusForTotals := summary.Status
		if !summary.TmuxAlive && summary.Status == session.StatusRunning {
			summary.Status = session.StatusStopped
			statusForTotals = session.StatusStopped
		}

		switch statusForTotals {
		case session.StatusRunning:
			if summary.Idle {
				totalIdle++
			} else {
				totalActive++
			}
		case session.StatusIdle:
			totalIdle++
		case session.StatusStopped:
			totalStopped++
		case session.StatusFailed:
			totalFailed++
		}

		projectBuckets[record.ProjectID] = append(projectBuckets[record.ProjectID], summary)
		topSessions = append(topSessions, summary)
	}

	projectSummaries := make([]ProjectSummary, 0, len(projectBuckets))
	for projectID, sessions := range projectBuckets {
		sortSessionsByLastOutput(sessions)
		projectSummaries = append(projectSummaries, ProjectSummary{
			ProjectID: projectID,
			RootPath:  c.projectPaths[projectID],
			Sessions:  sessions,
		})
	}
	sort.Slice(projectSummaries, func(i, j int) bool {
		return projectSummaries[i].ProjectID < projectSummaries[j].ProjectID
	})

	sortSessionsByLastOutput(topSessions)
	if len(topSessions) > c.topLimit {
		topSessions = append([]SessionSummary(nil), topSessions[:c.topLimit]...)
	}

	return Snapshot{
		CollectedAt: now,
		Totals: Totals{
			Active:  totalActive,
			Idle:    totalIdle,
			Stopped: totalStopped,
			Failed:  totalFailed,
		},
		Projects: projectSummaries,
		Top:      topSessions,
	}
}

// Top returns sessions sorted by most recent activity (descending).
func (c *Collector) Top(records []session.SessionRecord, limit int) []SessionSummary {
	if limit <= 0 {
		limit = c.topLimit
	}
	summaries := make([]SessionSummary, 0, len(records))
	now := c.now()
	for _, record := range records {
		summary := SessionSummary{
			ID:           record.ID,
			ProjectID:    record.ProjectID,
			Tool:         record.Tool,
			HumanName:    record.HumanName,
			Status:       record.Status,
			LastOutputAt: record.LastOutputAt,
			UserKilled:   record.UserKilled != nil && *record.UserKilled,
			Idle:         c.isIdle(now, record),
			TmuxAlive:    c.isAlive(record.ID),
		}
		if !summary.TmuxAlive && summary.Status == session.StatusRunning {
			summary.Status = session.StatusStopped
		}
		summaries = append(summaries, summary)
	}
	sortSessionsByLastOutput(summaries)
	if len(summaries) > limit {
		return append([]SessionSummary(nil), summaries[:limit]...)
	}
	return summaries
}

func (c *Collector) isIdle(now time.Time, record session.SessionRecord) bool {
	if record.LastOutputAt == nil {
		return record.Status == session.StatusIdle || record.Status == session.StatusStopped
	}
	if c.idleThreshold <= 0 {
		return record.Status == session.StatusIdle
	}
	return now.Sub(*record.LastOutputAt) >= c.idleThreshold
}

func (c *Collector) isAlive(id string) bool {
	if c.liveness == nil {
		return true
	}
	alive, ok := c.liveness[id]
	if !ok {
		return true
	}
	return alive
}

func sortSessionsByLastOutput(sessions []SessionSummary) {
	sort.SliceStable(sessions, func(i, j int) bool {
		ti := time.Time{}
		if sessions[i].LastOutputAt != nil {
			ti = *sessions[i].LastOutputAt
		}
		tj := time.Time{}
		if sessions[j].LastOutputAt != nil {
			tj = *sessions[j].LastOutputAt
		}
		if ti.Equal(tj) {
			return sessions[i].ID < sessions[j].ID
		}
		return ti.After(tj)
	})
}
