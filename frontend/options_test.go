package phpfront_test

import (
	"runtime"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpfront"
)

// The defaults are the ones an instance serving one tree wants, so a field
// added later keeps its own.
func TestNewOptions(t *testing.T) {
	t.Parallel()

	opts := phpfront.NewOptions("app")
	assert.Equal(t, "app", opts.Root)
	assert.Equal(t, phpfront.DefaultTimeout, opts.Timeout)
	assert.Equal(t, max(2, runtime.NumCPU()), opts.PoolSize)
	assert.Equal(t, "", opts.Base, "nothing is mounted under a prefix by default")
}

// A root is the one thing the front end cannot be built without.
func TestNew_NoRoot(t *testing.T) {
	t.Parallel()

	_, err := phpfront.New(phpfront.Options{})
	assert.Error(t, err)
}

// A zero timeout and a zero pool are filled, so an Options built by hand works.
func TestNew_Fills(t *testing.T) {
	t.Parallel()

	front, err := phpfront.New(phpfront.Options{Root: "testdata/routes"})
	assert.NoError(t, err)
	assert.NotNil(t, front)
}
