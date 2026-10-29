package phpscript

import (
	"io"
	"os"

	"github.com/titpetric/phpscript/runner"
	"github.com/titpetric/phpscript/stdlib"
)

// pool is one module's registered runtimes, kept between the requests they
// serve. Registration costs more than a request does, so a pooled runtime
// pays it once and takes a reset per request instead.
//
// One runtime serves one goroutine at a time. The interpreter keeps its
// symbols in plain maps, so two goroutines on one runtime is a concurrent map
// write and the process dies; the pool is what makes that impossible.
type pool struct {
	module   *Module
	sessions *runner.SessionPool

	// includes and exprs are the parsed programs and expressions, shared by
	// every runtime of this module: the parse is the same whichever runtime
	// runs it, and sharing them is why a second request is cheap.
	includes *runner.IncludeCache
	exprs    *runner.ExprCache
}

// newPool returns the pool of one module. Nothing is built yet: the first Get
// constructs a runtime, and the pool grows to PoolSize under load.
func newPool(opts Options, m *Module) *pool {
	p := &pool{
		module:   m,
		includes: runner.NewIncludeCache(),
		exprs:    runner.NewExprCache(),
	}
	p.sessions = runner.NewSessionPool(opts.PoolSize, func() *runner.Runtime {
		return p.build(opts)
	})
	return p
}

// build constructs one runtime of this module's pool: the standard library,
// the module's own directory as the filesystem it reads and writes, the host
// bindings, the shared caches and the telemetry observer. FreezeStdlib marks
// what the host registered, so the reset between requests keeps the bindings
// and drops what a script defined.
func (p *pool) build(opts Options) *runner.Runtime {
	rt := runner.New(io.Discard, runtimeOptions(p.module))
	rt.SetIncludeCache(p.includes)
	rt.SetExprCache(p.exprs)
	stdlib.Register(rt)
	stdlib.RegisterFS(rt, p.module.Root)
	Observe(rt, opts.Tracer)
	// A serving runtime registers no module: a modinfo.php runs once, at
	// startup, on a runtime of its own.
	Bind(rt, opts, nil)
	rt.FreezeStdlib()
	return rt
}

// get takes a runtime of the pool, and the function that returns it. The
// caller defers the second.
func (p *pool) get() (*runner.Runtime, func()) {
	rt := p.sessions.Get()
	return rt, func() { p.sessions.Put(rt) }
}

// runtimeOptions are the interpreter's options for one module: its own
// directory as the root every include and every entrypoint resolves against,
// and the working directory a relative path is read from.
//
// TimeLimit and ConcurrencyLimit are left alone. The interpreter holds both
// and enforces neither, so the Manager bounds a render with a context deadline
// and the pool bounds how many run at once.
func runtimeOptions(m *Module) runner.Options {
	return runner.Options{
		RootFS:  os.DirFS(m.Root),
		WorkDir: "/",
		SAPI:    "phpscript",
	}
}
