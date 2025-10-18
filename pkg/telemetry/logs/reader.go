package logs

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// Reader loads session logs from disk with optional redaction rules.
type Reader struct {
	rootDir   string
	redactors []*regexp.Regexp
}

// NewReader constructs a Reader rooted at the given directory.
func NewReader(root string, redactionPatterns ...string) *Reader {
	var compiled []*regexp.Regexp
	for _, pattern := range redactionPatterns {
		if pattern == "" {
			continue
		}
		re, err := regexp.Compile(pattern)
		if err == nil {
			compiled = append(compiled, re)
		}
	}
	return &Reader{
		rootDir:   root,
		redactors: compiled,
	}
}

// LastLines returns up to N most recent log lines for the session.
func (r *Reader) LastLines(sessionID string, limit int) ([]string, error) {
	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}
	sanitized := SanitizeSessionID(sessionID)
	pattern := filepath.Join(r.rootDir, sanitized+"*.log")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no logs found for session %s", sessionID)
	}
	path := matches[len(matches)-1]
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open log %q: %w", path, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, r.applyRedaction(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	return lines, nil
}

func (r *Reader) applyRedaction(line string) string {
	redacted := line
	for _, re := range r.redactors {
		redacted = re.ReplaceAllString(redacted, redactionReplacement)
	}
	return redacted
}
