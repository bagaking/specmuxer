package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/bagaking/specmuxer/pkg/domain/session"
	"github.com/bagaking/specmuxer/pkg/telemetry/stats"
)

func main() {
	iterations := flag.Int("iterations", 1000, "Number of collector runs to execute")
	flag.Parse()

	records := generateSessions(50)
	collector := stats.NewCollector()

	start := time.Now()
	for i := 0; i < *iterations; i++ {
		collector.BuildSnapshot(records)
	}
	duration := time.Since(start)

	fmt.Printf("iterations=%d duration=%s avg=%s\n", *iterations, duration, duration/time.Duration(*iterations))
}

func generateSessions(count int) []session.SessionRecord {
	records := make([]session.SessionRecord, 0, count)
	for i := 0; i < count; i++ {
		last := time.Now().Add(-time.Duration(rand.Intn(300)) * time.Second)
		records = append(records, session.SessionRecord{
			ID:        fmt.Sprintf("sess-%d", i),
			ProjectID: "benchmark",
			Tool:      "codex",
			Status:    session.StatusRunning,
			LastOutputAt: func() *time.Time {
				v := last
				return &v
			}(),
		})
	}
	return records
}
