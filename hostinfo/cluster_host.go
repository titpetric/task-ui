package hostinfo

// ClusterHost is a host in the cluster summary's legend: its name, its
// colour in the pools, and whether it counts in them: Left for a host
// that is down or sends no measures.
type ClusterHost struct {
	Name  string
	Color int
	Left  bool
}
