package phpscript_test

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/titpetric/oida"
	"github.com/titpetric/phpscript/runner"
	"github.com/titpetric/phpscript/stdlib"
	"github.com/titpetric/phpscript/telemetry"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
)

// tracer is an enabled tracer of this test's own, which records every trace:
// the instance reads its options out of a setting row and a test has none.
func tracer(t *testing.T) *oida.Tracer {
	t.Helper()
	opts := oida.NewOptions("task-ui-test")
	opts.ReadEnv = false
	opts.Enabled = true
	tracer, err := oida.New(opts)
	assert.NoError(t, err)
	return tracer
}

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

	recorder := tracer(t)
	phpscript.Observe(rt, recorder)
	assert.True(t, recorder.Enabled(), "the tracer the PHP reports into is the instance's own")
}

// TestObserver holds the gate: an observer is what the interpreter checks
// before it traces anything, so a nil tracer has to answer a nil interface and
// not a typed nil, which would read as an observer and publish nowhere.
func TestObserver(t *testing.T) {
	t.Parallel()
	assert.Nil(t, phpscript.Observer(nil), "no tracer is no observer")
	assert.NotNil(t, phpscript.Observer(tracer(t)), "a tracer is an observer")
}

// TestObserverTracesPHP is the bug the ported front end had: the runtime it
// served from carried no observer, so a page's trace held the queries the
// database stdlib opens on its own and not one line of the PHP above them.
func TestObserverTracesPHP(t *testing.T) {
	t.Parallel()

	source := `<?php
class Loader {
	public function load() {
		return 1;
	}
}

$loader = new Loader();
$loader->load();
`

	rt := runner.New(io.Discard, runner.Options{RootFS: os.DirFS(t.TempDir())})
	stdlib.Register(rt)
	rt.Observe(phpscript.Observer(tracer(t)))

	program, err := rt.Load(source)
	assert.NoError(t, err)

	// The spans are read inside the trace: Observe returns the trace to the
	// pool as it ends, and the copy a reader takes is the snapshot's.
	var names []string
	run := func(ctx context.Context) error {
		rt.SetContext(ctx)
		if err := rt.Run(program); err != nil {
			return err
		}
		trace := telemetry.TraceFromContext(ctx).Clone()
		for _, span := range trace.Spans {
			names = append(names, span.Name)
		}
		return nil
	}
	assert.NoError(t, tracer(t).Observe(context.Background(), "GET /runs", run))

	assert.Contains(t, names, "new Loader", "the allocation of a class is a span")
	// The interpreter names an instance call after the receiver it was
	// written on, not the class: the span of a page reads as the line did.
	assert.Contains(t, names, "loader.load", "a method call is a span")
}

// TestSpan holds what a host binding records. The interpreter traces what PHP
// declares, so a binding is a gap in the page's trace unless it opens its own,
// and the source it carries is the PHP statement that called it.
func TestSpan(t *testing.T) {
	t.Parallel()

	// No trace is a nil span, and a binding defers End on one unconditionally.
	phpscript.Span(context.Background(), "Task\\Host::folders").End()

	run := func(ctx context.Context) error {
		sourced := telemetry.WithSpanLine(telemetry.WithSpanFilename(ctx, "/modules/runs/mod_runs.php"), 37)
		span := phpscript.Span(sourced, "Task\\Host::folders")
		assert.Equal(t, "Task\\Host::folders", span.Name)
		assert.Equal(t, "/modules/runs/mod_runs.php", span.Filename)
		assert.Equal(t, 37, span.Line)
		assert.Equal(t, oida.KindInternal, span.Kind)
		span.End()
		return nil
	}
	assert.NoError(t, tracer(t).Observe(context.Background(), "GET /runs", run))
}
