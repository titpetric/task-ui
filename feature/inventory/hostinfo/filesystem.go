package hostinfo

import (
	"sort"
	"strconv"
	"strings"

	"github.com/titpetric/task-ui/beszel"
)

// Filesystem is a mounted filesystem: where, of which device and type,
// and its size, use and free space in bytes. Source says where the row
// comes from: df over ssh, or beszel.
type Filesystem struct {
	Mount  string `json:"mount"`
	Device string `json:"device"`
	Type   string `json:"type"`
	Size   uint64 `json:"size"`
	Used   uint64 `json:"used"`
	Free   uint64 `json:"free"`
	Source string `json:"source"`
}

// Percent is how full the filesystem is.
func (f Filesystem) Percent() float64 {
	if f.Size == 0 {
		return 0
	}
	return float64(f.Used) * 100 / float64(f.Size)
}

// ParseDF reads the output of df -PT -B1 into filesystems, mount points with spaces included.
func ParseDF(out string) []Filesystem {
	var filesystems []Filesystem
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 7 || fields[0] == "Filesystem" {
			continue
		}
		size, err1 := strconv.ParseUint(fields[2], 10, 64)
		used, err2 := strconv.ParseUint(fields[3], 10, 64)
		free, err3 := strconv.ParseUint(fields[4], 10, 64)
		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}
		filesystems = append(filesystems, Filesystem{
			Device: fields[0], Type: fields[1], Size: size, Used: used, Free: free,
			Mount: strings.Join(fields[6:], " "), Source: "df",
		})
	}
	return filesystems
}

// volatile are the filesystem types that hold nothing past a reboot, or
// show what another filesystem holds.
var volatile = map[string]bool{
	"tmpfs": true, "devtmpfs": true, "ramfs": true, "overlay": true, "squashfs": true,
	"efivarfs": true, "proc": true, "sysfs": true, "cgroup": true, "cgroup2": true,
	"nsfs": true, "autofs": true, "fuse.lxcfs": true, "devpts": true, "mqueue": true,
	"debugfs": true, "tracefs": true, "securityfs": true, "pstore": true, "bpf": true,
	"hugetlbfs": true, "configfs": true, "fusectl": true, "binfmt_misc": true,
	"rpc_pipefs": true, "shm": true, "none": true,
}

// Durable keeps the filesystems that hold data past a reboot, one per device, sorted by mount point.
func Durable(filesystems []Filesystem) []Filesystem {
	byDevice := map[string]Filesystem{}
	for _, fs := range filesystems {
		if volatile[fs.Type] || fs.Size == 0 || dockers(fs.Mount) || !strings.HasPrefix(fs.Mount, "/") {
			continue
		}
		if kept, ok := byDevice[fs.Device]; ok && len(kept.Mount) <= len(fs.Mount) {
			continue
		}
		byDevice[fs.Device] = fs
	}
	durable := make([]Filesystem, 0, len(byDevice))
	for _, fs := range byDevice {
		durable = append(durable, fs)
	}
	sort.Slice(durable, func(i, j int) bool { return durable[i].Mount < durable[j].Mount })
	return durable
}

// dockers reports a mount point docker made for a container.
func dockers(mount string) bool {
	for _, prefix := range []string{"/var/lib/docker/", "/run/docker/", "/var/lib/containers/", "/run/containerd/"} {
		if strings.HasPrefix(mount, prefix) {
			return true
		}
	}
	return false
}

// gib is a GB as Beszel counts one, in bytes.
const gib = 1 << 30

// BeszelFilesystems returns the root and each extra filesystem a host's agent watches, sizes from GB.
// Without stats it returns none.
func BeszelFilesystems(stats *beszel.Stats) []Filesystem {
	if stats == nil || stats.Disk == 0 {
		return nil
	}
	fs := func(mount string, total, used float64) Filesystem {
		size, use := uint64(total*gib), uint64(used*gib)
		return Filesystem{Mount: mount, Size: size, Used: use, Free: size - min(use, size), Source: "beszel"}
	}
	filesystems := []Filesystem{fs("/", stats.Disk, stats.DiskUsed)}
	names := make([]string, 0, len(stats.ExtraFs))
	for name := range stats.ExtraFs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		filesystems = append(filesystems, fs(name, stats.ExtraFs[name].Total, stats.ExtraFs[name].Used))
	}
	return filesystems
}
