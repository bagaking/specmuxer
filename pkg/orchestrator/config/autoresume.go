package config

// AutoResumeEnabled reports whether automatic resume mode is configured.
func (c *Config) AutoResumeEnabled() bool {
	if c == nil {
		return false
	}
	return c.ResumeMode == ResumeModeAutomatic
}

// SetResumeMode updates the resume mode while enforcing defaults.
func (c *Config) SetResumeMode(mode ResumeMode) {
	if c == nil {
		return
	}
	switch mode {
	case ResumeModeManual, ResumeModeAutomatic:
		c.ResumeMode = mode
	default:
		c.ResumeMode = defaultResumeMode
	}
}
