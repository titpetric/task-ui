package frontend

import (
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/frontend/view"
)

// loginForm shows the sign in screen. Someone signed in goes home.
func (h *Handlers) loginForm(w http.ResponseWriter, r *http.Request) {
	if h.opts.Guard.User(r) != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	p := &view.Page{Title: "Sign in", Back: access.SafeBack(r.URL.Query().Get("back")), Form: map[string]string{}, Version: h.opts.Version}
	h.render(w, http.StatusOK, "login", p)
}

// login checks the credentials and issues the session cookie.
func (h *Handlers) login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.PostFormValue("username"))
	back := access.SafeBack(r.PostFormValue("back"))
	user, err := h.opts.Auth.Authenticate(r.Context(), name, r.PostFormValue("password"))
	if err != nil {
		p := &view.Page{Title: "Sign in", Back: back, Error: "Wrong username or password.", Form: map[string]string{"username": name}, Version: h.opts.Version}
		h.render(w, http.StatusUnauthorized, "login", p)
		return
	}
	h.opts.Guard.SetCookie(w, r, user)
	http.Redirect(w, r, back, http.StatusSeeOther)
}

// logout ends the session.
func (h *Handlers) logout(w http.ResponseWriter, r *http.Request) {
	h.opts.Guard.ClearCookie(w, access.IsHTTPS(r))
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
