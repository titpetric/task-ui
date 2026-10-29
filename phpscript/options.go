package phpscript

import (
	"runtime"
	"time"

	"github.com/titpetric/oida"
)

// DefaultTimeout bounds one render. phpscript holds a TimeLimit option and
// enforces none of it, so the bound is kept here: a module that loops forever
// ends with the page and leaves the instance alone.
const DefaultTimeout = 30 * time.Second

// Options configures the Manager. Take NewOptions and override what the
// instance needs, so a field added later keeps its default.
type Options struct {
	// Roots are the directories walked for a modinfo.php, in order. The first
	// is the instance's own modules directory, where the embedded tree is
	// extracted; a workspace root after it reaches a module in a checkout.
	Roots []string

	// Registry records what a modinfo.php registered. Nil runs the modules
	// and keeps them in memory alone, as a test does.
	Registry Registry

	// Migrator applies a module's schema. Nil makes Task::migrate refuse.
	Migrator Migrator

	// Dispatcher starts the Claude runs. Nil makes Task::dispatch refuse.
	Dispatcher Dispatcher

	// Tracer records the spans of the PHP into the same traces as the
	// instance's own middleware. Nil records nothing.
	Tracer *oida.Tracer

	// PoolSize is how many registered runtimes a module keeps between
	// requests. One runtime serves one goroutine at a time, so this is also
	// the cheapest bound on how many renders run at once.
	PoolSize int

	// Timeout bounds one render.
	Timeout time.Duration

	// Version is the instance's version, which PHP reads as Task::VERSION.
	Version string

	// OnError receives what the Manager cannot report through a return: a
	// modinfo.php that threw, a module that registered nothing, a template
	// directory it could not seed. Nil drops them.
	OnError func(error)
}

// NewOptions returns the options of an instance serving the modules under
// root.
func NewOptions(root string) Options {
	return Options{
		Roots:    []string{root},
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
