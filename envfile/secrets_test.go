package envfile_test

import (
	"context"
	"errors"
	"testing"

	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/internal/assert"
)

// secrets returns the same secrets for every folder, as envfile.Secrets.
type secrets map[string]string

var _ envfile.Secrets = secrets(nil)

func (s secrets) For(context.Context, string) (map[string]string, error) {
	if s["FAIL"] != "" {
		return nil, errors.New(s["FAIL"])
	}
	return s, nil
}

func TestSecrets(t *testing.T) {
	t.Parallel()
	var s envfile.Secrets = secrets{"FAIL": "locked"}
	_, err := s.For(context.Background(), "infra")
	assert.Error(t, err)
}
