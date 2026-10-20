package view

import (
	"time"
)

// ActivityStats are the finished runs of a chart's window: how many, the
// fastest, the mean and the slowest, and their time together.
type ActivityStats struct {
	Count   int
	Fastest time.Duration
	Mean    time.Duration
	Slowest time.Duration
	Spent   time.Duration
}

// add counts a finished run's duration in.
func (s *ActivityStats) add(elapsed time.Duration) {
	if s.Count == 0 || elapsed < s.Fastest {
		s.Fastest = elapsed
	}
	s.Slowest = max(s.Slowest, elapsed)
	s.Spent += elapsed
	s.Count++
	s.Mean = s.Spent / time.Duration(s.Count)
}
