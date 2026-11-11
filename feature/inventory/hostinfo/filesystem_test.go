package hostinfo_test

import (
	"testing"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/hostinfo"
	"github.com/titpetric/task-ui/internal/assert"
)

const df = `Filesystem     Type     1-blocks         Used    Available Capacity Mounted on
udev           devtmpfs 33641611264            0  33641611264       0% /dev
tmpfs          tmpfs     6731628544      2240512   6729388032       1% /run
/dev/nvme0n1p2 ext4    100822970368  44238602240  51428372480      47% /
/dev/nvme0n1p1 vfat       535805952      6197248    529608704       2% /boot/efi
/dev/sda1      ext4   2000035000320 1185736835072 712648785920      63% /srv/backup
/dev/sda1      ext4   2000035000320 1185736835072 712648785920      63% /mnt/data
overlay        overlay 100822970368  44238602240  51428372480      47% /var/lib/docker/overlay2/3f1c0a/merged
/dev/nvme0n1p2 ext4    100822970368  44238602240  51428372480      47% /var/lib/docker/containers/ab/hostname
/dev/sdb1      xfs      1000000000    500000000    500000000      50% /mnt/my disk
broken line
`

// df's lines parse, the mount point with its spaces.
func TestParseDF(t *testing.T) {
	t.Parallel()
	all := hostinfo.ParseDF(df)
	assert.Len(t, all, 9)
	assert.Equal(t, hostinfo.Filesystem{Mount: "/", Device: "/dev/nvme0n1p2", Type: "ext4", Size: 100822970368, Used: 44238602240, Free: 51428372480, Source: "df"}, all[2])
	assert.Equal(t, "/mnt/my disk", all[8].Mount)
	assert.Equal(t, 50.0, all[8].Percent())
	assert.Equal(t, 0.0, hostinfo.Filesystem{}.Percent())
}

// The durable filesystems: no tmpfs or overlay, none of docker's mounts,
// one row a device under its shortest mount point, the extra disk kept.
func TestDurable(t *testing.T) {
	t.Parallel()
	var mounts []string
	for _, fs := range hostinfo.Durable(hostinfo.ParseDF(df)) {
		mounts = append(mounts, fs.Mount)
	}
	assert.Equal(t, []string{"/", "/boot/efi", "/mnt/data", "/mnt/my disk"}, mounts)
}

// Beszel's filesystems are the root and the extra ones, by name.
func TestBeszelFilesystems(t *testing.T) {
	t.Parallel()
	assert.Len(t, hostinfo.BeszelFilesystems(nil), 0)
	stats := &beszel.Stats{Disk: 100, DiskUsed: 25, ExtraFs: map[string]beszel.FsStats{"sdb": {Total: 2, Used: 1}, "data": {Total: 10, Used: 1}}}
	fs := hostinfo.BeszelFilesystems(stats)
	assert.Len(t, fs, 3)
	assert.Equal(t, "/", fs[0].Mount)
	assert.Equal(t, 25.0, fs[0].Percent())
	assert.Equal(t, "data", fs[1].Mount)
	assert.Equal(t, uint64(9<<30), fs[1].Free)
	assert.Equal(t, "beszel", fs[2].Source)
}
