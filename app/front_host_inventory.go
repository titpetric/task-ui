package app

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/hostinfo"
	"github.com/titpetric/task-ui/phpfront"
)

// frontProbe is how long the inventory waits for an ssh login to answer,
// and frontRead how long a host's own page waits for the host itself. Both
// are frontend/integration.go's.
const (
	frontProbe = 4 * time.Second
	frontRead  = 15 * time.Second
)

// Inventory is the hosts page's facts: what Beszel monitors largest first,
// who task-ui signs in to each as, the pools the hosts sum into, and the
// month of records the cluster's charts draw.
//
// hosts is the reader's permission, which PHP decides. It is what makes
// task-ui read a host rather than answer out of what the host last told.
func (h *FrontHost) Inventory(ctx context.Context, hosts bool) phpfront.Inventory {
	if h.Hosts == nil {
		return phpfront.Inventory{Off: true}
	}
	var out phpfront.Inventory
	systems, err := h.Hosts.Systems(ctx)
	switch {
	case errors.Is(err, beszel.ErrOff):
		out.Off = true
	case err != nil:
		out.Error = err.Error()
	}
	sorted := beszel.ByCapacity(systems)
	out.Hosts = frontSystems(sorted)
	if len(sorted) == 0 {
		return out
	}
	logins := h.logins(ctx, sorted)
	out.Logins = frontLogins(sorted, logins)

	var running map[string][]beszel.Container
	var records map[string][]beszel.Sample
	if hub, err := h.Hosts.Hub(ctx); err == nil {
		running, _ = hub.Containers(ctx)
		records, _ = hub.History(ctx)
	}
	disks := map[string][]hostinfo.Filesystem{}
	for _, each := range sorted {
		if told := h.told(ctx, logins[each.Host], hosts, 0); told != nil {
			disks[each.ID] = told.Filesystems
		}
	}
	summary := hostinfo.NewCluster(sorted, running, disks)
	out.Cluster = frontCluster(summary)
	out.History = frontHistory(records)
	return out
}

// HostFacts is one host of the inventory in full, the expanded row of its
// line: its containers as Beszel and docker ps report them together, its
// durable filesystems, and the records its charts draw.
func (h *FrontHost) HostFacts(ctx context.Context, name string, hosts bool) phpfront.HostFacts {
	if h.Hosts == nil {
		return phpfront.HostFacts{Error: beszel.ErrOff.Error()}
	}
	systems, err := h.Hosts.Systems(ctx)
	if err != nil {
		return phpfront.HostFacts{Error: err.Error()}
	}
	var found *beszel.System
	for i := range systems {
		if systems[i].Name == name {
			found = &systems[i]
		}
	}
	if found == nil {
		return phpfront.HostFacts{}
	}
	out := phpfront.HostFacts{
		Found:   true,
		System:  frontSystem(*found),
		Unsized: "task-ui has no ssh login to the host, for docker ps",
	}
	if h.SSH != nil {
		out.Login = h.SSH.Logins(ctx, []string{found.Host}, frontProbe)[found.Host]
	}
	var memory float64
	if found.Stats != nil {
		memory = found.Stats.Memory
	}
	var measured []beszel.Container
	if hub, err := h.Hosts.Hub(ctx); err == nil {
		running, _ := hub.Containers(ctx)
		measured, out.Measured = running[found.ID], len(running[found.ID]) > 0
		if records, _ := hub.History(ctx); len(records[found.ID]) > 0 {
			out.Samples = frontSamples(records[found.ID])
		}
		if samples, _ := hub.ContainerHistory(ctx, found.ID); len(samples) > 0 {
			out.ContainerSamples = frontContainerSamples(samples)
		}
	}
	var sized []hostinfo.DockerContainer
	told := h.told(ctx, out.Login, hosts, frontRead)
	switch {
	case out.Login == "":
	case told == nil:
		out.Unsized = "the host has not answered over ssh yet"
	case told.Err != nil:
		out.Unsized = "ssh: " + told.Err.Error()
	case !told.Docker:
		out.Unsized = "docker ps failed for " + out.Login + ", and through sudo"
	default:
		sized, out.Sized = told.Containers, true
	}
	if told != nil && !told.At.IsZero() {
		// Round(0) strips the monotonic reading a time.Now() carries, which
		// String() would write as a trailing m=+0.000000001 that no layout
		// reads.
		out.Read = told.At.Round(0).String()
	}
	out.Containers = frontContainers(hostinfo.Merge(measured, sized, memory))
	if told != nil && len(told.Filesystems) > 0 {
		out.Filesystems, out.FromDF = frontFilesystems(told.Filesystems), true
	} else {
		out.Filesystems = frontFilesystems(hostinfo.BeszelFilesystems(found.Stats))
	}
	return out
}

// Reachable is the hosts a command can be run on, largest capacity first:
// the ones Beszel monitors that task-ui has an ssh login to. It is what a
// job picker offers, and frontend/integration.go reads the same two calls
// for the same list.
func (h *FrontHost) Reachable(ctx context.Context) []string {
	if h.Hosts == nil || h.SSH == nil {
		return nil
	}
	systems, err := h.Hosts.Systems(ctx)
	if err != nil || len(systems) == 0 {
		return nil
	}
	sorted := beszel.ByCapacity(systems)
	logins := h.logins(ctx, sorted)
	out := make([]string, 0, len(sorted))
	for _, each := range sorted {
		if logins[each.Host] != "" {
			out = append(out, each.Host)
		}
	}
	return out
}

// logins is who task-ui signs in to each of the systems as.
func (h *FrontHost) logins(ctx context.Context, systems []beszel.System) map[string]string {
	if h.SSH == nil {
		return nil
	}
	addresses := make([]string, len(systems))
	for i, each := range systems {
		addresses[i] = each.Host
	}
	return h.SSH.Logins(ctx, addresses, frontProbe)
}

// told is what a login's host last said over ssh, waiting at most wait for
// a fresh answer. Only a reader with the hosts permission makes task-ui read
// a host; everyone else sees what it kept.
func (h *FrontHost) told(ctx context.Context, login string, hosts bool, wait time.Duration) *hostinfo.Facts {
	if h.Facts == nil || login == "" {
		return nil
	}
	if hosts {
		return h.Facts.Facts(ctx, login, wait)
	}
	return h.Facts.Cached(login)
}

// frontSystems is the hosts Beszel monitors, in the order they are listed.
func frontSystems(systems []beszel.System) []phpfront.System {
	out := make([]phpfront.System, 0, len(systems))
	for _, each := range systems {
		out = append(out, frontSystem(each))
	}
	return out
}

// frontSystem is one host as Beszel reports it. The shares its line meters
// are the agent's last word and the sizes behind them the hub's last minute
// of measures, which a host without one has none of.
func frontSystem(each beszel.System) phpfront.System {
	out := phpfront.System{
		ID:     each.ID,
		Name:   each.Name,
		Host:   each.Host,
		Status: each.Status,
		Up:     each.Up(),
		CPUs:   each.CPUs(),
		Agent:  each.Info.Agent,
		Uptime: frontUptime(each.UptimeDuration()),
		CPU:    each.Info.CPU,
		Memory: each.Info.Memory,
		Disk:   each.Info.Disk,
	}
	if each.Stats != nil {
		out.Measured = true
		out.MemoryTotal = each.Stats.Memory
		out.MemoryUsed = each.Stats.MemoryUsed
		out.DiskTotal = each.Stats.Disk
		out.DiskUsed = each.Stats.DiskUsed
	}
	return out
}

// frontUptime says how long a host has run: in days past two, else in
// hours, else in minutes.
func frontUptime(up time.Duration) string {
	switch {
	case up >= 48*time.Hour:
		return strconv.Itoa(int(up/(24*time.Hour))) + " days"
	case up >= 2*time.Hour:
		return strconv.Itoa(int(up/time.Hour)) + " hours"
	}
	return strconv.Itoa(int(up/time.Minute)) + " minutes"
}

// frontLogins is the ssh logins of the systems that have one, in the order
// the systems are listed, so the page reads the same list twice over.
func frontLogins(systems []beszel.System, logins map[string]string) []phpfront.Login {
	out := make([]phpfront.Login, 0, len(logins))
	for _, each := range systems {
		if login := logins[each.Host]; login != "" {
			out = append(out, phpfront.Login{Host: each.Host, Login: login})
		}
	}
	return out
}

// frontCluster is the hosts summed into pools.
func frontCluster(summary hostinfo.Cluster) phpfront.Cluster {
	out := phpfront.Cluster{
		Up:    summary.Up,
		Down:  summary.Down,
		Left:  summary.Left,
		Pools: make([]phpfront.Pool, 0, len(summary.Pools)),
		Hosts: make([]phpfront.ClusterHost, 0, len(summary.Hosts)),
	}
	for _, pool := range summary.Pools {
		each := phpfront.Pool{
			Name:     pool.Name,
			Unit:     pool.Unit,
			Figure:   pool.Figure,
			Used:     pool.Used,
			Total:    pool.Total,
			Percent:  pool.Percent(),
			Segments: make([]phpfront.Segment, 0, len(pool.Segments)),
		}
		for _, segment := range pool.Segments {
			each.Segments = append(each.Segments, phpfront.Segment{
				Host:  segment.Host,
				Color: segment.Color,
				Width: segment.Width,
				Fill:  segment.Fill,
				Title: segment.Title,
			})
		}
		out.Pools = append(out.Pools, each)
	}
	for _, each := range summary.Hosts {
		out.Hosts = append(out.Hosts, phpfront.ClusterHost{Name: each.Name, Color: each.Color, Left: each.Left})
	}
	return out
}

// frontHistory is Beszel's records of each host, by host ID and in that
// order, which a map has none of.
func frontHistory(records map[string][]beszel.Sample) []phpfront.Series {
	ids := make([]string, 0, len(records))
	for id := range records {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]phpfront.Series, 0, len(ids))
	for _, id := range ids {
		out = append(out, phpfront.Series{ID: id, Samples: frontSamples(records[id])})
	}
	return out
}

// frontSamples is a host's records.
func frontSamples(samples []beszel.Sample) []phpfront.Sample {
	out := make([]phpfront.Sample, 0, len(samples))
	for _, sample := range samples {
		sent, recv := sample.Stats.Net()
		out = append(out, phpfront.Sample{
			At:         sample.At.String(),
			CPU:        sample.Stats.CPU,
			Memory:     sample.Stats.Memory,
			MemoryUsed: sample.Stats.MemoryUsed,
			Disk:       sample.Stats.Disk,
			DiskUsed:   sample.Stats.DiskUsed,
			Sent:       sent,
			Recv:       recv,
			Extra:      frontMounts(sample.Stats.ExtraFs),
		})
	}
	return out
}

// frontMounts is the filesystems of a record beside the root, by name.
func frontMounts(extra map[string]beszel.FsStats) []phpfront.Mount {
	names := make([]string, 0, len(extra))
	for name := range extra {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]phpfront.Mount, 0, len(names))
	for _, name := range names {
		out = append(out, phpfront.Mount{Name: name, Used: extra[name].Used, Total: extra[name].Total})
	}
	return out
}

// frontContainerSamples is the records of a host's containers.
func frontContainerSamples(samples []beszel.ContainerSample) []phpfront.ContainerSample {
	out := make([]phpfront.ContainerSample, 0, len(samples))
	for _, sample := range samples {
		stats := make([]phpfront.ContainerStat, 0, len(sample.Containers))
		for _, stat := range sample.Containers {
			stats = append(stats, phpfront.ContainerStat{Name: stat.Name, CPU: stat.CPU, Memory: stat.Memory})
		}
		out = append(out, phpfront.ContainerSample{At: sample.At.String(), Containers: stats})
	}
	return out
}

// frontContainers is a host's containers, the two sources already merged.
func frontContainers(merged []hostinfo.Container) []phpfront.Container {
	out := make([]phpfront.Container, 0, len(merged))
	for _, each := range merged {
		out = append(out, phpfront.Container{
			Name:        each.Name,
			Image:       each.Image,
			State:       each.State,
			Status:      each.Status,
			CPU:         each.CPU,
			Memory:      each.Memory,
			MemoryShare: each.MemoryShare,
			Net:         each.Net,
			SizeRw:      int64(each.SizeRw),
			SizeVirtual: int64(each.SizeVirtual),
			Measured:    each.Measured,
			Sized:       each.Sized,
		})
	}
	return out
}

// frontFilesystems is a host's durable filesystems.
func frontFilesystems(filesystems []hostinfo.Filesystem) []phpfront.Filesystem {
	out := make([]phpfront.Filesystem, 0, len(filesystems))
	for _, fs := range filesystems {
		out = append(out, phpfront.Filesystem{
			Mount:   fs.Mount,
			Device:  fs.Device,
			Type:    fs.Type,
			Size:    int64(fs.Size),
			Used:    int64(fs.Used),
			Free:    int64(fs.Free),
			Percent: fs.Percent(),
		})
	}
	return out
}
