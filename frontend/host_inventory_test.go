package phpfront_test

import (
	"context"

	"github.com/titpetric/task-ui/frontend"
)

// What the fixture answers of the Inventory: the hub's hosts, the cluster and
// what a host last told over its login, grouped the way host_inventory.go
// declares the shapes.

// The hosts permission decides whether task-ui reads a host now, so the
// double answers a login only where the reader has it. A held address the
// double reports nothing of is a host with no measures, as the page's own
// inventory is.
func (h *host) Inventory(_ context.Context, hosts bool, held map[string]string) phpfront.Inventory {
	out := phpfront.Inventory{
		Hosts: []phpfront.System{{
			ID: "sys1", Name: "web1", Host: "10.0.0.1", Status: "up", Up: true,
			CPUs: 8, Agent: "0.9.1", Uptime: "3 days",
			CPU: 12.5, Memory: 48, Disk: 71,
			Measured: true, MemoryTotal: 32, MemoryUsed: 15.5, DiskTotal: 500, DiskUsed: 355,
		}},
		Cluster: phpfront.Cluster{
			Up: 1, Pools: []phpfront.Pool{{
				Name: "CPU", Unit: "cores", Figure: "1 / 8 cores", Used: 1, Total: 8, Percent: 12.5,
				Segments: []phpfront.Segment{{Host: "web1", Width: 100, Fill: 12.5, Title: "web1: 1 of 8 cores"}},
			}},
			Hosts: []phpfront.ClusterHost{{Name: "web1"}},
		},
		History: []phpfront.Series{{ID: "sys1", Samples: []phpfront.Sample{{At: "2026-11-02 10:00:00 +0000 UTC", CPU: 9, Sent: 1024}}}},
	}
	if hosts {
		out.Logins = []phpfront.Login{{Host: "10.0.0.1", Login: "debian@10.0.0.1"}}
	}
	for address, name := range held {
		if address == "10.0.0.1" {
			continue
		}
		out.Hosts = append(out.Hosts, phpfront.System{Name: name, Host: address})
	}
	return out
}

func (h *host) Reachable(context.Context) []string {
	return []string{"10.0.0.1"}
}

// Reach signs in to the one address the double knows.
func (h *host) Reach(_ context.Context, address string) string {
	if address != "10.0.0.2" {
		return ""
	}
	return "debian@" + address
}

func (h *host) HostFacts(_ context.Context, name string, hosts bool) phpfront.HostFacts {
	if name != "web1" {
		return phpfront.HostFacts{}
	}
	out := phpfront.HostFacts{
		Found:   true,
		System:  phpfront.System{ID: "sys1", Name: "web1", Host: "10.0.0.1", Up: true, CPUs: 8},
		Unsized: "task-ui has no ssh login to the host, for docker ps",
		Containers: []phpfront.Container{{
			Name: "api", Image: "api:1", State: "running", Status: "Up 2 days",
			CPU: 3.5, Memory: 512, MemoryShare: 1.5, Net: 2048, Measured: true,
		}},
		Measured:    true,
		Filesystems: []phpfront.Filesystem{{Mount: "/", Size: 1 << 30, Used: 1 << 29, Free: 1 << 29, Percent: 50}},
	}
	if hosts {
		out.Login = "debian@10.0.0.1"
		out.Read = "2026-11-02 10:32:12.414 +0000 UTC"
	}
	return out
}
