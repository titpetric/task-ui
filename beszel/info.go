package beszel

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
