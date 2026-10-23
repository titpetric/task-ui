//go:build integration

package access_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
	_ "modernc.org/sqlite"
)

func newGuard(t *testing.T) (*access.Guard, *model.User) {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	users := storage.NewUserStorage(db)
	ops := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.NoError(t, users.Create(context.Background(), ops))
	return access.NewGuard(users, access.NewSessions(storage.NewUserSessionStorage(db), users, time.Hour)), ops
}

// issue opens a session for a user and returns its token.
func issue(t *testing.T, guard *access.Guard, user *model.User) string {
	t.Helper()
	token, err := guard.Sessions.Issue(context.Background(), user)
	assert.NoError(t, err)
	return token
}

// hello is a handler that names the user on the request context.
var hello = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("hello " + access.UserFrom(r.Context()).Name))
})

func TestNewGuard(t *testing.T) {
	t.Parallel()
	guard, _ := newGuard(t)
	assert.Equal(t, "/login", guard.LoginPath)
	assert.True(t, guard.LoginRequired())
	guard.AutoLogin = "ops"
	assert.False(t, guard.LoginRequired())
}

func TestGuard_User(t *testing.T) {
	t.Parallel()
	guard, ops := newGuard(t)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, guard.User(r), "no cookie")
	r.AddCookie(&http.Cookie{Name: access.Cookie, Value: "a.b.c"})
	assert.Nil(t, guard.User(r), "forged cookie")

	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: access.Cookie, Value: issue(t, guard, ops)})
	assert.Equal(t, ops.ID, guard.User(r).ID)
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: access.Cookie, Value: issue(t, guard, &model.User{ID: model.NewID(), Name: "gone"})})
	assert.Nil(t, guard.User(r), "a session of a user that was removed")

	guard.AutoLogin = "ops"
	assert.Equal(t, "ops", guard.User(httptest.NewRequest(http.MethodGet, "/", nil)).Name)
	guard.AutoLogin = "nobody"
	assert.Nil(t, guard.User(httptest.NewRequest(http.MethodGet, "/", nil)))
}

func TestGuard_Require(t *testing.T) {
	t.Parallel()
	guard, ops := newGuard(t)
	handler := guard.Require(hello)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/f/echo?x=1", nil))
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/login?back=%2Ff%2Fecho%3Fx%3D1", w.Header().Get("Location"))

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/runs", nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/f/echo", nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	token := issue(t, guard, ops)
	r := httptest.NewRequest(http.MethodGet, "/f/echo", nil)
	r.AddCookie(&http.Cookie{Name: access.Cookie, Value: token})
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hello ops", w.Body.String())

	// A tampered or unknown cookie gets the login page, or a 401.
	for _, bad := range []string{token[:len(token)-1] + "A", "unknown"} {
		if bad == token {
			continue
		}
		r = httptest.NewRequest(http.MethodGet, "/f/echo", nil)
		r.AddCookie(&http.Cookie{Name: access.Cookie, Value: bad})
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusSeeOther, w.Code)
		r = httptest.NewRequest(http.MethodGet, "/api/runs", nil)
		r.AddCookie(&http.Cookie{Name: access.Cookie, Value: bad})
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	}
}

func TestGuard_SetCookie(t *testing.T) {
	t.Parallel()
	guard, ops := newGuard(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/login", nil)
	r.Header.Set("X-Forwarded-Proto", "https")
	guard.SetCookie(w, r, ops)
	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, access.Cookie, cookies[0].Name)
	assert.True(t, cookies[0].Secure)
	assert.True(t, cookies[0].HttpOnly)
	assert.Equal(t, 3600, cookies[0].MaxAge)
	assert.False(t, strings.Contains(cookies[0].Value, "."), "opaque, not a JWT")
	user, err := guard.Sessions.Resolve(r.Context(), cookies[0].Value)
	assert.NoError(t, err)
	assert.Equal(t, ops.ID, user.ID)
}

// Clearing the cookie ends the session: the old cookie signs in nobody.
func TestGuard_ClearCookie(t *testing.T) {
	t.Parallel()
	guard, ops := newGuard(t)
	token := issue(t, guard, ops)
	r := httptest.NewRequest(http.MethodPost, "/logout", nil)
	r.AddCookie(&http.Cookie{Name: access.Cookie, Value: token})
	w := httptest.NewRecorder()
	assert.NoError(t, guard.ClearCookie(w, r))
	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, -1, cookies[0].MaxAge)
	assert.False(t, cookies[0].Secure)
	assert.Nil(t, guard.User(r))

	// Without a cookie there is nothing to end.
	assert.NoError(t, guard.ClearCookie(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/logout", nil)))
}

// A bearer token signs in its user; a wrong one signs in nobody, even
// beside a good session cookie.
func TestGuard_UserBearer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	users := storage.NewUserStorage(db)
	ops := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.NoError(t, users.Create(ctx, ops))
	guard := access.NewGuard(users, access.NewSessions(storage.NewUserSessionStorage(db), users, time.Hour))

	bearer := func(token string) *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		return r
	}
	assert.Nil(t, guard.User(bearer("tui_x")), "tokens off")

	guard.Tokens = storage.NewUserAuthTokenStorage(db)
	token, hash, err := access.NewToken()
	assert.NoError(t, err)
	assert.NoError(t, guard.Tokens.Set(ctx, ops.ID, "cli", hash))

	assert.Equal(t, ops.ID, guard.User(bearer(token)).ID)
	row, err := guard.Tokens.Find(ctx, hash)
	assert.NoError(t, err)
	assert.NotNil(t, row.UsedAt)

	r := bearer("tui_wrong")
	r.AddCookie(&http.Cookie{Name: access.Cookie, Value: issue(t, guard, ops)})
	assert.Nil(t, guard.User(r))

	w := httptest.NewRecorder()
	guard.Require(hello).ServeHTTP(w, bearer(token))
	assert.Equal(t, "hello ops", w.Body.String())
}

// A token with an expiry signs in until then, and after it Require
// answers 401 with what to do; a token without one keeps signing in.
func TestGuard_UserBearerExpiry(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	users := storage.NewUserStorage(db)
	ops := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.NoError(t, users.Create(ctx, ops))
	guard := access.NewGuard(users, access.NewSessions(storage.NewUserSessionStorage(db), users, time.Hour))
	now := time.Date(2026, time.October, 22, 12, 0, 0, 0, time.UTC)
	guard.Now = func() time.Time { return now }
	guard.Tokens = storage.NewUserAuthTokenStorage(db)

	mcp, mcpHash, err := access.NewToken()
	assert.NoError(t, err)
	_, err = guard.Tokens.SetUntil(ctx, ops.ID, access.MCPToken, mcpHash, now.Add(access.MCPTokenTTL))
	assert.NoError(t, err)
	cli, cliHash, err := access.NewToken()
	assert.NoError(t, err)
	assert.NoError(t, guard.Tokens.Set(ctx, ops.ID, "cli", cliHash))

	call := func(token string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, "/api/mcp", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		guard.Require(hello).ServeHTTP(w, r)
		return w
	}
	assert.Equal(t, "hello ops", call(mcp).Body.String())

	now = now.Add(access.MCPTokenTTL)
	expired := call(mcp)
	assert.Equal(t, http.StatusUnauthorized, expired.Code)
	assert.Contains(t, expired.Body.String(), "the token has expired; rotate it on your profile")
	assert.Equal(t, "hello ops", call(cli).Body.String())
	wrong := call("tui_wrong")
	assert.Equal(t, http.StatusUnauthorized, wrong.Code)
	assert.NotContains(t, wrong.Body.String(), "expired")
}
