package phpscript

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/titpetric/phpscript/runner"
	"github.com/titpetric/phpscript/stdlib"
)

// Start walks the roots, runs every modinfo.php once and records what each
// registered. It is the whole of a module's registration: nothing else runs a
// modinfo.php, and a module that did not register here does not serve.
//
// A module that throws, registers nothing or names itself wrongly is reported
// through OnError and left out; the modules that did register serve. A
// modinfo.php the server cannot read is the one failure that comes back,
// because an instance that cannot read a file it found is misconfigured rather
// than empty. A root that is not there, or that holds no module, is neither.
func (m *Manager) Start(ctx context.Context) error {
	if m == nil {
		return nil
	}
	found, err := Discover(m.opts.Roots...)
	if err != nil {
		return err
	}
	var names []string
	for _, entry := range found {
		if !entry.Startup {
			m.opts.report(fmt.Errorf("%s: no @startup annotation, so the module is not registered", entry.File))
			continue
		}
		module, err := m.startup(ctx, entry)
		if err != nil {
			m.opts.report(fmt.Errorf("%s: %w", entry.File, err))
			continue
		}
		m.put(module)
		names = append(names, module.Name())
		if m.opts.Registry != nil {
			if err := m.opts.Registry.Register(ctx, module); err != nil {
				m.opts.report(fmt.Errorf("%s: %w", module.Name(), err))
			}
		}
	}
	if m.opts.Registry != nil {
		if _, err := m.opts.Registry.Forget(ctx, names); err != nil {
			m.opts.report(err)
		}
	}
	return nil
}

// Reload drops every module and reads the roots again, so a module added to a
// checkout appears without a restart.
func (m *Manager) Reload(ctx context.Context) error {
	if m == nil {
		return nil
	}
	m.reset()
	return m.Start(ctx)
}

// startup runs one modinfo.php on a runtime of its own and returns what
// Task::register recorded.
//
// The runtime is not pooled. It runs once, at startup, and a module's
// registration is not a request: a fresh runtime holds nothing of another
// module's, and nothing keeps it alive afterwards.
func (m *Manager) startup(ctx context.Context, entry Found) (*Module, error) {
	own := &Module{Root: entry.Root}

	var registered *Module
	rt := runner.New(io.Discard, runtimeOptions(own))
	stdlib.Register(rt)
	stdlib.RegisterFS(rt, entry.Root)
	Bind(rt, m.opts, func(module *Module) { registered = module })
	rt.SetContext(WithModule(ctx, own))

	prog, err := rt.LoadFile(ModInfo)
	if err != nil {
		return nil, err
	}
	if err := rt.Run(prog); err != nil {
		if _, exited := runner.IsExit(err); !exited {
			return nil, err
		}
	}
	if registered == nil {
		return nil, errors.New("the script called no Task::register")
	}
	return registered, nil
}

// Nothing here prepares a compile directory. A module's templates compile
// into its own templates_c, which it ships precompiled and MiniTPL's compiler
// creates with its own recursive mkdir when a template is newer than what is
// there. The module resolves the path, so Go holds no setting for it.
