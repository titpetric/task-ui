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
	"github.com/titpetric/task-ui/markdown"
	"github.com/titpetric/task-ui/phpscript"
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
			// The interpreter's own spans, marked as detail: a frame under
			// the recorder's minimum is dropped and counted, and the files
			// a page included arrive as one span at the end of it. debug
			// publishes a span per frame and per include instead, which is
			// a page being taken apart and not a page being served.
			TraceLevel: runner.TraceLevelInfo,
		}),
		// Per request, after the request context sits on the runtime.
		annotations.WithRuntimeFunc(func(rt *runner.Runtime) { Bind(rt, opts, static) }),
		// Once per pooled runtime, which is why it is not in the line above:
		// Observe appends and nothing clears the list, so a per-request
		// registration would publish every span once per request served.
		// Without it the interpreter traces nothing and a page's trace holds
		// the queries alone.
		annotations.WithObservers(phpscript.Observer(opts.Tracer)),
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

// Public is what the front end answers before a sign in: the login page, and
// nothing else.
//
// The Go router puts the same paths outside its Guard group rather than at the
// mount, for the reason a list of one makes plain — a sign in page behind a
// sign in redirects to itself, so nobody could ever reach it. The assets are
// outside too, by being mounted on their own. docs/phpfront.md names this the
// one documented exclusion from "every page".
var Public = []string{"/login"}

// Gate is the Guard as the front end mounts it: the check that keeps a reader
// without a session out, and the one that only names the reader there is.
type Gate interface {
	Require(http.Handler) http.Handler
	Attach(http.Handler) http.Handler
}

// Guarded is the front end behind the gate: Require for every page, Attach for
// the ones in Public, which answer before a sign in and still have to know who
// is reading.
//
// The Go router does this with a chi group and the public routes outside it.
// Everything under a mount is one handler, so this is that shape at the mount,
// matched on the path the base gives each public route — chi leaves the
// request's own URL whole and only the route context carries the remainder.
func (f *Front) Guarded(gate Gate, next http.Handler) http.Handler {
	if f == nil {
		return next
	}
	open := make(map[string]bool, len(Public))
	for _, path := range Public {
		open[f.Href(path)] = true
	}
	named := gate.Attach(next)
	guarded := gate.Require(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if open[r.URL.Path] {
			named.ServeHTTP(w, r)
			return
		}
		guarded.ServeHTTP(w, r)
	})
}

// Policy sets the content security policy of every page and asset this front
// end answers with, which is the one the Go pages carry.
//
// The Go router sets it for everything under it with one r.Use, so a page, a
// stylesheet and a refusal all carry it. Nothing in PHP can: a header set
// inside the request is the page's, and a response the Guard writes before the
// page runs would carry none. So it is a middleware at the mount, wrapped
// around the Guard rather than inside it, and the cutover keeps the line.
func Policy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", markdown.CSP)
		next.ServeHTTP(w, r)
	})
}

// Href is a link built in Go: the one place outside the tree that has to agree
// with Task\Host::base().
func (f *Front) Href(path string) string {
	if f == nil {
		return path
	}
	return f.opts.Base + "/" + strings.TrimPrefix(path, "/")
}
