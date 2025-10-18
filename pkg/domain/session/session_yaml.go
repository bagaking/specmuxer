package session

// SetUserKilled ensures the user_killed flag is persisted with pointer semantics.
func SetUserKilled(record *SessionRecord, value bool) {
	if record == nil {
		return
	}
	record.UserKilled = boolPtr(value)
}

// boolPtr returns a pointer to the provided boolean value.
func boolPtr(v bool) *bool {
	return &v
}

// IsUserKilled returns true when the session has been explicitly marked as user killed.
func IsUserKilled(record SessionRecord) bool {
	return record.UserKilled != nil && *record.UserKilled
}
