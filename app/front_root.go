package app

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	chi "github.com/go-chi/chi/v5"

	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/phpfront"
)

// rootRoutes registers what is no page of the front end: the session, the readiness check, a plugin's pages.
func (a *App) rootRoutes(r chi.Router, auth *access.Authenticator) {
	r.Method(http.MethodPost, "/login", access.SameOrigin(a.login(auth)))
	r.Method(http.MethodPost, "/logout", access.SameOrigin(http.HandlerFunc(a.logout)))
	r.Method(http.MethodGet, "/healthz", http.HandlerFunc(health))

	plugin := a.Guard.Require(access.SameOrigin(http.HandlerFunc(a.plugin)))
	for _, pattern := range []string{"/m/{vendor}/{plugin}", "/m/{vendor}/{plugin}/*"} {
		r.Handle(pattern, plugin)
	}
}

// login reads the credentials the PHP form posts and mints the session; a wrong password goes back under ?error.
func (a *App) login(auth *access.Authenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		name := strings.TrimSpace(r.PostFormValue("username"))
		back := access.SafeBack(r.PostFormValue("back"))
		user, err := auth.Authenticate(r.Context(), name, r.PostFormValue("password"))
		if err != nil {
			http.Redirect(w, r, "/login?error=credentials&back="+url.QueryEscape(back), http.StatusSeeOther)
			return
		}
		if err := a.Guard.SetCookie(w, r, user); err != nil {
			http.Error(w, "sign in failed", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, back, http.StatusSeeOther)
	})
}

// logout ends the session: the server forgets it and the browser drops the cookie.
func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	if err := a.Guard.ClearCookie(w, r); err != nil {
		http.Error(w, "sign out failed", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// health answers the readiness check whatever watches the instance asks for.
func health(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("ok\n"))
}

// plugin serves a page of a plugin written in PHP under /m, as its index.php echoed it.
func (a *App) plugin(w http.ResponseWriter, r *http.Request) {
	// Not rewritten on the way in: the module reads the whole path, prefix and
	// all, out of REQUEST_URI.
	out, err := a.Modules.Render(r.Context(), w, r)
	switch {
	case errors.Is(err, phpfront.ErrNotFound):
		http.Error(w, "no module answers "+r.URL.Path, http.StatusNotFound)
		return
	case err != nil:
		// The trace of this request holds the span the module recorded.
		slog.Error("rendering a module", "path", r.URL.Path, "err", err)
		http.Error(w, "the module failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// A module that staged a redirect gets one. Its forms post to themselves
	// and answer with "see other", and a body under a 303 would leave the
	// browser on the form with the work already done.
	if location := out.Header.Get("Location"); location != "" && out.Status >= 300 && out.Status < 400 {
		http.Redirect(w, r, location, out.Status)
		return
	}

	for name, values := range out.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	if out.File {
		// A module's asset is served from the instance, so it carries the same
		// caching as the instance's own: short, and revalidated.
		w.Header().Set("Cache-Control", "public, max-age=300")
	}
	w.WriteHeader(out.Status)
	if _, err := w.Write(out.Body); err != nil {
		slog.Debug("writing a module's answer", "path", r.URL.Path, "err", err)
	}
}
