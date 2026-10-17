package access_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "modernc.org/sqlite" // the driver storage opens

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func newGuard(t *testing.T) (*access.Guard, *model.User) {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	users := storage.NewUserStorage(db)
	ops := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.NoError(t, users.Create(context.Background(), ops))
	return access.NewGuard(users, access.NewSessions("secret", time.Hour)), ops
}

// hello is a handler that names the user it was reached with.
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
	r.AddCookie(&http.Cookie{Name: access.Cookie, Value: guard.Sessions.Issue("ops")})
	assert.Equal(t, ops.ID, guard.User(r).ID)
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: access.Cookie, Value: guard.Sessions.Issue("gone")})
	assert.Nil(t, guard.User(r), "a session of a user that was removed")

	guard.AutoLogin = "ops"
	assert.Equal(t, "ops", guard.User(httptest.NewRequest(http.MethodGet, "/", nil)).Name)
	guard.AutoLogin = "nobody"
	assert.Nil(t, guard.User(httptest.NewRequest(http.MethodGet, "/", nil)))
}

func TestGuard_Require(t *testing.T) {
	t.Parallel()
	guard, _ := newGuard(t)
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

	r := httptest.NewRequest(http.MethodGet, "/f/echo", nil)
	r.AddCookie(&http.Cookie{Name: access.Cookie, Value: guard.Sessions.Issue("ops")})
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hello ops", w.Body.String())
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
	name, err := guard.Sessions.Verify(cookies[0].Value)
	assert.NoError(t, err)
	assert.Equal(t, "ops", name)
}

func TestGuard_ClearCookie(t *testing.T) {
	t.Parallel()
	guard, _ := newGuard(t)
	w := httptest.NewRecorder()
	guard.ClearCookie(w, false)
	cookies := w.Result().Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, -1, cookies[0].MaxAge)
	assert.False(t, cookies[0].Secure)
}

func TestIsHTTPS(t *testing.T) {
	t.Parallel()
	assert.True(t, access.IsHTTPS(httptest.NewRequest(http.MethodGet, "https://x/", nil)))
	assert.False(t, access.IsHTTPS(httptest.NewRequest(http.MethodGet, "http://x/", nil)))
	r := httptest.NewRequest(http.MethodGet, "http://x/", nil)
	r.Header.Set("X-Forwarded-Proto", "HTTPS")
	assert.True(t, access.IsHTTPS(r))
}
