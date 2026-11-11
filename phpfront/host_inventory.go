package phpfront

// Inventory is the hosts page's facts.
//
// Off is an instance with no Beszel integration, which the page says instead
// of listing nothing, and Error what the hub answered when it answered badly.
type Inventory struct {
	Off    bool     `json:"off"`
	Error  string   `json:"error"`
	Hosts  []System `json:"hosts"`
	Logins []Login  `json:"logins"`

	// Cluster is the hosts summed into pools, from hostinfo.
	Cluster Cluster `json:"cluster"`

	// History is Beszel's records of each host, by host ID, which the
	// cluster's charts draw.
	History []Series `json:"history"`
}

// Login is the ssh login task-ui signs in to a host as.
type Login struct {
	Host  string `json:"host"`
	Login string `json:"login"`
}

// System is one host as Beszel reports it: the shares its line meters, and the
// sizes behind the memory and the disk.
//
// Measured is false where the hub holds no minute of measures, and then every
// size is zero rather than a reading.
type System struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Host   string `json:"host"`
	Status string `json:"status"`
	Up     bool   `json:"up"`
	CPUs   int    `json:"cpus"`
	Agent  string `json:"agent"`

	// Uptime is how long the host has run, as the host's page words it.
	Uptime string `json:"uptime"`

	// CPU, Memory and Disk are shares, 0 to 100.
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	Disk   float64 `json:"disk"`

	Measured    bool    `json:"measured"`
	MemoryTotal float64 `json:"memory_total"`
	MemoryUsed  float64 `json:"memory_used"`
	DiskTotal   float64 `json:"disk_total"`
	DiskUsed    float64 `json:"disk_used"`
}

// Cluster is the hosts summed into pools, as hostinfo computes it.
type Cluster struct {
	Up    int           `json:"up"`
	Down  int           `json:"down"`
	Left  int           `json:"left"`
	Pools []Pool        `json:"pools"`
	Hosts []ClusterHost `json:"hosts"`
}

// Pool is one resource the hosts share: the cores, the memory, the disks and
// the running containers.
type Pool struct {
	Name     string    `json:"name"`
	Unit     string    `json:"unit"`
	Figure   string    `json:"figure"`
	Used     float64   `json:"used"`
	Total    float64   `json:"total"`
	Percent  float64   `json:"percent"`
	Segments []Segment `json:"segments"`
}

// Segment is one host's share of a pool, as the pool's bar draws it.
type Segment struct {
	Host  string  `json:"host"`
	Color int     `json:"color"`
	Width float64 `json:"width"`
	Fill  float64 `json:"fill"`
	Title string  `json:"title"`
}

// ClusterHost is a host of the pools' legend. Left is a host the pools leave
// out, which is down or sends no measures.
type ClusterHost struct {
	Name  string `json:"name"`
	Color int    `json:"color"`
	Left  bool   `json:"left"`
}

// Series is Beszel's records of one host.
type Series struct {
	ID      string   `json:"id"`
	Samples []Sample `json:"samples"`
}

// Sample is one record of a host: the minute Beszel kept, as the charts read
// it. Sent and Recv are bytes a second.
type Sample struct {
	At         string  `json:"at"`
	CPU        float64 `json:"cpu"`
	Memory     float64 `json:"memory"`
	MemoryUsed float64 `json:"memory_used"`
	Disk       float64 `json:"disk"`
	DiskUsed   float64 `json:"disk_used"`
	Sent       float64 `json:"sent"`
	Recv       float64 `json:"recv"`

	// Extra are the filesystems the agent watches beside the root.
	Extra []Mount `json:"extra"`
}

// Mount is a filesystem of a record, by the name the agent set for it.
type Mount struct {
	Name  string  `json:"name"`
	Used  float64 `json:"used"`
	Total float64 `json:"total"`
}

// HostFacts is one host of the inventory in full.
//
// Found is false where Beszel monitors no such host, and Error what the hub
// answered when it refused.
type HostFacts struct {
	Found  bool   `json:"found"`
	Error  string `json:"error"`
	System System `json:"system"`
	Login  string `json:"login"`

	// Containers are the host's, busiest first. Measured says Beszel
	// reported them and Sized that docker ps answered over ssh; Unsized
	// says why it did not.
	Containers []Container `json:"containers"`
	Measured   bool        `json:"measured"`
	Sized      bool        `json:"sized"`
	Unsized    string      `json:"unsized"`

	// Filesystems are the durable ones; FromDF says they come from df over
	// ssh rather than from Beszel.
	Filesystems []Filesystem `json:"filesystems"`
	FromDF      bool         `json:"from_df"`

	// Read is when the host last answered over ssh, empty for never.
	Read string `json:"read"`

	// Samples are Beszel's records of the host, and Containers' of its
	// containers, which the charts draw.
	Samples          []Sample          `json:"samples"`
	ContainerSamples []ContainerSample `json:"container_samples"`
}

// Container is one container of a host, Beszel's measures and docker's sizes
// put together.
type Container struct {
	Name        string  `json:"name"`
	Image       string  `json:"image"`
	State       string  `json:"state"`
	Status      string  `json:"status"`
	CPU         float64 `json:"cpu"`
	Memory      float64 `json:"memory"`
	MemoryShare float64 `json:"memory_share"`
	Net         float64 `json:"net"`
	SizeRw      int64   `json:"size_rw"`
	SizeVirtual int64   `json:"size_virtual"`
	Measured    bool    `json:"measured"`
	Sized       bool    `json:"sized"`
}

// Filesystem is one durable filesystem of a host.
type Filesystem struct {
	Mount   string  `json:"mount"`
	Device  string  `json:"device"`
	Type    string  `json:"type"`
	Size    int64   `json:"size"`
	Used    int64   `json:"used"`
	Free    int64   `json:"free"`
	Percent float64 `json:"percent"`
}

// ContainerSample is one record of a host's containers.
type ContainerSample struct {
	At         string          `json:"at"`
	Containers []ContainerStat `json:"containers"`
}

// ContainerStat is one container of a record.
type ContainerStat struct {
	Name   string  `json:"name"`
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
}
