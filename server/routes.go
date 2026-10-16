package server

import (
	"io/fs"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// routes is the one place the router is reached: every handler is mounted
// here, and param reads what the patterns bound.
func (svc *Server) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(sameOrigin)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok\n")) })
	r.Handle("/static/*", http.StripPrefix("/static/", staticHandler(svc.assets)))
	r.Get("/login", svc.login)
	r.Post("/login", svc.login)
	r.Post("/logout", svc.logout)

	r.Group(func(r chi.Router) {
		r.Use(svc.requireUser)

		r.Get("/", svc.index)
		r.Get("/f/{folder}", svc.folderPage)
		r.Get("/f/{folder}/{job}", svc.jobPage)

		r.Get("/api/runs", svc.apiRuns)
		r.Post("/api/runs/{folder}/{job}/stop", svc.apiStop)
		r.Get("/api/history/{folder}/{job}/{id}", svc.historyFile)
		r.Handle("/ws/run/{folder}/{job}", svc.runSocket())

		r.Group(func(r chi.Router) {
			r.Use(requireAdmin)
			r.Get("/settings", svc.settingsPage)
			r.Get("/settings/keys", svc.keysPage)
			r.Get("/settings/keys/generate", svc.keysGeneratePage)
			r.Post("/settings/keys/generate", svc.keysGenerate)
			r.Get("/settings/keys/upload", svc.keysUploadPage)
			r.Post("/settings/keys/upload", svc.keysUpload)
			r.Post("/settings/keys/delete", svc.keysDelete)
			r.Get("/settings/secrets", svc.secretsPage)
			r.Get("/settings/secrets/new", svc.secretForm)
			r.Get("/settings/secrets/edit", svc.secretForm)
			r.Post("/settings/secrets/save", svc.secretSave)
			r.Post("/settings/secrets/delete", svc.secretDelete)
			r.Get("/settings/projects", svc.projectsPage)
			r.Get("/settings/projects/new", svc.projectNewPage)
			r.Post("/settings/projects/clone", svc.projectClone)
			r.Post("/settings/projects/delete", svc.projectDelete)
			r.Get("/f/{folder}/env/*", svc.envFilePage)
			r.Post("/f/{folder}/env/*", svc.envFileSave)
		})
	})
	return r
}

// param returns what the route bound under name.
func param(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}

func staticHandler(assets fs.FS) http.Handler {
	files := http.FileServer(http.FS(assets))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=300")
		files.ServeHTTP(w, r)
	})
}
