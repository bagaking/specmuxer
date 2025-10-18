package stats

// CalculateResumeSuccessRate returns the ratio of successful resumes to attempts.
func CalculateResumeSuccessRate(attempts, successes int) float64 {
	if attempts == 0 {
		return 0
	}
	if successes < 0 {
		successes = 0
	}
	if successes > attempts {
		successes = attempts
	}
	return float64(successes) / float64(attempts)
}
