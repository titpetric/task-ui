package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// The detail words the uptime and why a cell is empty.
func TestHostDetail(t *testing.T) {
	t.Parallel()
	d := &view.HostDetail{System: beszel.System{Info: beszel.Info{Uptime: 3 * 86400}}, Unsized: "no login"}
	assert.Equal(t, "3 days", d.Uptime())
	assert.Equal(t, "no login", d.SizeNote())
	assert.Equal(t, "Beszel reports no containers of the host", d.MeasureNote())
	d = &view.HostDetail{System: beszel.System{Info: beszel.Info{Uptime: 5 * 3600}}, Sized: true, Measured: true}
	assert.Equal(t, "5 hours", d.Uptime())
	assert.Equal(t, "docker ps does not list the container", d.SizeNote())
	assert.Equal(t, "Beszel reports no measures of the container", d.MeasureNote())
	d.System.Info.Uptime = 600
	assert.Equal(t, "10 minutes", d.Uptime())
}
