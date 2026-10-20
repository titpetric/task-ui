package access

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// Cookie is the name of the session cookie.
const Cookie = "task_ui_session"

// Guard resolves the user of a request and keeps the ones without one
// out. Login is the only way in, unless the development auto login names
// a user: then every request is that user.
type Guard struct {
	Sessions *Sessions
	Users    *storage.UserStorage
	// Tokens sign clients in without a session, through an
	// Authorization: Bearer header. Nil turns them off.
	Tokens    *storage.UserAuthTokenStorage
	AutoLogin string
	// LoginPath is where the guard sends a browser without a session.
	LoginPath string
}

// NewGuard returns a guard over the users and the sessions.
func NewGuard(users *storage.UserStorage, sessions *Sessions) *Guard {
	return &Guard{Sessions: sessions, Users: users, LoginPath: "/login"}
}

// LoginRequired reports whether requests have to carry a session.
func (g *Guard) LoginRequired() bool {
	return g.AutoLogin == ""
}

// User resolves the user of a request: a bearer token, the auto login,
// or the session cookie. Nobody is nil, and so is a request with a token
// that signs in nobody, whatever else it carries.
func (g *Guard) User(r *http.Request) *model.User {
	if token := BearerToken(r); token != "" {
		return g.bearer(r.Context(), token)
	}
	if !g.LoginRequired() {
		return g.lookup(r.Context(), g.AutoLogin)
	}
	cookie, err := r.Cookie(Cookie)
	if err != nil {
		return nil
	}
	name, err := g.Sessions.Verify(cookie.Value)
	if err != nil {
		return nil
	}
	return g.lookup(r.Context(), name)
}

// bearer resolves the user of a token, and records its use.
func (g *Guard) bearer(ctx context.Context, token string) *model.User {
	if g.Tokens == nil {
		return nil
	}
	row, err := g.Tokens.Find(ctx, HashToken(token))
	if err != nil {
		return nil
	}
	user, err := g.Users.Get(ctx, row.UserID)
	if err != nil {
		return nil
	}
	_ = g.Tokens.Touch(ctx, row.ID)
	return user
}

func (g *Guard) lookup(ctx context.Context, name string) *model.User {
	user, err := g.Users.GetByName(ctx, name)
	if err != nil {
		return nil
	}
	return user
}

// Require sends visitors without a session to the login screen, and
// answers API and websocket requests with 401.
func (g *Guard) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := g.User(r)
		if user == nil {
			if g.wantsPage(r) {
				http.Redirect(w, r, g.LoginPath+"?back="+url.QueryEscape(r.URL.RequestURI()), http.StatusSeeOther)
				return
			}
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), user)))
	})
}

// wantsPage reports a browser's request for a page, which gets a
// redirect; a client's request for data gets a status.
func (g *Guard) wantsPage(r *http.Request) bool {
	return r.Method == http.MethodGet && g.LoginRequired() &&
		!strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, "/ws/")
}

// SetCookie issues the session cookie of a user.
func (g *Guard) SetCookie(w http.ResponseWriter, r *http.Request, user *model.User) {
	http.SetCookie(w, &http.Cookie{
		Name:     Cookie,
		Value:    g.Sessions.Issue(user.Name),
		Path:     "/",
		HttpOnly: true,
		Secure:   IsHTTPS(r),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(g.Sessions.TTL().Seconds()),
	})
}

// ClearCookie ends the session. Secure says whether the site runs
// over https, which IsHTTPS tells from the request.
func (g *Guard) ClearCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     Cookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// IsHTTPS reports whether the request came over TLS, directly or through
// a proxy that says so.
func IsHTTPS(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}
