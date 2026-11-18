package phpfront

import (
	"context"

	"github.com/titpetric/phpscript/runner"
)

// registerModule registers the two symbols a modinfo.php of the tree calls.
func registerModule(rt *runner.Runtime) {
	rt.RegisterConstructor("Task\\Module", NewModule)

	rt.RegisterFunc("Task::register", func(m *Module) (bool, error) {
		if m == nil || m.Name() == "" {
			return false, ErrName
		}
		return true, nil
	})
}

// registerSearch registers Task::reindex for a front end startup job.
func registerSearch(rt *runner.Runtime, search func(ctx context.Context) error) {
	rt.RegisterFunc("Task::reindex", func(ctx context.Context) (bool, error) {
		if search == nil {
			return false, nil
		}
		return true, search(ctx)
	})
}
