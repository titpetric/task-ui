package config_test

import (
	"testing"

	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/internal/assert"
)

// No section of the embedded file adds a rule: no key/read remains, and
// a policy cannot set it. A file whose section adds one lists it among
// every path and leaves it out of the section's all; an extra of no
// section, or one that names a rule every section has, fails to parse.
func TestExtra(t *testing.T) {
	t.Parallel()
	p := config.Default()
	assert.Len(t, p.Extra, 0)
	assert.False(t, p.IsExtra("key/read"))
	assert.False(t, p.Valid("key/read"))
	assert.False(t, p.Checkable("key/read"))
	assert.NotContains(t, p.Checks(), "key/read")

	base := "rules:\n  all: A\n  view: V\n  edit: E\n  delete: D\nsections:\n  key: Keys\npermissions:\n  run: Run\n"
	with, err := config.Parse([]byte(base + "extra:\n  key:\n    sign: S\n"))
	assert.NoError(t, err)
	assert.Equal(t, []config.Extra{{Section: "key", Entry: config.Entry{Key: "sign", Title: "S"}}}, with.Extra)
	assert.Equal(t, "key/sign", with.Extra[0].Path())
	assert.True(t, with.IsExtra("key/sign"))
	assert.False(t, with.IsExtra("key/edit"))
	assert.True(t, with.Valid("key/sign"))
	assert.True(t, with.Checkable("key/sign"))
	assert.Contains(t, with.Checks(), "key/sign")

	_, err = config.Parse([]byte(base + "extra:\n  nope:\n    read: R\n"))
	assert.ErrorContains(t, err, "no section")
	_, err = config.Parse([]byte(base + "extra:\n  key:\n    view: R\n"))
	assert.ErrorContains(t, err, "rule of every section")
}
