package stats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/bagaking/specmuxer/pkg/domain/session"
)

// JSONStatus is the serialized representation of a Snapshot.
type JSONStatus struct {
	CollectedAt time.Time     `json:"collectedAt"`
	Totals      Totals        `json:"totals"`
	Projects    []JSONProject `json:"projects"`
	Top         []JSONSession `json:"top,omitempty"`
}

// JSONProject summarizes sessions in a project.
type JSONProject struct {
	ProjectID string        `json:"projectId"`
	Sessions  []JSONSession `json:"sessions"`
}

// JSONSession summarizes an individual session.
type JSONSession struct {
	ID           string                `json:"id"`
	Tool         string                `json:"tool"`
	HumanName    string                `json:"humanName,omitempty"`
	Status       session.SessionStatus `json:"status"`
	LastOutputAt string                `json:"lastOutputAt,omitempty"`
	Idle         bool                  `json:"idle"`
	UserKilled   bool                  `json:"userKilled"`
}

// FormatJSON converts a snapshot to JSON representation.
func FormatJSON(snapshot Snapshot) ([]byte, error) {
	j := JSONStatus{
		CollectedAt: snapshot.CollectedAt,
		Totals:      snapshot.Totals,
		Projects:    make([]JSONProject, 0, len(snapshot.Projects)),
		Top:         convertSessions(snapshot.Top),
	}
	for _, project := range snapshot.Projects {
		j.Projects = append(j.Projects, JSONProject{
			ProjectID: project.ProjectID,
			Sessions:  convertSessions(project.Sessions),
		})
	}
	return json.MarshalIndent(j, "", "  ")
}

// FormatTable formats the snapshot as a tabular string.
func FormatTable(snapshot Snapshot, wide bool) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 4, 2, ' ', 0)

	fmt.Fprintf(w, "Collected:\t%s\n", snapshot.CollectedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "Totals:\tactive %d\tidle %d\tstopped %d\tfailed %d\n", snapshot.Totals.Active, snapshot.Totals.Idle, snapshot.Totals.Stopped, snapshot.Totals.Failed)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "PROJECT\tSESSION\tTOOL\tSTATUS\tLAST OUTPUT\tIDLE\tUSER KILLED")
	for _, project := range snapshot.Projects {
		for _, sess := range project.Sessions {
			last := "-"
			if sess.LastOutputAt != nil {
				last = sess.LastOutputAt.Format(time.RFC3339)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%t\t%t\n",
				project.ProjectID,
				displaySessionName(sess, wide),
				sess.Tool,
				sess.Status,
				last,
				sess.Idle,
				sess.UserKilled,
			)
		}
	}

	_ = w.Flush()
	return buf.String()
}

func convertSessions(sessions []SessionSummary) []JSONSession {
	result := make([]JSONSession, 0, len(sessions))
	for _, sess := range sessions {
		js := JSONSession{
			ID:         sess.ID,
			Tool:       sess.Tool,
			HumanName:  sess.HumanName,
			Status:     sess.Status,
			Idle:       sess.Idle,
			UserKilled: sess.UserKilled,
		}
		if sess.LastOutputAt != nil {
			js.LastOutputAt = sess.LastOutputAt.Format(time.RFC3339)
		}
		result = append(result, js)
	}
	return result
}

func displaySessionName(sess SessionSummary, wide bool) string {
	if wide && sess.HumanName != "" {
		return fmt.Sprintf("%s (%s)", sess.HumanName, sess.ID)
	}
	if sess.HumanName != "" {
		return sess.HumanName
	}
	return sess.ID
}
