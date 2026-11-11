package phpfront

import (
	"github.com/titpetric/phpscript/model"
)

// inventory is the hosts page's facts as a PHP array.
func inventory(found Inventory) *model.Array {
	return phpMap(
		"off", found.Off,
		"error", found.Error,
		"hosts", systems(found.Hosts),
		"logins", logins(found.Logins),
		"cluster", cluster(found.Cluster),
		"history", history(found.History),
	)
}

// systems is the hosts Beszel monitors as a PHP array.
func systems(found []System) *model.Array {
	out := model.NewArraySize(len(found))
	for _, each := range found {
		out.Append(system(each))
	}
	return out
}

// system is one host as Beszel reports it.
func system(found System) *model.Array {
	return phpMap(
		"id", found.ID,
		"name", found.Name,
		"host", found.Host,
		"status", found.Status,
		"up", found.Up,
		"cpus", found.CPUs,
		"agent", found.Agent,
		"uptime", found.Uptime,
		"cpu", found.CPU,
		"memory", found.Memory,
		"disk", found.Disk,
		"measured", found.Measured,
		"memory_total", found.MemoryTotal,
		"memory_used", found.MemoryUsed,
		"disk_total", found.DiskTotal,
		"disk_used", found.DiskUsed,
	)
}

// logins is the ssh logins as a PHP array, keyed by host so a page reads one
// by the address of its line.
func logins(found []Login) *model.Array {
	out := model.NewArraySize(len(found))
	for _, login := range found {
		out.Set(login.Host, login.Login)
	}
	return out
}

// cluster is the pools as a PHP array.
func cluster(found Cluster) *model.Array {
	return phpMap(
		"up", found.Up,
		"down", found.Down,
		"left", found.Left,
		"pools", pools(found.Pools),
		"hosts", legend(found.Hosts),
	)
}

// pools is the resources the hosts share as a PHP array.
func pools(found []Pool) *model.Array {
	out := model.NewArraySize(len(found))
	for _, pool := range found {
		out.Append(phpMap(
			"name", pool.Name,
			"unit", pool.Unit,
			"figure", pool.Figure,
			"used", pool.Used,
			"total", pool.Total,
			"percent", pool.Percent,
			"segments", segments(pool.Segments),
		))
	}
	return out
}

// segments is a pool's bar as a PHP array.
func segments(found []Segment) *model.Array {
	out := model.NewArraySize(len(found))
	for _, segment := range found {
		out.Append(phpMap(
			"host", segment.Host,
			"color", segment.Color,
			"width", segment.Width,
			"fill", segment.Fill,
			"title", segment.Title,
		))
	}
	return out
}

// legend is the pools' legend as a PHP array.
func legend(found []ClusterHost) *model.Array {
	out := model.NewArraySize(len(found))
	for _, entry := range found {
		out.Append(phpMap("name", entry.Name, "color", entry.Color, "left", entry.Left))
	}
	return out
}

// history is Beszel's records of each host as a PHP array, keyed by host ID.
func history(found []Series) *model.Array {
	out := model.NewArraySize(len(found))
	for _, series := range found {
		out.Set(series.ID, samples(series.Samples))
	}
	return out
}

// samples is a host's records as a PHP array.
func samples(found []Sample) *model.Array {
	out := model.NewArraySize(len(found))
	for _, sample := range found {
		out.Append(phpMap(
			"at", sample.At,
			"cpu", sample.CPU,
			"memory", sample.Memory,
			"memory_used", sample.MemoryUsed,
			"disk", sample.Disk,
			"disk_used", sample.DiskUsed,
			"sent", sample.Sent,
			"recv", sample.Recv,
			"extra", mounts(sample.Extra),
		))
	}
	return out
}

// mounts is the filesystems of a record as a PHP array.
func mounts(found []Mount) *model.Array {
	out := model.NewArraySize(len(found))
	for _, mount := range found {
		out.Append(phpMap("name", mount.Name, "used", mount.Used, "total", mount.Total))
	}
	return out
}

// facts is one host of the inventory in full as a PHP array.
func facts(found HostFacts) *model.Array {
	return phpMap(
		"found", found.Found,
		"error", found.Error,
		"system", system(found.System),
		"login", found.Login,
		"containers", containers(found.Containers),
		"measured", found.Measured,
		"sized", found.Sized,
		"unsized", found.Unsized,
		"filesystems", filesystems(found.Filesystems),
		"from_df", found.FromDF,
		"read", found.Read,
		"samples", samples(found.Samples),
		"container_samples", containerSamples(found.ContainerSamples),
	)
}

// containers is a host's containers as a PHP array.
func containers(found []Container) *model.Array {
	out := model.NewArraySize(len(found))
	for _, container := range found {
		out.Append(phpMap(
			"name", container.Name,
			"image", container.Image,
			"state", container.State,
			"status", container.Status,
			"cpu", container.CPU,
			"memory", container.Memory,
			"memory_share", container.MemoryShare,
			"net", container.Net,
			"size_rw", container.SizeRw,
			"size_virtual", container.SizeVirtual,
			"measured", container.Measured,
			"sized", container.Sized,
		))
	}
	return out
}

// filesystems is a host's durable filesystems as a PHP array.
func filesystems(found []Filesystem) *model.Array {
	out := model.NewArraySize(len(found))
	for _, fs := range found {
		out.Append(phpMap(
			"mount", fs.Mount,
			"device", fs.Device,
			"type", fs.Type,
			"size", fs.Size,
			"used", fs.Used,
			"free", fs.Free,
			"percent", fs.Percent,
		))
	}
	return out
}

// containerSamples is the records of a host's containers as a PHP array.
func containerSamples(found []ContainerSample) *model.Array {
	out := model.NewArraySize(len(found))
	for _, sample := range found {
		stats := model.NewArraySize(len(sample.Containers))
		for _, stat := range sample.Containers {
			stats.Append(phpMap("name", stat.Name, "cpu", stat.CPU, "memory", stat.Memory))
		}
		out.Append(phpMap("at", sample.At, "containers", stats))
	}
	return out
}
