package server

import (
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/access"
)

// login serves and handles the sign in form.
func (svc *Server) login(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "Sign in")
	data.Back = safeBack(r.FormValue("back"))

	if !svc.loginRequired() {
		http.Redirect(w, r, data.Back, http.StatusSeeOther)
		return
	}
	if !svc.hasUsers() {
		data.Error = "No users are configured. Add one to task-ui.yml, with a password from `task-ui hash`, and sign in."
		w.WriteHeader(http.StatusServiceUnavailable)
		svc.render(w, r, "login.tpl", data)
		return
	}

	if r.Method == http.MethodPost {
		username := strings.TrimSpace(r.PostFormValue("username"))
		user := svc.file.Get().Users.Find(username)
		if access.CheckPassword(user, r.PostFormValue("password")) {
			http.SetCookie(w, &http.Cookie{
				Name:     sessionCookie,
				Value:    svc.sessions.Issue(user.Name),
				Path:     "/",
				HttpOnly: true,
				Secure:   isHTTPS(r),
				SameSite: http.SameSiteStrictMode,
				MaxAge:   int(svc.sessions.TTL().Seconds()),
			})
			http.Redirect(w, r, data.Back, http.StatusSeeOther)
			return
		}
		data.Username = username
		data.Error = "Unknown user or wrong password."
		w.WriteHeader(http.StatusUnauthorized)
	}

	svc.render(w, r, "login.tpl", data)
}

// logout clears the session.
func (svc *Server) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   isHTTPS(r),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
