package phpfront

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	chi "github.com/go-chi/chi/v5"

	"github.com/titpetric/phpscript/annotations"
	"github.com/titpetric/phpscript/runner"

	"github.com/titpetric/task-ui/internal/assets"
)

// Front serves the application tree.
type Front struct {
	opts   Options
	routes *annotations.Route
	static *assets.Tree
}

// New returns the front end over opts, its routes discovered once here.
func New(opts Options) (*Front, error) {
	if opts.Root == "" {
		return nil, errors.New("phpfront: no application root")
	}
	if opts.Timeout <= 0 {
		opts.Timeout = DefaultTimeout
	}
	if opts.PoolSize <= 0 {
		opts.PoolSize = 2
	}
	static := assets.New(assets.Dirs(opts.Root))
	return &Front{opts: opts, routes: route(static, opts), static: static}, nil
}

// Assets returns the static tree, mounted at the root beside the pages, so an
// asset URL survives the cutover unchanged.
func (f *Front) Assets() *assets.Tree {
	if f == nil {
		return nil
	}
	return f.static
}

// route builds the annotation scanner over the tree.
func route(static *assets.Tree, opts Options) *annotations.Route {
	return annotations.NewRoute(os.DirFS(opts.Root),
		// A real directory, which composer's autoloader and MiniTPL both need.
		annotations.WithRootDir(opts.Root),
		annotations.WithRunnerOptions(runner.Options{
			WorkDir:  "/",
			Database: database(opts),
		}),
		// Per request, after the request context sits on the runtime.
		annotations.WithRuntimeFunc(func(rt *runner.Runtime) { Bind(rt, opts, static) }),
		// The scanner excludes vendor; composer's install directory is ext.
		annotations.WithExcludedDirectory("ext"),
		annotations.WithModuleSuffix("taskui"),
	)
}

// Mount registers every discovered route on the router.
func (f *Front) Mount(ctx context.Context, router chi.Router) error {
	if f == nil {
		return nil
	}
	return f.routes.Mount(ctx, router)
}

// Routes returns a router of the front end's own.
func (f *Front) Routes(ctx context.Context) (http.Handler, error) {
	r := chi.NewRouter()
	if err := f.Mount(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Href is a link built in Go: the one place outside the tree that has to agree
// with Task\Host::base().
func (f *Front) Href(path string) string {
	if f == nil {
		return path
	}
	return f.opts.Base + "/" + strings.TrimPrefix(path, "/")
}
