package hostinfo

import (
	"sort"

	"github.com/titpetric/task-ui/beszel"
)

// Container is a row of a host's containers, the two sources put
// together by name: Beszel's measures, CPU in percent of the host,
// memory in MB and its share of the host's, the traffic in bytes/s, and
// docker's sizes; image and status from either. Measured and Sized say
// which source had the container.
type Container struct {
	Name        string
	Image       string
	State       string
	Status      string
	CPU         float64
	Memory      float64
	MemoryShare float64
	Net         float64
	SizeRw      uint64
	SizeVirtual uint64
	Measured    bool
	Sized       bool
}

// Merge puts Beszel's containers of a host and docker's together by
// name, the busiest first: by CPU, then memory, then name. hostMemory is
// the host's RAM in GB, for each container's share of it.
func Merge(measured []beszel.Container, sized []DockerContainer, hostMemory float64) []Container {
	byName := map[string]*Container{}
	var rows []*Container
	row := func(name string) *Container {
		if c, ok := byName[name]; ok {
			return c
		}
		c := &Container{Name: name}
		byName[name] = c
		rows = append(rows, c)
		return c
	}
	for _, m := range measured {
		c := row(m.Name)
		c.Measured, c.CPU, c.Memory, c.Net = true, m.CPU, m.Memory, m.Net
		c.Image, c.Status = m.Image, m.Status
		if hostMemory > 0 {
			c.MemoryShare = m.Memory / (hostMemory * 1024) * 100
		}
	}
	for _, d := range sized {
		c := row(d.Name)
		c.Sized, c.SizeRw, c.SizeVirtual, c.State = true, d.SizeRw, d.SizeVirtual, d.State
		c.Image, c.Status = d.Image, d.Status
	}
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.CPU != b.CPU {
			return a.CPU > b.CPU
		}
		if a.Memory != b.Memory {
			return a.Memory > b.Memory
		}
		return a.Name < b.Name
	})
	merged := make([]Container, len(rows))
	for i, c := range rows {
		merged[i] = *c
	}
	return merged
}
