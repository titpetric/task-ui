package phpfront

import (
	"context"
	"strings"

	"github.com/titpetric/phpscript/model"
	"github.com/titpetric/phpscript/runner"

	"github.com/titpetric/task-ui/internal/assets"
	"github.com/titpetric/task-ui/markdown"
	"github.com/titpetric/task-ui/phpscript"
)

// The calls are functions and not constants: the handler freezes the standard
// library when it builds a runtime and the reset between requests drops a
// constant registered after that, while the host function table survives.
//
// The name is namespaced. A root class needs a leading backslash from inside
// the tree's own Task namespace, and `phpscript fmt` strips one off a
// single-segment name, which repoints the call at Task\Task. See
// docs/todo-correction.md.
//
// Task\Host::publish is the only call that changes anything. The writes go
// through code/Task over the instance's database, and the spelling is the
// contract: Task\Host::x() is the host, every other Task\ name is not.

// HostClass is the class the host calls hang off.
const HostClass = `Task\Host`

// Bind registers the host calls on a runtime serving one request.
func Bind(rt *runner.Runtime, opts Options, static *assets.Tree) {
	// A call, not a constant, so dropping the prefix is one Go line.
	rt.RegisterFunc(HostClass+"::base", func() (string, error) {
		return opts.Base, nil
	})

	rt.RegisterFunc(HostClass+"::version", func() (string, error) {
		return opts.Version, nil
	})

	// The static files sit at the root, outside the base, so an asset URL
	// survives the cutover unchanged. The version is their content hash,
	// which the URLs carry as ?v= so an hour of caching never outlives a
	// deploy.
	rt.RegisterFunc(HostClass+"::assets", func() (string, error) {
		return strings.TrimSuffix(assets.Prefix, "/"), nil
	})
	rt.RegisterFunc(HostClass+"::assetVersion", func() (string, error) {
		return static.Version(), nil
	})

	// The name the Guard resolved; the row behind it is Task\User.
	rt.RegisterFunc(HostClass+"::whoami", func(ctx context.Context) (string, error) {
		if opts.Host == nil {
			return "", nil
		}
		return opts.Host.Whoami(ctx), nil
	})

	// The catalogue, which nothing records: it is read off the checkouts.
	rt.RegisterFunc(HostClass+"::folders", func(ctx context.Context) (*model.Array, error) {
		if opts.Host == nil {
			return phpList(), nil
		}
		return folders(opts.Host.Folders(ctx)), nil
	})

	// The runner's memory, where a run lives until it ends.
	rt.RegisterFunc(HostClass+"::running", func(ctx context.Context) (*model.Array, error) {
		if opts.Host == nil {
			return phpList(), nil
		}
		return running(opts.Host.Running(ctx)), nil
	})

	// git's answer, not a row.
	rt.RegisterFunc(HostClass+"::dirty", func(ctx context.Context, dir string) (int, error) {
		if opts.Host == nil {
			return 0, nil
		}
		return opts.Host.Dirty(ctx, dir), nil
	})

	// config/permissions.yml. Task\Permission evaluates it over the policy rows.
	rt.RegisterFunc(HostClass+"::rules", func(ctx context.Context) (*model.Array, error) {
		if opts.Host == nil {
			return rules(Rules{}), nil
		}
		return rules(opts.Host.Rules(ctx)), nil
	})

	// A write that does not publish leaves every open page stale. The payload
	// is declared as the runtime's own array: coerceArg maps no PHP array onto
	// a Go map, so a map parameter refuses every call.
	rt.RegisterFunc(HostClass+"::publish", func(ctx context.Context, topic string, payload *model.Array) (bool, error) {
		if opts.Host == nil {
			return false, nil
		}
		fields := map[string]any{}
		if payload != nil {
			fields = payload.Map()
		}
		if err := opts.Host.Publish(ctx, topic, fields); err != nil {
			return false, err
		}
		return true, nil
	})

	// The renderer that owns the content security policy.
	rt.RegisterFunc(HostClass+"::markdown", func(src, folder string) (string, error) {
		return markdown.HTML([]byte(src), phpscript.ArtifactHref(folder), ""), nil
	})

	// A note's text, which is the same renderer with a line break kept and
	// @name marked. A name in link becomes a link under prefix, and every
	// other mention is marked and not linked.
	rt.RegisterFunc(HostClass+"::message", func(src, prefix string, link *model.Array) (string, error) {
		return Message(src, prefix, names(link)), nil
	})

	// The artifact tree, which sits outside the directory the PHP may open.
	rt.RegisterFunc(HostClass+"::artifacts", func(folder string) ([]string, error) {
		return opts.Artifacts.Files(folder), nil
	})
	rt.RegisterFunc(HostClass+"::artifact", func(rel string) (string, error) {
		return opts.Artifacts.Read(rel), nil
	})

	rt.RegisterClass(&model.Class{Name: HostClass})
}

// database is what `new Database("task-ui")` resolves through.
func database(opts Options) model.DatabaseProvider {
	if opts.Database == nil {
		return nil
	}
	return opts.Database
}
