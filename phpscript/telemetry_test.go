package phpscript_test

import (
	"io"
	"os"
	"testing"

	"github.com/titpetric/oida"
	"github.com/titpetric/phpscript/runner"
	"github.com/titpetric/phpscript/stdlib"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
)

func TestObserve(t *testing.T) {
	t.Parallel()
	rt := runner.New(io.Discard, runner.Options{RootFS: os.DirFS(t.TempDir())})
	stdlib.Register(rt)

	// A nil tracer records nothing and changes no behaviour: start_span is
	// still there and the span it returns still ends.
	phpscript.Observe(rt, nil)
	prog, err := rt.Load(`<?php $span = start_span("load", "database"); $span->end(); echo "done";`)
	assert.NoError(t, err)
	assert.NoError(t, rt.Run(prog))

	opts := oida.NewOptions("task-ui-test")
	opts.ReadEnv = false
	opts.Enabled = true
	tracer, err := oida.New(opts)
	assert.NoError(t, err)

	phpscript.Observe(rt, tracer)
	assert.True(t, tracer.Enabled(), "the tracer the PHP reports into is the instance's own")
}
