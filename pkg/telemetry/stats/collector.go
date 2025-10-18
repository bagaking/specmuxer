package stats

// Collector aggregates status metrics.
type Collector struct{}

// NewCollector builds a Collector with default configuration.
func NewCollector() *Collector {
	return &Collector{}
}
