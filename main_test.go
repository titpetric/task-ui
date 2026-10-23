package main

import (
	"bytes"
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
