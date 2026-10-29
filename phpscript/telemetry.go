package phpscript

import (
	"github.com/titpetric/oida"
	"github.com/titpetric/phpscript/runner"
	"github.com/titpetric/phpscript/telemetry"
)

// Observe points a runtime's telemetry at the instance's own tracer, so one
// trace in /debug/oida holds the request and the PHP that ran under it.
// phpscript's telemetry package aliases the same oida, so the tracer passes
// through unwrapped. A nil tracer records nothing.
func Observe(rt *runner.Runtime, tracer *oida.Tracer) {
	if tracer == nil {
		return
	}
	rt.Observe(telemetry.NewModule(tracer))
}
