package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestHash(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	assert.NoError(t, hash([]string{"secret"}, strings.NewReader(""), &out))
	assert.True(t, access.CheckHash(strings.TrimSpace(out.String()), "secret"))

	out.Reset()
	assert.NoError(t, hash(nil, strings.NewReader("from stdin\n"), &out))
	assert.True(t, access.CheckHash(strings.TrimSpace(out.String()), "from stdin"))

	assert.Error(t, hash(nil, strings.NewReader(""), &out))
}

// get-secret takes one name, in a run's environment.
func TestGetSecret(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	none := func(string) string { return "" }
	assert.Error(t, getSecret(context.Background(), nil, none, &out))
	err := getSecret(context.Background(), []string{"DEMO"}, none, &out)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TASK_UI_RUN_TOKEN")
	assert.Equal(t, "", out.String())
}
