package access

import (
	"net/http"
	"net/url"
	"strings"
)

// RequireAdmin answers anyone but an admin with 403. It runs after
// Guard.Require.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !UserFrom(r.Context()).CanAdmin() {
			http.Error(w, "forbidden: admin role required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// SameOrigin rejects state changing requests from other sites. The session
// cookie is SameSite=Strict as well; this covers browsers that ignore it.
func SameOrigin(next http.Handler) http.Handler {
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
		// Non-browser clients and older browsers on same origin form
		// posts send no Origin; fall back to the Referer.
		origin = r.Header.Get("Referer")
		if origin == "" {
			return true
		}
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return SameHost(r, u.Host)
}

// SameHost reports whether host is the host the request targets. A
// reverse proxy that rewrites Host passes the original in X-Forwarded-Host.
func SameHost(r *http.Request, host string) bool {
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

// SafeBack returns a local redirect target, or /.
func SafeBack(back string) string {
	if back == "" || !strings.HasPrefix(back, "/") || strings.HasPrefix(back, "//") || strings.HasPrefix(back, "/\\") {
		return "/"
	}
	return back
}
