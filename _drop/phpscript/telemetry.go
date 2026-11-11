package phpscript

import (
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
