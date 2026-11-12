package phpfront

import (
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
