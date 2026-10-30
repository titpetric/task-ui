// Package frontend serves the pages of task-ui: the sign in screen, the
// dashboard, the folders and their jobs, the runs, the artifacts, the
// schedules and the settings. Pages read; the API makes every change,
// and the pages call it from the browser.
//
// The guard sits in front of every page but the sign in screen and the
// assets, so a handler always finds its user on the context.
package frontend

import (
	"net/http"

	chi "github.com/go-chi/chi/v5"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/markdown"
)

// Handlers serves the pages.
type Handlers struct {
	opts Options
	view *view.View
	// subs holds the directories of Files lists, as detected.
	subs *subCache
	// commits holds the time of each project's last commit, for the
	// sidebar of a workspace that lists its newest projects.
	commits *LastCommits
}

// New returns the handlers over the options, with the templates parsed.
func New(opts Options) (*Handlers, error) {
	v, err := view.New()
	if err != nil {
		return nil, err
	}
	h := &Handlers{opts: opts, view: v, subs: &subCache{}, commits: &LastCommits{}}
	if opts.Git != nil {
		h.commits.Read = opts.Git.LastCommit
	}
	return h, nil
}

// Routes returns the router of the pages, to mount at the root, every response under markdown.CSP.
func (h *Handlers) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(policy)

	r.Handle(view.AssetPrefix+"*", assets())
	r.Get("/login", h.loginForm)
	r.With(access.SameOrigin).Post("/login", h.login)
	r.With(access.SameOrigin).Post("/logout", h.logout)
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok\n")) })

	r.Group(func(r chi.Router) {
		r.Use(h.opts.Guard.Require)
		r.Get("/", h.home)
		r.Get("/f/*", h.folderPath)
		r.Get("/workflows", h.workflows)
		r.Get("/workflows/new", h.workflowNew)
		r.Get("/workflows/{id}", h.workflow)
		r.Get("/workflows/{id}/edit", h.workflowEdit)
		r.Get("/workflows/{id}/runs/{run}", h.workflowRun)
		r.Get("/runs", h.runs)
		r.Get("/runs/{id}", h.run)
		r.Get("/runs/{id}/report", h.report)
		r.Get("/artifacts", h.artifacts)
		r.Get("/artifacts/files/*", h.artifactFile)
		r.Get("/files/*", h.checkoutFile)
		r.Get("/artifacts.tgz", h.artifactArchive)
		r.Get("/artifacts/*", h.artifactArchive)
		r.Get("/schedules", h.schedules)
		r.Get("/notes", h.notes)
		r.Get("/inventory", h.inventory)
		r.Get("/inventory/hosts/{name}", h.inventoryHost)
		r.Get("/settings", h.settings)
		r.Get("/settings/instance", h.instance)
		r.Get("/settings/retention", h.dataRetention)
		r.Get("/settings/telemetry", h.telemetry)
		r.Get("/settings/sidebar", h.sidebarSettings)
		// A module's pages, inside the guard, so a plugin inherits the
		// sign in, the content policy and the layout rather than carrying
		// any of them. Two patterns: the bare link and everything under
		// it, which index.php answers as a catch-all.
		//
		// A module is the one part of the front end that takes a POST. A
		// page of the instance is read only and its forms call /api,
		// which is behind access.SameOrigin; a module's forms post to
		// itself, so the same check wraps these routes. Without it a
		// third party page could make a signed in browser send a module
		// anything.
		mod := access.SameOrigin(http.HandlerFunc(h.module))
		for _, pattern := range []string{"/m/{vendor}/{plugin}", "/m/{vendor}/{plugin}/*"} {
			r.Handle(pattern, mod)
		}
		r.Get("/settings/projects", h.projects)
		r.Get("/settings/projects/new", h.projectNew)
		r.Get("/settings/projects/*", h.projectEdit)
		r.Get("/settings/workspaces", h.workspaces)
		r.Get("/settings/workspaces/new", h.workspaceNew)
		r.Get("/settings/workspaces/{folder}", h.workspaceEdit)
		r.Get("/settings/workspaces/{folder}/keys", h.workspaceKeyList)
		r.Get("/settings/workspaces/{folder}/keys/generate", h.workspaceKeyGenerate)
		r.Get("/settings/workspaces/{folder}/keys/upload", h.workspaceKeyUpload)
		r.Get("/settings/keys", h.keys)
		r.Get("/settings/keys/generate", h.keyGenerate)
		r.Get("/settings/keys/upload", h.keyUpload)
		r.Get("/settings/keys/{name}", h.keyEdit)
		r.Get("/keys", h.myKeys)
		r.Get("/settings/secrets", h.secrets)
		r.Get("/settings/integrations", h.integrations)
		r.Get("/settings/secrets/new", h.secretNew)
		r.Get("/settings/secrets/edit", h.secretEdit)
		r.Get("/settings/secrets/{id}/reads", h.secretReads)
		r.Get("/settings/users", h.users)
		r.Get("/settings/users/new", h.userNew)
		r.Get("/settings/users/{id}", h.userByID)
		r.Get("/settings/permissions", h.permissionMatrix)
		r.Get("/users/{name}", h.profile)
		r.Get("/users/{name}/edit", h.userEdit)
	})
	r.NotFound(h.opts.Guard.Require(http.HandlerFunc(h.notFound)).ServeHTTP)

	return r
}

// policy sets markdown.CSP on a response; a served file sets a stricter one of its own.
func policy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", markdown.CSP)
		next.ServeHTTP(w, r)
	})
}

// assets serves the embedded assets, cacheable for an hour. The
// stylesheet comes from view.Stylesheet, its imports versioned.
func assets() http.Handler {
	files := http.StripPrefix(view.AssetPrefix, http.FileServerFS(view.AssetsFS()))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		if r.URL.Path == view.AssetPrefix+"task-ui.css" {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
			_, _ = w.Write(view.Stylesheet())
			return
		}
		files.ServeHTTP(w, r)
	})
}

// param returns what the route bound under name.
func param(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}
