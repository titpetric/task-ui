package beszel

import (
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
}

// Info is what a host's agent last sent: the machine, and its load.
type Info struct {
	Hostname string  `json:"h"`
	Kernel   string  `json:"k"`
	Cores    int     `json:"c"`
	Threads  int     `json:"t"`
	CPUModel string  `json:"m"`
	Uptime   float64 `json:"u"`
	CPU      float64 `json:"cpu"`
	Memory   float64 `json:"mp"`
	Disk     float64 `json:"dp"`
	// Bandwidth is the network traffic, in MB/s.
	Bandwidth float64 `json:"b"`
	Agent     string  `json:"v"`
}

// Up reports whether the hub hears from the host.
func (s System) Up() bool {
	return s.Status == "up"
}

// UptimeDuration is how long the host has run.
func (s System) UptimeDuration() time.Duration {
	return time.Duration(s.Info.Uptime) * time.Second
}
