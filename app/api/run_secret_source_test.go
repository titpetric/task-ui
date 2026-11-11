package api_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
)

// The manager is a source of a run's secrets; a token of no running run
// gets none.
func TestRunSecretSource(t *testing.T) {
	t.Parallel()
	var source api.RunSecretSource = runner.New(runner.Options{})
	_, err := source.Secret(context.Background(), "nope", "DEMO")
	assert.ErrorIs(t, err, runner.ErrRunToken)
	_, err = fakeRunSecrets{}.Secret(context.Background(), "run-token", "DEMO")
	assert.NoError(t, err)
}
