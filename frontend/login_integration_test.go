//go:build integration

package frontend_test

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
)

// post submits a form as a browser does.
func (f *fixture) post(t *testing.T, path string, form url.Values) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, f.server.URL+path, strings.NewReader(form.Encode()))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := client().Do(req)
	assert.NoError(t, err)
	res.Body.Close()
	return res
}

func TestLogin(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	hash, err := access.HashPassword("demo")
	assert.NoError(t, err)
	ctx := context.Background()
	admin := f.users[asAdmin]
	assert.NoError(t, f.passwords.SetPassword(ctx, admin.ID, hash))

	form := f.get(t, "", "/login?back=/f/internal/infra")
	assert.Equal(t, http.StatusOK, form.Status)
	assert.Contains(t, form.Body, "Sign in")
	assert.Contains(t, form.Body, `name="back" value="/f/internal/infra"`)
	assert.NotContains(t, form.Body, "Recents")
	assert.NotContains(t, form.Body, "header-prompt")

	wrong := f.post(t, "/login", url.Values{"username": {"admin"}, "password": {"nope"}, "back": {"/f/internal/infra"}})
	assert.Equal(t, http.StatusUnauthorized, wrong.StatusCode)

	right := f.post(t, "/login", url.Values{"username": {" admin "}, "password": {"demo"}, "back": {"/f/internal/infra"}})
	assert.Equal(t, http.StatusSeeOther, right.StatusCode)
	assert.Equal(t, "/f/internal/infra", right.Header.Get("Location"))
	var cookie *http.Cookie
	for _, c := range right.Cookies() {
		if c.Name == access.Cookie {
			cookie = c
		}
	}
	assert.NotNil(t, cookie)
	assert.True(t, cookie.HttpOnly)
	assert.False(t, strings.Contains(cookie.Value, "."), "an opaque token, not a JWT")
	user, err := f.opts.Guard.Sessions.Resolve(ctx, cookie.Value)
	assert.NoError(t, err)
	assert.Equal(t, admin.ID, user.ID)

	// The cookie signs the next request in.
	req, _ := http.NewRequest(http.MethodGet, f.server.URL+"/", nil)
	req.AddCookie(cookie)
	res, err := client().Do(req)
	assert.NoError(t, err)
	res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// The server refuses a sign in from another site.
	req, _ = http.NewRequest(http.MethodPost, f.server.URL+"/login", strings.NewReader(url.Values{"username": {"admin"}, "password": {"demo"}}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://evil.example")
	res, err = client().Do(req)
	assert.NoError(t, err)
	res.Body.Close()
	assert.Equal(t, http.StatusForbidden, res.StatusCode)

	// Off site targets land home.
	elsewhere := f.post(t, "/login", url.Values{"username": {"admin"}, "password": {"demo"}, "back": {"https://evil.example"}})
	assert.Equal(t, "/", elsewhere.Header.Get("Location"))
}

func TestLogin_SignedIn(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	res := f.get(t, asOps, "/login")
	assert.Equal(t, http.StatusSeeOther, res.Status)
	assert.Equal(t, "/", res.Location)
}

// Logout ends the session on the server: the old cookie signs in
// nobody afterwards.
func TestLogout(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	token, err := f.opts.Guard.Sessions.Issue(context.Background(), f.users[asOps])
	assert.NoError(t, err)
	session := &http.Cookie{Name: access.Cookie, Value: token}
	home := func() *http.Response {
		req, _ := http.NewRequest(http.MethodGet, f.server.URL+"/", nil)
		req.AddCookie(session)
		res, err := client().Do(req)
		assert.NoError(t, err)
		res.Body.Close()
		return res
	}
	assert.Equal(t, http.StatusOK, home().StatusCode)

	req, _ := http.NewRequest(http.MethodPost, f.server.URL+"/logout", nil)
	req.AddCookie(session)
	res, err := client().Do(req)
	assert.NoError(t, err)
	res.Body.Close()
	assert.Equal(t, http.StatusSeeOther, res.StatusCode)
	assert.Equal(t, "/login", res.Header.Get("Location"))
	for _, c := range res.Cookies() {
		if c.Name == access.Cookie {
			assert.Equal(t, -1, c.MaxAge)
		}
	}

	again := home()
	assert.Equal(t, http.StatusSeeOther, again.StatusCode)
	assert.True(t, strings.HasPrefix(again.Header.Get("Location"), "/login"))

	// Without a session, logout still lands on the login page.
	res = f.post(t, "/logout", nil)
	assert.Equal(t, http.StatusSeeOther, res.StatusCode)
}

// An email address signs in as the user it belongs to, and a backup
// password signs in once.
func TestLogin_EmailAndBackup(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	ops := f.users[asOps]
	assert.NoError(t, f.emails.Add(ctx, ops.ID, "Ops@Example.com"))
	codes, hashes, err := access.BackupPasswords(1)
	assert.NoError(t, err)
	assert.NoError(t, f.passwords.SetBackups(ctx, ops.ID, hashes))

	first := f.post(t, "/login", url.Values{"username": {"ops@example.com"}, "password": {codes[0]}, "back": {"/"}})
	assert.Equal(t, http.StatusSeeOther, first.StatusCode)
	again := f.post(t, "/login", url.Values{"username": {"ops"}, "password": {codes[0]}, "back": {"/"}})
	assert.Equal(t, http.StatusUnauthorized, again.StatusCode)
}
