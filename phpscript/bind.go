package phpscript

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/phpscript/model"
	"github.com/titpetric/phpscript/runner"
)

// ErrNoModule is returned when a binding runs on a runtime the Manager did not
// set a module on.
var ErrNoModule = errors.New("no module on the context")

// ErrEscapes is returned when a module names a path outside its own root.
var ErrEscapes = errors.New("the path leaves the module root")

// Bind registers the four host symbols of a module's runtime: the
// `Task\Module` constructor, and `Task::register`, `Task::migrate` and
// `Task::dispatch` as host functions whose names hold `::`. registered takes
// what a modinfo.php registers, and nil refuses the call.
//
// The context.Context each binding takes first is the interpreter's and names
// the module; PHP counts it as no argument. docs/modules.md is the contract
// and docs/todo-correction.md says why the constructor is not `Task::Module`.
func Bind(rt *runner.Runtime, opts Options, registered func(*Module)) {
	// A constructor publishes every exported method of the value it returns,
	// so setName and set_name both reach Module.SetName.
	rt.RegisterConstructor("Task\\Module", NewModule)

	rt.RegisterFunc("Task::register", func(ctx context.Context, m *Module) (bool, error) {
		own, ok := ModuleFrom(ctx)
		if !ok {
			return false, ErrNoModule
		}
		if m == nil {
			return false, errors.New("Task::register takes a Task\\Module")
		}
		// The root is the directory the modinfo.php sits in, never what the
		// script says, so a module cannot claim another's files.
		m.Root = own.Root
		if err := m.Validate(); err != nil {
			return false, err
		}
		if registered != nil {
			registered(m)
		}
		return true, nil
	})

	rt.RegisterFunc("Task::migrate", func(ctx context.Context, dir string) (bool, error) {
		m, ok := ModuleFrom(ctx)
		if !ok {
			return false, ErrNoModule
		}
		if opts.Migrator == nil {
			return false, errors.New("this instance applies no module schema")
		}
		rel, err := Inside(m.Root, dir)
		if err != nil {
			return false, err
		}
		if err := opts.Migrator.Migrate(ctx, m, rel); err != nil {
			return false, err
		}
		return true, nil
	})

	rt.RegisterFunc("Task::dispatch", func(ctx context.Context, comment, action string) (string, error) {
		m, ok := ModuleFrom(ctx)
		if !ok {
			return "", ErrNoModule
		}
		if opts.Dispatcher == nil {
			return "", errors.New("this instance starts no module runs")
		}
		if strings.TrimSpace(comment) == "" {
			return "", errors.New("Task::dispatch takes the comment to act on")
		}
		rel, err := Inside(m.Root, action)
		if err != nil {
			return "", err
		}
		if _, err := os.Stat(filepath.Join(m.Root, rel)); err != nil {
			return "", fmt.Errorf("%s: %w", action, err)
		}
		return opts.Dispatcher.Dispatch(ctx, m, comment, rel)
	})

	rt.RegisterClass(taskClass(opts))
}

// taskClass is the host class the constants live on. RegisterClass takes PHP
// declarations alone, so this declares no method: the three static calls are
// host functions of the same spelling.
func taskClass(opts Options) *model.Class {
	return &model.Class{
		Name: "Task",
		Consts: []model.Field{
			{Name: "VERSION", Default: &model.Lit{Value: opts.Version}},
			{Name: "PREFIX", Default: &model.Lit{Value: Prefix}},
		},
	}
}

// Inside returns path as a path relative to root, and refuses one that leaves
// it. A module reads its own directory: a `..` segment, an absolute path and a
// symlink pointing out all come back as an error.
func Inside(root, path string) (string, error) {
	if root == "" {
		return "", ErrNoModule
	}
	clean := filepath.Clean(strings.TrimPrefix(filepath.ToSlash(path), "/"))
	if clean == "." {
		return ".", nil
	}
	full := filepath.Join(root, clean)
	rel, err := filepath.Rel(root, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: %q", ErrEscapes, path)
	}
	// A symlink out of the root resolves to a path above it, which the same
	// comparison catches. A path that is not there yet resolves to itself.
	if real, err := filepath.EvalSymlinks(full); err == nil {
		base, err := filepath.EvalSymlinks(root)
		if err != nil {
			return "", err
		}
		out, err := filepath.Rel(base, real)
		if err != nil || out == ".." || strings.HasPrefix(out, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("%w: %q", ErrEscapes, path)
		}
	}
	return rel, nil
}
