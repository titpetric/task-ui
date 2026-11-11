package forgejo_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/titpetric/task-ui/forgejo"
	"github.com/titpetric/task-ui/internal/assert"
)

// The client follows a listing over its pages, the users out of their
// envelope, sends the token, and reports what Forgejo answered when it
// refuses, the status readable.
func TestClient(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake := newFakeForgejo("t0k3n")
	fake.orgs = []forgejo.Org{{Name: "malta", FullName: "Malta"}}
	for i := 0; i < 120; i++ {
		fake.repos["malta"] = append(fake.repos["malta"], repo("malta", fmt.Sprintf("repo-%03d", i), "main", false))
	}
	client := fake.client("t0k3n")

	orgs, err := client.Orgs(ctx)
	assert.NoError(t, err)
	assert.Len(t, orgs, 1)
	repos, err := client.Repos(ctx, "malta")
	assert.NoError(t, err)
	assert.Len(t, repos, 120)
	assert.Equal(t, "repo-119", repos[119].Name)
	assert.True(t, repos[0].Permissions.Admin)

	for i := 0; i < 51; i++ {
		fake.users = append(fake.users, forgejo.User{Login: fmt.Sprintf("user-%02d", i), FullName: "User"})
	}
	fake.userRepos["user-50"] = []forgejo.Repo{repo("user-50", "notes", "main", false)}
	users, err := client.Users(ctx)
	assert.NoError(t, err)
	assert.Len(t, users, 51, "the users come out of the data envelope, over two pages")
	assert.Equal(t, "user-50", users[50].Login)
	assert.Equal(t, "User", users[50].FullName)
	repos, err = client.UserRepos(ctx, "user-50")
	assert.NoError(t, err)
	assert.Len(t, repos, 1)
	assert.Equal(t, "user-50/notes", repos[0].FullName)
	_, err = client.UserRepos(ctx, "user-99")
	assert.ErrorIs(t, err, forgejo.ErrStatus)
	assert.Equal(t, 404, forgejo.Status(err))
	fake.noUserScope = true
	_, err = client.Users(ctx)
	assert.ErrorIs(t, err, forgejo.ErrStatus)
	assert.Equal(t, 403, forgejo.Status(err))
	assert.Contains(t, err.Error(), "read:user")
	assert.Equal(t, 0, forgejo.Status(nil))
	assert.Equal(t, 0, forgejo.Status(errors.New("dial")))

	hooks, err := client.Hooks(ctx, "malta", "repo-000")
	assert.NoError(t, err)
	assert.Len(t, hooks, 0)
	made, err := client.CreateHook(ctx, "malta", "repo-000", forgejo.PushHook("https://task.example/api/webhooks/forgejo", "s3"))
	assert.NoError(t, err)
	assert.Equal(t, int64(1), made.ID)
	hooks, err = client.Hooks(ctx, "malta", "repo-000")
	assert.NoError(t, err)
	assert.Len(t, hooks, 1)
	_, shown := hooks[0].Config["secret"]
	assert.False(t, shown, "a listing leaves the secret out")
	changed, err := client.UpdateHook(ctx, "malta", "repo-000", made.ID, forgejo.PushHook("https://task.example/api/webhooks/forgejo", "s4"))
	assert.NoError(t, err)
	assert.Equal(t, "s4", changed.Config["secret"])

	_, err = client.Repos(ctx, "nobody")
	assert.ErrorIs(t, err, forgejo.ErrStatus)
	assert.Contains(t, err.Error(), "404")
	assert.Contains(t, err.Error(), "the organisation does not exist")

	_, err = fake.client("wrong").Orgs(ctx)
	assert.ErrorIs(t, err, forgejo.ErrStatus)
	assert.Contains(t, err.Error(), "401")
}

// fakeForgejo is a Forgejo API in process, behind an http.RoundTripper:
// the organisations of one token, the users of the instance, the
// repositories of each and the hooks of those, listed by the page. As
// Forgejo does, a listing of hooks leaves the secret out, and a users
// search answers in an envelope, or with a 403 when the token lacks
// read:user.
type fakeForgejo struct {
	mu    sync.Mutex
	token string
	orgs  []forgejo.Org
	users []forgejo.User
	// noUserScope refuses the users search as a token without read:user is.
	noUserScope bool
	// repos are by organisation, userRepos by login, hooks by owner/name.
	repos            map[string][]forgejo.Repo
	userRepos        map[string][]forgejo.Repo
	hooks            map[string][]forgejo.Hook
	created, updated int
	mux              *http.ServeMux
}

func newFakeForgejo(token string) *fakeForgejo {
	f := &fakeForgejo{token: token, repos: map[string][]forgejo.Repo{}, userRepos: map[string][]forgejo.Repo{}, hooks: map[string][]forgejo.Hook{}, mux: http.NewServeMux()}
	f.mux.HandleFunc("GET /api/v1/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, page(r, f.orgs))
	})
	f.mux.HandleFunc("GET /api/v1/orgs/{org}/repos", func(w http.ResponseWriter, r *http.Request) {
		repos, ok := f.repos[r.PathValue("org")]
		if !ok {
			fail(w, http.StatusNotFound, "the organisation does not exist")
			return
		}
		writeJSON(w, http.StatusOK, page(r, repos))
	})
	f.mux.HandleFunc("GET /api/v1/users/search", func(w http.ResponseWriter, r *http.Request) {
		if f.noUserScope {
			fail(w, http.StatusForbidden, "token does not have at least one of required scope(s): [read:user]")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "data": page(r, f.users)})
	})
	f.mux.HandleFunc("GET /api/v1/users/{login}/repos", func(w http.ResponseWriter, r *http.Request) {
		repos, ok := f.userRepos[r.PathValue("login")]
		if !ok {
			fail(w, http.StatusNotFound, "user does not exist")
			return
		}
		writeJSON(w, http.StatusOK, page(r, repos))
	})
	f.mux.HandleFunc("GET /api/v1/repos/{owner}/{repo}/hooks", func(w http.ResponseWriter, r *http.Request) {
		shown := []forgejo.Hook{}
		for _, hook := range f.hooks[r.PathValue("owner")+"/"+r.PathValue("repo")] {
			config := map[string]string{}
			for k, v := range hook.Config {
				if k != "secret" {
					config[k] = v
				}
			}
			hook.Config = config
			shown = append(shown, hook)
		}
		writeJSON(w, http.StatusOK, shown)
	})
	f.mux.HandleFunc("POST /api/v1/repos/{owner}/{repo}/hooks", func(w http.ResponseWriter, r *http.Request) {
		var hook forgejo.Hook
		if err := json.NewDecoder(r.Body).Decode(&hook); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		key := r.PathValue("owner") + "/" + r.PathValue("repo")
		hook.ID = int64(len(f.hooks[key]) + 1)
		f.hooks[key] = append(f.hooks[key], hook)
		f.created++
		writeJSON(w, http.StatusCreated, hook)
	})
	f.mux.HandleFunc("PATCH /api/v1/repos/{owner}/{repo}/hooks/{id}", func(w http.ResponseWriter, r *http.Request) {
		var hook forgejo.Hook
		if err := json.NewDecoder(r.Body).Decode(&hook); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
		key := r.PathValue("owner") + "/" + r.PathValue("repo")
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
		for i := range f.hooks[key] {
			if f.hooks[key][i].ID == id {
				hook.ID = id
				f.hooks[key][i] = hook
				f.updated++
				writeJSON(w, http.StatusOK, hook)
				return
			}
		}
		fail(w, http.StatusNotFound, "no such hook")
	})
	return f
}

// RoundTrip serves the request in process.
func (f *fakeForgejo) RoundTrip(r *http.Request) (*http.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	rec := httptest.NewRecorder()
	if r.Header.Get("Authorization") != "token "+f.token {
		fail(rec, http.StatusUnauthorized, "token is required")
		return rec.Result(), nil
	}
	f.mux.ServeHTTP(rec, r)
	return rec.Result(), nil
}

// client returns a client of the fake.
func (f *fakeForgejo) client(token string) *forgejo.Client {
	return &forgejo.Client{URL: "https://git.example/", Token: token, HTTP: &http.Client{Transport: f}}
}

// hooksOf returns the hooks of a repository, secrets included.
func (f *fakeForgejo) hooksOf(fullName string) []forgejo.Hook {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.hooks[fullName]
}

// page returns the page of a listing the query names.
func page[T any](r *http.Request, rows []T) []T {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 30
	}
	from := min((page-1)*limit, len(rows))
	to := min(from+limit, len(rows))
	return rows[from:to]
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"message": message})
}

// repo returns a repository of an owner on the fake instance, which the
// token administers.
func repo(owner, name, branch string, archived bool) forgejo.Repo {
	r := forgejo.Repo{
		Name: name, FullName: owner + "/" + name, DefaultBranch: branch, Archived: archived,
		SSHURL:   "ssh://git@git.example:2222/" + owner + "/" + name + ".git",
		CloneURL: "https://git.example/" + owner + "/" + name + ".git",
	}
	r.Permissions.Admin = true
	return r
}
