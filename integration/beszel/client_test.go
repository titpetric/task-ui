package beszel_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/integration/beszel"
	"github.com/titpetric/task-ui/internal/assert"
)

// A client with no hub reaches nothing: the request has no host.
func TestClient(t *testing.T) {
	t.Parallel()
	client := &beszel.Client{}
	_, err := client.Systems(context.Background())
	assert.Error(t, err)
	_, err = client.Alive(context.Background())
	assert.Error(t, err)
}
