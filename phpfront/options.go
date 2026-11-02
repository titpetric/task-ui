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
