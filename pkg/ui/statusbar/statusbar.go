package statusbar

import (
	"fmt"
	"strings"
	"time"

	"github.com/bagaking/specmuxer/pkg/telemetry/stats"
)

// Renderer produces status bar summaries for SpecMuxer sessions.
type Renderer struct{}

// NewRenderer constructs a Renderer.
func NewRenderer() *Renderer {
	return &Renderer{}
}

// Render returns a concise status segment summarizing session activity.
func (r *Renderer) Render(snapshot stats.Snapshot) string {
	segments := []string{
		fmt.Sprintf("Active:%d", snapshot.Totals.Active),
		fmt.Sprintf("Idle:%d", snapshot.Totals.Idle),
		fmt.Sprintf("Stopped:%d", snapshot.Totals.Stopped),
		fmt.Sprintf("Failed:%d", snapshot.Totals.Failed),
	}

	var mostRecent time.Time
	if len(snapshot.Top) > 0 && snapshot.Top[0].LastOutputAt != nil {
		mostRecent = *snapshot.Top[0].LastOutputAt
	}
	if !mostRecent.IsZero() {
		segments = append(segments, fmt.Sprintf("Last:%s", mostRecent.Format(time.Kitchen)))
	}

	return strings.Join(segments, " | ")
}
