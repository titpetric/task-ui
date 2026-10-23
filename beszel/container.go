package beszel

// Container is a container on a host, as the hub reports it: CPU in
// percent of the whole host, memory in MB, the traffic in bytes/s.
// The containers collection carries the image and the status; a record
// of container_stats only the measures.
type Container struct {
	Name   string  `json:"name"`
	Image  string  `json:"image,omitempty"`
	Status string  `json:"status,omitempty"`
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	Net    float64 `json:"net,omitempty"`
}

// containerStat is a container in a record of container_stats.
type containerStat struct {
	Name      string  `json:"n"`
	CPU       float64 `json:"c"`
	Memory    float64 `json:"m"`
	NetSent   float64 `json:"ns"`
	NetRecv   float64 `json:"nr"`
	Bandwidth Pair    `json:"b"`
}

// container is the stat as a Container.
func (c containerStat) container() Container {
	net := c.Bandwidth[0] + c.Bandwidth[1]
	if net == 0 {
		net = (c.NetSent + c.NetRecv) * mb
	}
	return Container{Name: c.Name, CPU: c.CPU, Memory: c.Memory, Net: net}
}
