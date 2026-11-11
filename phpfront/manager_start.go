package phpfront

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

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
		module, err := startModule(ctx, m.opts, entry)
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

// startModule runs one modinfo.php on an unpooled runtime and returns what Task::register recorded.
func startModule(ctx context.Context, opts ModuleOptions, entry Found) (*Module, error) {
	own := &Module{Root: entry.Root}

	var registered *Module
	rt := runner.New(io.Discard, runtimeOptions(opts, own))
	stdlib.Register(rt)
	stdlib.RegisterFS(rt, entry.Root)
	BindModule(rt, opts, func(module *Module) { registered = module })
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

	// The module's other startup files, on the same runtime and with the
	// module it just registered on the context, so migrate.php migrates
	// the module that registered and not the directory it was found in.
	rt.SetContext(WithModule(ctx, registered))
	files, err := Startups(entry.Root)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		rel, err := filepath.Rel(entry.Root, file)
		if err != nil {
			return nil, err
		}
		if err := runStartup(rt, filepath.ToSlash(rel)); err != nil {
			return nil, fmt.Errorf("%s: %w", rel, err)
		}
	}
	return registered, nil
}

// runStartup loads and runs one of a module's startup files, counting exit as an end and not a failure.
func runStartup(rt *runner.Runtime, rel string) error {
	prog, err := rt.LoadFile(rel)
	if err != nil {
		return err
	}
	if err := rt.Run(prog); err != nil {
		if _, exited := runner.IsExit(err); !exited {
			return err
		}
	}
	return nil
}

// Nothing here prepares a compile directory. A module's templates compile
// into its own templates_c, which it ships precompiled and MiniTPL's compiler
// creates with its own recursive mkdir when a template is newer than what is
// there. The module resolves the path, so Go holds no setting for it.
