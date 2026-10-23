package app_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestNewUsersFile(t *testing.T) {
	t.Parallel()

	f := app.NewUsersFile("/etc/task-ui/.htpasswd")
	assert.Equal(t, "/etc/task-ui/.htpasswd", f.Path)
	assert.Equal(t, 5*time.Second, f.Interval)
	assert.False(t, f.Changed())
}
