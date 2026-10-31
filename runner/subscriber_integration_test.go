//go:build integration

package runner_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

func TestSubscriber(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	// A window that never reads loses output but still gets the status:
	// the channel has a bound, and the run never slows down.
	session, err := f.manager.Start(context.Background(), f.shellJob("flood", "seq 1 5000; exit 0", false))
	assert.NoError(t, err)
	sub := session.Attach(true)
	<-session.Done()
	// No wait here: the range below blocks until the subscription closes,
	// which is the thing a sleep was standing in for.
	var last *runner.Status
	for event := range sub.C {
		if event.Status != nil {
			last = event.Status
		}
	}
	assert.NotNil(t, last)
	assert.Equal(t, model.StateOK, last.State)
}
