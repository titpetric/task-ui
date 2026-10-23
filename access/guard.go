package access

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// Cookie is the name of the session cookie.
const Cookie = "task_ui_session"

// Guard resolves the user of a request and keeps the ones without one
// out. Login is the only way in, unless the development auto login names
// a user: then every request is that user. A bearer token decides alone:
// one that signs in nobody leaves the request without a user, whatever
// cookie it carries.
type Guard struct {
	Sessions *Sessions
	Users    *storage.UserStorage
	// Tokens sign clients in without a session, through an
	// Authorization: Bearer header. Nil turns them off.
	Tokens *storage.UserAuthTokenStorage
	// RunTokens sign in the Claude runs task-ui starts, for as long as
	// each runs. Nil turns them off.
	RunTokens *RunTokens
	AutoLogin string
	// LoginPath is where the guard sends a browser without a session.
	LoginPath string
	// Now is the clock a token's expiry reads; nil is the real time.
	Now func() time.Time
}

// NewGuard returns a guard over the users and the sessions.
func NewGuard(users *storage.UserStorage, sessions *Sessions) *Guard {
	return &Guard{Sessions: sessions, Users: users, LoginPath: "/login"}
}

// LoginRequired reports whether requests have to carry a session.
func (g *Guard) LoginRequired() bool {
	return g.AutoLogin == ""
}

// User resolves the user of a request from a bearer token, the auto login or the session cookie, nil for nobody.
func (g *Guard) User(r *http.Request) *model.User {
	user, _, _ := g.resolve(r)
	return user
}

// resolve is User, with the ID of the token it signed in with and why
// a bearer token signs in nobody: ErrTokenExpired for one past its
// expiry.
func (g *Guard) resolve(r *http.Request) (*model.User, string, error) {
	if token := BearerToken(r); token != "" {
		return g.bearer(r.Context(), token)
	}
	if !g.LoginRequired() {
		return g.lookup(r.Context(), g.AutoLogin), "", nil
	}
	cookie, err := r.Cookie(Cookie)
	if err != nil {
		return nil, "", nil
	}
	user, err := g.Sessions.Resolve(r.Context(), cookie.Value)
	if err != nil {
		return nil, "", nil
	}
	return user, "", nil
}

// bearer resolves the user of a token, and records its use. A token
// with an expiry signs in until then; one without signs in until a
// delete.
func (g *Guard) bearer(ctx context.Context, token string) (*model.User, string, error) {
	if id, ok := g.RunTokens.User(token); ok {
		user, err := g.Users.Get(ctx, id)
		if err != nil {
			return nil, "", nil
		}
		return user, "", nil
	}
	if g.Tokens == nil {
		return nil, "", nil
	}
	row, err := g.Tokens.Find(ctx, HashToken(token))
	if err != nil {
		return nil, "", nil
	}
	if row.ExpiresAt != nil && !g.now().Before(*row.ExpiresAt) {
		return nil, "", ErrTokenExpired
	}
	user, err := g.Users.Get(ctx, row.UserID)
	if err != nil {
		return nil, "", nil
	}
	_ = g.Tokens.Touch(ctx, row.ID)
	return user, row.ID, nil
}

func (g *Guard) now() time.Time {
	if g.Now != nil {
		return g.Now()
	}
	return time.Now()
}

func (g *Guard) lookup(ctx context.Context, name string) *model.User {
	user, err := g.Users.GetByName(ctx, name)
	if err != nil {
		return nil
	}
	return user
}

// Require sends visitors without a session to the login screen and answers API and websocket requests with 401.
func (g *Guard) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, token, err := g.resolve(r)
		if user == nil {
			if errors.Is(err, ErrTokenExpired) {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			if g.wantsPage(r) {
				http.Redirect(w, r, g.LoginPath+"?back="+url.QueryEscape(r.URL.RequestURI()), http.StatusSeeOther)
				return
			}
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := WithUser(r.Context(), user)
		if token != "" {
			ctx = WithToken(ctx, token)
		}
		if run, ok := g.RunTokens.run(BearerToken(r)); ok {
			ctx = withRun(ctx, run)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// wantsPage reports a browser's request for a page, which gets a
// redirect; a client's request for data gets a status.
func (g *Guard) wantsPage(r *http.Request) bool {
	return r.Method == http.MethodGet && g.LoginRequired() &&
		!strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, "/ws/")
}

// SetCookie opens a session for a user and issues its cookie.
func (g *Guard) SetCookie(w http.ResponseWriter, r *http.Request, user *model.User) error {
	token, err := g.Sessions.Issue(r.Context(), user)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     Cookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   IsHTTPS(r),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(g.Sessions.TTL().Seconds()),
	})
	return nil
}

// ClearCookie deletes the session of a request, so its old cookie signs in nobody, and clears the cookie.
func (g *Guard) ClearCookie(w http.ResponseWriter, r *http.Request) error {
	var err error
	if cookie, cerr := r.Cookie(Cookie); cerr == nil {
		err = g.Sessions.Revoke(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     Cookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   IsHTTPS(r),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	return err
}

// IsHTTPS reports whether the request came over TLS, directly or through a proxy that says so.
func IsHTTPS(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}
