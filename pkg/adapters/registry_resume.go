package adapters

// ResumeCommand returns the lifecycle command to execute when resuming a session.
func (d Definition) ResumeCommand() LifecycleCommand {
	if d.Resume != nil && len(d.Resume.Exec) > 0 {
		cmd := d.Resume.Clone()
		if len(cmd.Exec) == 0 {
			cmd.Exec = append([]string(nil), d.Start.Exec...)
		}
		return cmd
	}
	return d.Start.Clone()
}

// ExtractStateCommand returns the optional extract_state lifecycle command.
func (d Definition) ExtractStateCommand() (LifecycleCommand, bool) {
	if d.ExtractState == nil || len(d.ExtractState.Exec) == 0 {
		return LifecycleCommand{}, false
	}
	cmd := d.ExtractState.Clone()
	return cmd, true
}
