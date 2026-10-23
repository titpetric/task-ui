package envfile_test

import (
	"testing"

	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestSafeRelative(t *testing.T) {
	t.Parallel()
	rel, err := envfile.SafeRelative(" conf/./x.env ")
	assert.NoError(t, err)
	assert.Equal(t, "conf/x.env", rel)
	for _, bad := range []string{"", "/etc/env", "../x.env", "$HOME/.env", "."} {
		_, err := envfile.SafeRelative(bad)
		assert.Error(t, err, bad)
	}
}

func TestStripGenerated(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "A=1\n", envfile.StripGenerated("A=1\n\n"+envfile.Marker+"\nX=y\n"))
	assert.Equal(t, "A=1\n", envfile.StripGenerated("A=1\n"))
}

func TestRender(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", envfile.Render("", nil))
	assert.Equal(t, "A=1\n", envfile.Render("A=1", nil))
	assert.Equal(t, envfile.Marker+"\nX=y\n", envfile.Render("", map[string]string{"X": "y"}))
	assert.Equal(t, "A=1\n\n"+envfile.Marker+"\nX=y\n", envfile.Render("A=1\n", map[string]string{"X": "y"}))
}
