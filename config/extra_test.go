package config_test

import (
	"testing"

	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/internal/assert"
)

// The keys section adds key/read, which every path list holds and the
// section's all leaves out; an extra of no section, or one that names a
// rule every section has, fails to parse.
func TestExtra(t *testing.T) {
	t.Parallel()
	p := config.Default()
	assert.Equal(t, []config.Extra{{Section: "key", Entry: config.Entry{Key: "read", Title: "Allow Read Private"}}}, p.Extra)
	assert.Equal(t, "key/read", p.Extra[0].Path())
	assert.True(t, p.IsExtra("key/read"))
	assert.False(t, p.IsExtra("key/edit"))
	assert.True(t, p.Valid("key/read"))
	assert.False(t, p.Valid("secret/read"))
	assert.True(t, p.Checkable("key/read"))
	assert.Contains(t, p.Checks(), "key/read")

	base := "rules:\n  all: A\n  view: V\n  edit: E\n  delete: D\nsections:\n  key: Keys\npermissions:\n  run: Run\n"
	_, err := config.Parse([]byte(base + "extra:\n  nope:\n    read: R\n"))
	assert.ErrorContains(t, err, "no section")
	_, err = config.Parse([]byte(base + "extra:\n  key:\n    view: R\n"))
	assert.ErrorContains(t, err, "rule of every section")
	plain, err := config.Parse([]byte(base))
	assert.NoError(t, err)
	assert.Len(t, plain.Extra, 0)
}
