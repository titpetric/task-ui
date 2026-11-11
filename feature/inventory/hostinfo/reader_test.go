package hostinfo_test

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/titpetric/task-ui/hostinfo"
	"github.com/titpetric/task-ui/internal/assert"
)

// The reader reads a host once in a TTL, waits as long as it's told, and
// keeps a host that does not answer as an error.
func TestReader(t *testing.T) {
	t.Parallel()
	var asks atomic.Int32
	release := make(chan struct{})
	r := &hostinfo.Reader{Timeout: time.Second, TTL: time.Hour, Run: func(ctx context.Context, login, remote string) ([]byte, error) {
		asks.Add(1)
		assert.Contains(t, remote, "docker ps --size")
		if login == "debian@slow" {
			<-release
		}
		if login == "root@down" {
			return nil, errors.New("exit status 255")
		}
		return []byte(df + "@@docker\n@@exit 0\n"), nil
	}}
	ctx := context.Background()
	assert.Nil(t, r.Cached("debian@a"))
	facts := r.Facts(ctx, "debian@a", time.Second)
	assert.NotNil(t, facts)
	assert.Len(t, facts.Filesystems, 4)
	assert.False(t, facts.At.IsZero())
	r.Facts(ctx, "debian@a", time.Second)
	assert.Equal(t, int32(1), asks.Load(), "kept")

	assert.Nil(t, r.Facts(ctx, "debian@slow", 0), "not answered yet")
	close(release)
	assert.Eventually(t, func() bool { return r.Cached("debian@slow") != nil }, time.Second, 10*time.Millisecond)

	down := r.Facts(ctx, "root@down", time.Second)
	assert.True(t, strings.Contains(down.Err.Error(), "255"))
}
