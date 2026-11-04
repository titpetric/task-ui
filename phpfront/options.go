package phpfront

import (
	"runtime"
	"time"

	"github.com/titpetric/oida"

	"github.com/titpetric/task-ui/phpscript"
)

// DefaultTimeout bounds one page; phpscript holds a TimeLimit and enforces none.
const DefaultTimeout = 30 * time.Second

// Options configures the front end. Take NewOptions and override.
type Options struct {
	// Root is the application tree, as Extract left it.
	Root string

	// Base is where the router is mounted, bound as Task::base(). The one
	// place the prefix is spelled; rules/ refuses the literal.
	Base string

	// Database resolves `new Database("task-ui")`.
	Database *phpscript.Database

	// Artifacts is the artifact tree.
	Artifacts *phpscript.Tree

	// Host answers what is not a row. Nil answers every such binding empty.
	Host Host

	// Tracer records the spans into the instance's traces.
	Tracer *oida.Tracer

	// Version is the instance's version, bound as Task::version().
	Version string

	// Login reports whether the instance requires a sign in, bound as
	// Task::login(). It is the Guard's, which mints and reads the session
	// outside the base, so the header asks rather than deciding: with auto
	// login on there is nothing to log out of.
	Login bool

	// Now is the instance's clock, bound as Task::now(). It is the one the
	// hours gate and the Go pages read, which TASK_UI_NOW pins for the
	// screenshots: a page reading the process clock instead would draw a
	// window of thirty days ending on another day than the page it is
	// compared against. Nil is time.Now.
	Now func() time.Time

	// PoolSize bounds how many pages render at once: one runtime, one goroutine.
	PoolSize int

	// Timeout bounds one page.
	Timeout time.Duration

	// OnError receives a refused route, a duplicate, a tree it could not read.
	OnError func(error)
}

// NewOptions returns the options of an instance serving the tree at root.
func NewOptions(root string) Options {
	return Options{
		Root:     root,
		PoolSize: max(2, runtime.NumCPU()),
		Timeout:  DefaultTimeout,
	}
}

// report hands err to OnError, if there is one.
func (o Options) report(err error) {
	if err != nil && o.OnError != nil {
		o.OnError(err)
	}
}
