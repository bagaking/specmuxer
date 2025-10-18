package sidebar

import (
	"fmt"
	"strings"
)

// RenderGCPrompt builds an adopt/ignore prompt for orphan sessions.
func RenderGCPrompt(orphans []string) string {
	if len(orphans) == 0 {
		return "No orphaned sessions detected."
	}
	return fmt.Sprintf("Orphaned sessions: %s\n[A] Adopt records\n[I] Ignore and exit", strings.Join(orphans, ", "))
}
