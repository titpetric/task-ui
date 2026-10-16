package server

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/titpetric/task-ui/access"
)

// sessionCookie is the name of the session cookie.
const sessionCookie = "task_ui_session"

type userKey struct{}

// currentUser returns the user of a request, nil for nobody.
func currentUser(r *http.Request) *access.User {
	user, _ := r.Context().Value(userKey{}).(*access.User)
	return user
}

// loginRequired reports whether requests carry a session. Login is the
// only way in, unless the development auto login is set.
func (svc *Server) loginRequired() bool {
	return svc.config.AutoLogin == ""
}

// hasUsers reports whether task-ui.yml names anyone.
func (svc *Server) hasUsers() bool {
	return len(svc.file.Get().Users) > 0
}

// authenticate resolves the user of a request from its session cookie, or
// from the auto login when one is configured.
func (svc *Server) authenticate(r *http.Request) *access.User {
	users := svc.file.Get().Users
	if !svc.loginRequired() {
		return users.Find(svc.config.AutoLogin)
	}
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return nil
	}
	name, err := svc.sessions.Verify(cookie.Value)
	if err != nil {
		return nil
	}
	return users.Find(name)
}

// requireUser sends visitors without a session to the login screen, and
// answers API and websocket requests with 401.
func (svc *Server) requireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := svc.authenticate(r)
		if user == nil {
			if r.Method == http.MethodGet && !strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, "/ws/") && svc.loginRequired() {
				http.Redirect(w, r, "/login?back="+url.QueryEscape(r.URL.RequestURI()), http.StatusSeeOther)
				return
			}
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey{}, user)))
	})
}

// requireAdmin answers anyone but an admin with 403.
func requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !currentUser(r).CanAdmin() {
			http.Error(w, "forbidden: admin role required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// sameOrigin rejects state changing requests from other sites. The session
// cookie is SameSite=Strict as well; this covers browsers that ignore it.
func sameOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead && !originMatches(r) {
			http.Error(w, "cross origin request refused", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func originMatches(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Non-browser clients and older browsers on same origin
		// form posts send no Origin; fall back to the Referer.
		origin = r.Header.Get("Referer")
		if origin == "" {
			return true
		}
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return sameHost(r, u.Host)
}

// sameHost reports whether host is the host the request was made to. A
// reverse proxy that rewrites Host passes the original in X-Forwarded-Host.
func sameHost(r *http.Request, host string) bool {
	if host == "" {
		return false
	}
	if host == r.Host {
		return true
	}
	for _, forwarded := range strings.Split(r.Header.Get("X-Forwarded-Host"), ",") {
		if strings.TrimSpace(forwarded) == host {
			return true
		}
	}
	return false
}

// safeBack returns a local redirect target, or /.
func safeBack(back string) string {
	if back == "" || !strings.HasPrefix(back, "/") || strings.HasPrefix(back, "//") || strings.HasPrefix(back, "/\\") {
		return "/"
	}
	return back
}

func isHTTPS(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}
