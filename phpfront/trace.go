package phpfront

import (
	"context"
	"reflect"

	"github.com/titpetric/phpscript/runner"

	"github.com/titpetric/task-ui/phpscript"
)

// traced is the runtime the host bindings register onto: the same RegisterFunc,
// with a span around every call.
//
// The interpreter's own trace sites all take a PHP declaration, so a method
// call, a new, an include and a template are measured and a function registered
// from Go is not. That is where a ported page spends its time: the queries show
// up because the database stdlib opens its own spans, Task\Host::folders walks
// every checkout and shows up as a gap between two of them. One wrapper here
// names all fifty-five of them at once.
type traced struct {
	rt *runner.Runtime
}

// RegisterFunc registers fn under name, measured.
func (t traced) RegisterFunc(name string, fn any) {
	t.rt.RegisterFunc(name, measured(name, fn))
}

// ctxType is the interface a binding's first parameter has when the runtime
// injects the request context into it.
var ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()

// measured returns fn wrapped in a span named after the binding and sourced at
// the PHP statement that called it. The wrapper is made over fn's own type, so
// the runtime builds the same invoker it would have built for fn itself.
//
// A binding that takes no context is returned unchanged: there is no trace to
// hang a span on, and the ones in that group answer a configured string. So is
// a variadic one, which would need CallSlice; none of the bindings is.
func measured(name string, fn any) any {
	value := reflect.ValueOf(fn)
	signature := value.Type()
	if signature.NumIn() == 0 || signature.In(0) != ctxType || signature.IsVariadic() {
		return fn
	}
	call := func(in []reflect.Value) []reflect.Value {
		ctx, _ := in[0].Interface().(context.Context)
		defer phpscript.Span(ctx, name).End()
		return value.Call(in)
	}
	wrapper := reflect.MakeFunc(signature, call)
	return wrapper.Interface()
}
