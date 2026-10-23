package hostinfo

import (
	"strconv"

	"github.com/titpetric/task-ui/beszel"
)

// Colors is how many colours the hosts of a cluster take turns at.
const Colors = 6

// Cluster is the summary of the inventory: the hosts up and down, and a
// pool a resource, CPU, memory, disk and containers, over the hosts that
// are up and report their measures. Hosts is the legend, in the order of
// the systems.
type Cluster struct {
	Up    int
	Down  int
	Pools []Pool
	Hosts []ClusterHost
	// Left counts the hosts the pools leave out.
	Left int
}

// NewCluster sums the systems into pools, with each host's running containers and durable disks by host ID.
func NewCluster(systems []beszel.System, containers map[string][]beszel.Container, disks map[string][]Filesystem) Cluster {
	var c Cluster
	cpu := Pool{Name: "CPU", Unit: "cores"}
	memory := Pool{Name: "Memory", Unit: "GiB"}
	disk := Pool{Name: "Disk", Unit: "GiB"}
	running := Pool{Name: "Containers", Unit: "running"}
	color := 0
	for _, system := range systems {
		if system.Up() {
			c.Up++
		} else {
			c.Down++
		}
		if !system.Up() || system.Stats == nil {
			c.Left++
			c.Hosts = append(c.Hosts, ClusterHost{Name: system.Name, Color: -1, Left: true})
			continue
		}
		host := ClusterHost{Name: system.Name, Color: color % Colors}
		color++
		c.Hosts = append(c.Hosts, host)
		segment := func(used, total float64) Segment {
			return Segment{Host: host.Name, Color: host.Color, Used: used, Total: total}
		}
		cores := float64(system.CPUs())
		cpu.add(segment(cores*system.Info.CPU/100, cores))
		memory.add(segment(system.Stats.MemoryUsed, system.Stats.Memory))
		filesystems := disks[system.ID]
		if len(filesystems) == 0 {
			filesystems = BeszelFilesystems(system.Stats)
		}
		var size, used uint64
		for _, fs := range filesystems {
			size += fs.Size
			used += fs.Used
		}
		disk.add(segment(float64(used)/gib, float64(size)/gib))
		n := float64(len(containers[system.ID]))
		running.add(segment(n, n))
	}
	for _, pool := range []*Pool{&cpu, &memory, &disk, &running} {
		pool.finish()
	}
	running.Figure = strconv.Itoa(int(running.Total)) + " running"
	for i := range running.Segments {
		s := &running.Segments[i]
		s.Title = s.Host + ": " + strconv.Itoa(int(s.Total)) + " running"
	}
	c.Pools = []Pool{cpu, memory, disk, running}
	return c
}

// add puts a host's segment in the pool and its amounts in the sums.
func (p *Pool) add(s Segment) {
	p.Used += s.Used
	p.Total += s.Total
	p.Segments = append(p.Segments, s)
}
