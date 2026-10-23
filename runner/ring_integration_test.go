//go:build integration

package runner_test

import (
	"context"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

func TestRing(t *testing.T) {
	t.Parallel()
	// The tail bounds the replay a late window gets: a run that printed more
	// than the tail keeps replays only the end of it.
	f := newFixture(t)
	session, err := f.manager.Start(context.Background(), f.shellJob("tail", "yes abcdefghij | head -c 600000; echo END", false))
	assert.NoError(t, err)
	<-session.Done()
	out, _ := collect(t, session.Attach(true))
	assert.True(t, strings.HasSuffix(strings.TrimSpace(out), "END"))
	assert.LessOrEqual(t, len(out), 256<<10)
	assert.Greater(t, len(out), 200<<10)
}
