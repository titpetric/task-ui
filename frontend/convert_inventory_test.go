package phpfront

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// The keys the Inventory and a host's page index by. The helpers are
// convert_test.go's.
//
// Two of these converters key by something other than position, because the
// page reads one entry and does not walk the list looking for it: the logins
// by host address, and the history by host ID.

// The page's four facts, and that nothing having answered is the unset state
// rather than an error: off is what the page draws its "not configured" line
// from, and error what the hub said when it refused.
func TestInventory(t *testing.T) {
	t.Parallel()

	out := inventory(Inventory{
		Hosts:   []System{{ID: "sys1", Name: "web1"}},
		Logins:  []Login{{Host: "10.0.0.1", Login: "debian@10.0.0.1"}},
		Cluster: Cluster{Up: 1},
		History: []Series{{ID: "sys1", Samples: []Sample{{At: "2026-11-02 10:00:00 +0000 UTC"}}}},
	})

	assert.False(t, flag(t, out, "off"))
	assert.Equal(t, "", text(t, out, "error"))
	assert.Equal(t, 1, list(t, out, "hosts").Len())
	assert.Equal(t, 1, value[int](t, list(t, out, "cluster"), "up"))

	off := inventory(Inventory{Off: true})
	assert.True(t, flag(t, off, "off"))
	assert.Equal(t, 0, list(t, off, "hosts").Len())
	assert.Equal(t, 0, list(t, off, "logins").Len())
	assert.Equal(t, 0, list(t, off, "history").Len())

	failed := inventory(Inventory{Error: "beszel: 502 Bad Gateway"})
	assert.Equal(t, "beszel: 502 Bad Gateway", text(t, failed, "error"))
	assert.False(t, flag(t, failed, "off"), "a hub that refused is configured")
}

// A share is a float64 all the way over and a core count an int, because the
// page prints the one with a decimal and the other without: widening either
// would change what a row says.
func TestSystem(t *testing.T) {
	t.Parallel()

	out := system(System{
		ID: "sys1", Name: "web1", Host: "10.0.0.1", Status: "up", Up: true,
		CPUs: 8, Agent: "0.9.1", Uptime: "3 days",
		CPU: 12.5, Memory: 48, Disk: 71,
		Measured: true, MemoryTotal: 32, MemoryUsed: 15.5, DiskTotal: 500, DiskUsed: 355,
	})

	assert.Equal(t, "sys1", text(t, out, "id"))
	assert.True(t, flag(t, out, "up"))
	assert.Equal(t, 8, value[int](t, out, "cpus"))
	assert.Equal(t, 12.5, value[float64](t, out, "cpu"))
	assert.Equal(t, 15.5, value[float64](t, out, "memory_used"))
	assert.Equal(t, "0.9.1", text(t, out, "agent"))
	assert.Equal(t, "3 days", text(t, out, "uptime"))

	// Nothing having read the host is measured false: the page draws the
	// shares and not the totals, which are zero and not small.
	assert.False(t, flag(t, system(System{}), "measured"))
	assert.Equal(t, 0, systems(nil).Len())
}

// The ssh logins cross keyed by host address, because the row draws its own
// login off the address it is already printing.
func TestLogins(t *testing.T) {
	t.Parallel()

	out := logins([]Login{
		{Host: "10.0.0.1", Login: "debian@10.0.0.1"},
		{Host: "10.0.0.2", Login: "root@10.0.0.2"},
	})
	assert.Equal(t, 2, out.Len())
	assert.Equal(t, "debian@10.0.0.1", text(t, out, "10.0.0.1"))

	// A host with no login is no key, which is the empty string a row of it
	// reads, so the page gates the terminal button on the login being set.
	_, ok := out.Get("10.0.0.3")
	assert.False(t, ok)
}

// A pool's bar is drawn from the segments and nothing else: each carries the
// width of its slice, the fill of what that host uses, and the sentence the
// hover says, so the template does no arithmetic.
func TestCluster(t *testing.T) {
	t.Parallel()

	out := cluster(Cluster{
		Up: 2, Down: 1, Left: 1,
		Pools: []Pool{{
			Name: "CPU", Unit: "cores", Figure: "1 / 8 cores", Used: 1, Total: 8, Percent: 12.5,
			Segments: []Segment{{Host: "web1", Color: 0, Width: 100, Fill: 12.5, Title: "web1: 1 of 8 cores"}},
		}},
		Hosts: []ClusterHost{{Name: "web1"}, {Name: "web3", Color: 2, Left: true}},
	})

	assert.Equal(t, 2, value[int](t, out, "up"))
	assert.Equal(t, 1, value[int](t, out, "down"))
	assert.Equal(t, 1, value[int](t, out, "left"))

	pool := at(t, list(t, out, "pools"), 0)
	assert.Equal(t, "1 / 8 cores", text(t, pool, "figure"))
	assert.Equal(t, 8.0, value[float64](t, pool, "total"))
	assert.Equal(t, 12.5, value[float64](t, pool, "percent"))

	segment := at(t, list(t, pool, "segments"), 0)
	assert.Equal(t, "web1", text(t, segment, "host"))
	assert.Equal(t, 100.0, value[float64](t, segment, "width"))
	assert.Equal(t, 12.5, value[float64](t, segment, "fill"))
	assert.Equal(t, "web1: 1 of 8 cores", text(t, segment, "title"))

	// The legend lists a host the pools leave out, marked as left: it is in
	// the inventory and sends no measures, which is not the same as absent.
	assert.True(t, flag(t, at(t, list(t, out, "hosts"), 1), "left"))
	assert.Equal(t, 2, value[int](t, at(t, list(t, out, "hosts"), 1), "color"))
}

// The records cross keyed by host ID, because a chart reads the series of the
// row it is drawing. Every sample carries the filesystems beside the root,
// which the host's own page draws and the list does not.
func TestHistory(t *testing.T) {
	t.Parallel()

	out := history([]Series{{
		ID: "sys1",
		Samples: []Sample{{
			At: "2026-11-02 10:00:00 +0000 UTC", CPU: 9, MemoryUsed: 15.5, Sent: 1024, Recv: 2048,
			Extra: []Mount{{Name: "/data", Used: 128, Total: 512}},
		}},
	}})
	assert.Equal(t, 1, out.Len())

	sample := at(t, list(t, out, "sys1"), 0)
	assert.Equal(t, "2026-11-02 10:00:00 +0000 UTC", text(t, sample, "at"))
	assert.Equal(t, 9.0, value[float64](t, sample, "cpu"))
	assert.Equal(t, 1024.0, value[float64](t, sample, "sent"))
	assert.Equal(t, 2048.0, value[float64](t, sample, "recv"))

	mount := at(t, list(t, sample, "extra"), 0)
	assert.Equal(t, "/data", text(t, mount, "name"))
	assert.Equal(t, 512.0, value[float64](t, mount, "total"))

	_, ok := history(nil).Get("sys1")
	assert.False(t, ok)
	assert.Equal(t, 0, samples(nil).Len())
	assert.Equal(t, 0, mounts(nil).Len())
}

// One host in full, with the two sentences that say why a number is missing:
// sized false with unsized saying what task-ui could not run, and measured
// false where the agent sent nothing.
func TestFacts(t *testing.T) {
	t.Parallel()

	out := facts(HostFacts{
		Found:   true,
		System:  System{ID: "sys1", Name: "web1", CPUs: 8},
		Login:   "debian@10.0.0.1",
		Read:    "2026-11-02 10:32:12.414 +0000 UTC",
		Unsized: "task-ui has no ssh login to the host, for docker ps",
		Containers: []Container{{
			Name: "api", Image: "api:1", State: "running", Status: "Up 2 days",
			CPU: 3.5, Memory: 512, MemoryShare: 1.5, Net: 2048, Measured: true,
		}},
		Measured:    true,
		FromDF:      true,
		Filesystems: []Filesystem{{Mount: "/", Size: 1 << 30, Used: 1 << 29, Free: 1 << 29, Percent: 50}},
		ContainerSamples: []ContainerSample{{
			At:         "2026-11-02 10:00:00 +0000 UTC",
			Containers: []ContainerStat{{Name: "api", CPU: 3.5, Memory: 512}},
		}},
	})

	assert.True(t, flag(t, out, "found"))
	assert.Equal(t, "web1", text(t, list(t, out, "system"), "name"))
	assert.Equal(t, "debian@10.0.0.1", text(t, out, "login"))
	assert.Equal(t, "2026-11-02 10:32:12.414 +0000 UTC", text(t, out, "read"))
	assert.True(t, flag(t, out, "measured"))
	assert.True(t, flag(t, out, "from_df"), "a size read over ssh, not one the agent reported")
	assert.False(t, flag(t, out, "sized"))
	assert.Equal(t, "task-ui has no ssh login to the host, for docker ps", text(t, out, "unsized"))

	container := at(t, list(t, out, "containers"), 0)
	assert.Equal(t, "api:1", text(t, container, "image"))
	assert.Equal(t, "Up 2 days", text(t, container, "status"))
	assert.Equal(t, 1.5, value[float64](t, container, "memory_share"))
	assert.True(t, flag(t, container, "measured"))
	assert.False(t, flag(t, container, "sized"))

	fs := at(t, list(t, out, "filesystems"), 0)
	assert.Equal(t, "/", text(t, fs, "mount"))
	assert.Equal(t, int64(1<<30), value[int64](t, fs, "size"))
	assert.Equal(t, 50.0, value[float64](t, fs, "percent"))

	record := at(t, list(t, out, "container_samples"), 0)
	assert.Equal(t, "2026-11-02 10:00:00 +0000 UTC", text(t, record, "at"))
	assert.Equal(t, "api", text(t, at(t, list(t, record, "containers"), 0), "name"))

	// A host Beszel monitors none of answers found false with a system of
	// empty keys, so the page draws its heading before it gates.
	none := facts(HostFacts{})
	assert.False(t, flag(t, none, "found"))
	assert.Equal(t, "", text(t, list(t, none, "system"), "name"))
	assert.Equal(t, 0, list(t, none, "containers").Len())
	assert.Equal(t, 0, list(t, none, "filesystems").Len())
	assert.Equal(t, 0, list(t, none, "container_samples").Len())
}
