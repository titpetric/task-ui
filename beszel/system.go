package beszel

import (
	"sort"
	"time"
)

// System is a host Beszel monitors, as its hub reports it.
type System struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Host   string `json:"host"`
	Port   string `json:"port"`
	Status string `json:"status"`
	Info   Info   `json:"info"`
	// Updated is when the hub last heard from the host's agent.
	Updated string `json:"updated"`
	// Stats are the host's latest minute of measures, with the sizes the
	// systems list leaves out; nil when the hub has none.
	Stats *Stats `json:"stats,omitempty"`
}

// CPUs is how many cores the host has, or threads when the agent sends
// no core count.
func (s System) CPUs() int {
	if s.Info.Cores > 0 {
		return s.Info.Cores
	}
	return s.Info.Threads
}

// Up reports whether the hub hears from the host.
func (s System) Up() bool {
	return s.Status == "up"
}

// UptimeDuration is how long the host has run.
func (s System) UptimeDuration() time.Duration {
	return time.Duration(s.Info.Uptime) * time.Second
}

// ByCapacity returns the systems largest first: by cores, then by RAM,
// then by name, in a new slice.
func ByCapacity(systems []System) []System {
	sorted := append([]System(nil), systems...)
	sort.SliceStable(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]
		if a.CPUs() != b.CPUs() {
			return a.CPUs() > b.CPUs()
		}
		if a.memory() != b.memory() {
			return a.memory() > b.memory()
		}
		return a.Name < b.Name
	})
	return sorted
}

// memory is the host's RAM in GB, 0 without stats.
func (s System) memory() float64 {
	if s.Stats == nil {
		return 0
	}
	return s.Stats.Memory
}
