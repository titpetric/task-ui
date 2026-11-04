package phpscript

import (
	"context"

	"github.com/titpetric/oida"
	"github.com/titpetric/phpscript/runner"
	"github.com/titpetric/phpscript/telemetry"
)

// Observer is the telemetry module as a runtime option takes it.
//
// An observer is what makes the interpreter trace at all: Runtime.trace answers
// a noop span while the runtime has none, so a PHP method call, a new, an
// include and a template render are measured only once one is attached. A nil
// tracer answers a nil interface rather than a typed nil, which both rt.Observe
// and annotations.WithObservers skip.
func Observer(tracer *oida.Tracer) runner.Observer {
	if tracer == nil {
		return nil
	}
	return telemetry.NewModule(tracer)
}

// Observe points a runtime's telemetry at the instance's own tracer, so one
// trace in /debug/oida holds the request and the PHP that ran under it.
// phpscript's telemetry package aliases the same oida, so the tracer passes
// through unwrapped. A nil tracer records nothing.
func Observe(rt *runner.Runtime, tracer *oida.Tracer) {
	observer := Observer(tracer)
	if observer == nil {
		return
	}
	rt.Observe(observer)
}

// Span opens a span over the Go a PHP statement called, sourced at that
// statement.
//
// The interpreter's four trace sites all take a declaration, so they measure
// what PHP declares and nothing else: a function registered from Go is not a
// PHP frame, and a binding that walks a catalogue or reads the runner spends
// its time where no span is. A host binding opens its own, and this is it.
//
// A context with no trace under it answers a nil span, and every method of one
// does nothing, so the caller defers End unconditionally.
func Span(ctx context.Context, name string) *oida.Span {
	if ctx == nil {
		return nil
	}
	span := oida.StartSpan(ctx, name, oida.KindInternal)
	// The same attribution the database spans carry: the interpreter publishes
	// the file and line of the statement it ran into the binding's context.
	if filename, line := telemetry.SpanSource(ctx); filename != "" || line > 0 {
		span.SetSource(filename, line)
	}
	return span
}
