package hostpath_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/hostpath"
	"github.com/titpetric/task-ui/internal/assert"
)

// mountinfo is the mounts of a task-ui container: /data and /keys bound
// from the host, the terraform state at its host path, and the mounts
// of the container itself.
const mountinfo = `530 354 0:56 / / rw,relatime - overlay overlay rw,lowerdir=/x
537 530 0:86 / /dev rw,nosuid - tmpfs tmpfs rw,size=65536k
564 530 8:1 /mnt/apps/task/data /data rw,relatime - ext4 /dev/sda1 rw
565 530 8:1 /mnt/apps/task/keys /keys rw,relatime - ext4 /dev/sda1 rw
566 530 8:1 /var/lib/docker/containers/6f91/hostname /etc/hostname rw,relatime - ext4 /dev/sda1 rw
588 530 0:25 /docker.sock /run/docker.sock rw,nosuid - tmpfs tmpfs rw
597 530 8:1 /mnt/apps/task/terraform /mnt/apps/task/terraform rw,relatime - ext4 /dev/sda1 rw
598 564 8:1 /srv/with\040space /data/workspaces/odd rw,relatime - ext4 /dev/sda1 rw
`

func TestMounts_Host(t *testing.T) {
	t.Parallel()
	mounts, err := hostpath.Parse(strings.NewReader(mountinfo))
	assert.NoError(t, err)
	assert.Equal(t, "/mnt/apps/task/data/workspaces/malta/vhosts", mounts.Host("/data/workspaces/malta/vhosts"))
	assert.Equal(t, "/mnt/apps/task/data", mounts.Host("/data"))
	assert.Equal(t, "/mnt/apps/task/data", mounts.Host("/data/"))
	assert.Equal(t, "/srv/with space/x", mounts.Host("/data/workspaces/odd/x"), "the deepest mount, unescaped")
	assert.Equal(t, "/mnt/apps/task/terraform", mounts.Host("/mnt/apps/task/terraform"))
	assert.Equal(t, "/datafile", mounts.Host("/datafile"), "a prefix that is no directory")
	assert.Equal(t, "/app", mounts.Host("/app"), "the image's own file system maps nothing")
	assert.Equal(t, "/tmp/x", (hostpath.Mounts)(nil).Host("/tmp/x"), "outside a container")
	_ = hostpath.Read()
}
