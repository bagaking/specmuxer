package logs

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestAppendAppliesRedactionAndPermissions(t *testing.T) {
	temp := t.TempDir()
	manager, err := NewManager(temp, WithRedactionRules([]string{`sk-[A-Za-z0-9]+`}))
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	path, err := manager.Append("specmuxer:p-1", "token sk-abcdefg")
	if err != nil {
		t.Fatalf("Append: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if got := string(content); !strings.Contains(got, redactionReplacement) {
		t.Fatalf("expected redacted token, got %q", got)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != defaultFilePerm {
		t.Fatalf("unexpected file permission %o", info.Mode().Perm())
	}
}

func TestAppendRotatesWhenExceedingMaxBytes(t *testing.T) {
	temp := t.TempDir()
	manager, err := NewManager(temp, WithMaxBytes(16))
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	path1, err := manager.Append("session-1", "alpha")
	if err != nil {
		t.Fatalf("Append alpha: %v", err)
	}
	path2, err := manager.Append("session-1", "abcdefghijklmnopqrstuvwxyz")
	if err != nil {
		t.Fatalf("Append beta: %v", err)
	}

	if path1 == path2 {
		t.Fatalf("expected rotation to create new file, both writes targeted %q", path1)
	}

	entries, err := os.ReadDir(temp)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 log files, found %d", len(entries))
	}
}

func TestPruneRemovesExpiredSegments(t *testing.T) {
	temp := t.TempDir()
	fixed := time.Date(2025, 10, 18, 12, 0, 0, 0, time.UTC)
	manager, err := NewManager(temp, WithRetention(time.Hour), WithClock(func() time.Time { return fixed }))
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	path, err := manager.Append("session", "hello")
	if err != nil {
		t.Fatalf("Append: %v", err)
	}

	if err := os.Chtimes(path, fixed.Add(-2*time.Hour), fixed.Add(-2*time.Hour)); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}

	if err := manager.Prune(); err != nil {
		t.Fatalf("Prune: %v", err)
	}

	entries, err := os.ReadDir(temp)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 log files after prune, found %d", len(entries))
	}
}
