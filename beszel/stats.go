package beszel

// Stats is a minute of a host's measures, from the system_stats
// collection, or the average of a longer record: sizes in GB.
type Stats struct {
	CPU        float64 `json:"cpu"`
	Memory     float64 `json:"m"`
	MemoryUsed float64 `json:"mu"`
	Disk       float64 `json:"d"`
	DiskUsed   float64 `json:"du"`
	// NetSent and NetRecv are the traffic in MB/s an older agent sends;
	// Bandwidth is the same in bytes/s, sent and received.
	NetSent   float64 `json:"ns,omitempty"`
	NetRecv   float64 `json:"nr,omitempty"`
	Bandwidth Pair    `json:"b,omitzero"`
	// ExtraFs are the filesystems the agent watches beside the root, by
	// the name the agent sets for each.
	ExtraFs map[string]FsStats `json:"efs,omitempty"`
}

// Net is the traffic in bytes/s, sent and received.
func (s Stats) Net() (sent, recv float64) {
	if s.Bandwidth != (Pair{}) {
		return s.Bandwidth[0], s.Bandwidth[1]
	}
	return s.NetSent * mb, s.NetRecv * mb
}

// mb is a megabyte as Beszel counts one, in bytes.
const mb = 1024 * 1024
