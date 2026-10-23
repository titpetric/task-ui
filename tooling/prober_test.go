package tooling_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/tooling"
)

func TestProber_Bin(t *testing.T) {
	var none *tooling.Prober
	assert.Equal(t, "atkins", none.Bin())
	assert.Equal(t, "atkins", (&tooling.Prober{}).Bin())
	assert.Equal(t, "/usr/local/bin/atkins", (&tooling.Prober{Command: "/usr/local/bin/atkins"}).Bin())
}
