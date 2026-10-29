package phpscript_test

import (
	"context"
	"sync"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
)

// The pool is what makes a module safe to serve from more than one goroutine:
// one runtime answers one request at a time, and the interpreter keeps its
// symbols in plain maps. Two concurrent renders on one runtime is a concurrent
// map write and the process dies, so this is the test that would crash rather
// than fail.
func TestPoolConcurrent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	module(t, root, "titpetric/refinement", `<?php
$rows = array();
for ($i = 0; $i < 20; $i++) { $rows[] = "row" . $i; }
echo "rows=", count($rows), " uri=", $_SERVER["REQUEST_URI"];
`)

	opts := phpscript.NewOptions(root)
	opts.PoolSize = 4
	m := phpscript.New(opts)
	assert.NoError(t, m.Start(context.Background()))

	var wg sync.WaitGroup
	bodies := make([]string, 16)
	errs := make([]error, 16)
	for i := range bodies {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			out, err := render(t, m, "/m/titpetric/refinement/threads")
			bodies[i], errs[i] = string(out.Body), err
		}(i)
	}
	wg.Wait()

	for i := range bodies {
		assert.NoError(t, errs[i])
		assert.Equal(t, "rows=20 uri=/m/titpetric/refinement/threads", bodies[i])
	}
}

// A pool is per module, so one module's runtime never answers for another's
// files: the root a runtime reads is fixed when it is built.
func TestPoolPerModule(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	module(t, root, "titpetric/one", `<?php echo file_get_contents("/own.txt");`)
	module(t, root, "titpetric/two", `<?php echo file_get_contents("/own.txt");`)
	write(t, root+"/titpetric/one/own.txt", "first")
	write(t, root+"/titpetric/two/own.txt", "second")

	m := phpscript.New(phpscript.NewOptions(root))
	assert.NoError(t, m.Start(context.Background()))

	one, err := render(t, m, "/m/titpetric/one")
	assert.NoError(t, err)
	assert.Equal(t, "first", string(one.Body))

	two, err := render(t, m, "/m/titpetric/two")
	assert.NoError(t, err)
	assert.Equal(t, "second", string(two.Body))
}
