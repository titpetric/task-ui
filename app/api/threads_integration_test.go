//go:build integration

package api_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/app/api"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/feature/note"
	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/permission"
	"github.com/titpetric/task-ui/storage"
)

// asUser calls a handler with the user a request is served for.
func asUser(user *model.User, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(access.WithUser(r.Context(), user)))
	})
}

// ask calls an API route as a user and returns its answer.
func ask(t *testing.T, user *model.User, next http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	rec := httptest.NewRecorder()
	asUser(user, next).ServeHTTP(rec, httptest.NewRequest(method, path, reader))
	return rec
}

// A thread's agent is changed by its author or an admin, never while it works.
func TestThreadAgentAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:", app.Migrations())
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	users := storage.NewUserStorage(db)
	alice := &model.User{Name: "alice", Role: model.RoleRunner, Admin: true}
	bob := &model.User{Name: "bob", Role: model.RoleRunner}
	claude := &model.User{Name: "claude", Role: model.RoleViewer, AgentReply: true, Mcp: true}
	helper := &model.User{Name: "helper", Role: model.RoleViewer, AgentReply: true, Mcp: true}
	plain := &model.User{Name: "plain", Role: model.RoleViewer}
	for _, user := range []*model.User{alice, bob, claude, helper, plain} {
		assert.NoError(t, users.Create(ctx, user))
	}
	policies := storage.NewPermissionPolicyStorage(db)
	assert.NoError(t, policies.Put(ctx, permission.KindRole, model.RoleRunner, "prompt/view", permission.Allow))
	notes := storage.NewNoteStorage(db)
	board := &note.Board{Notes: notes, Chats: storage.NewChatStorage(db), Users: users}
	routes := api.New(api.Options{Users: users, Board: board}).Routes()

	thread := &model.Note{Channel: model.ChannelNotes, UserID: alice.ID, UserName: alice.Name, Body: "why did the build fail", WorkspaceID: model.Default, Dir: model.Default, State: model.NoteDone, Agent: "claude"}
	assert.NoError(t, notes.Create(ctx, thread))
	path := "/threads/" + thread.ID + "/agent"

	rec := ask(t, alice, routes, http.MethodPost, path, `{"agent":"helper"}`)
	assert.Equal(t, http.StatusOK, rec.Code)
	stored, err := notes.Get(ctx, thread.ID)
	assert.NoError(t, err)
	assert.Equal(t, "helper", stored.Agent)

	rec = ask(t, bob, routes, http.MethodPost, path, `{"agent":"helper"}`)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	rec = ask(t, alice, routes, http.MethodPost, path, `{"agent":"plain"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	assert.NoError(t, notes.SetState(ctx, thread.ID, model.NoteWorking, ""))
	rec = ask(t, alice, routes, http.MethodPost, path, `{"agent":"helper"}`)
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.NoError(t, notes.SetState(ctx, thread.ID, model.NoteDone, ""))

	rec = ask(t, alice, routes, http.MethodGet, "/threads?threads=1&limit=10", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), thread.ID)
}

// A chat is posted, listed and read by its participants alone, and its read marker clears the list's count.
func TestThreadsAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:", app.Migrations())
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	users := storage.NewUserStorage(db)
	alice := &model.User{Name: "alice", Role: model.RoleRunner, Admin: true}
	bob := &model.User{Name: "bob", Role: model.RoleRunner, Admin: true}
	carol := &model.User{Name: "carol", Role: model.RoleViewer}
	for _, user := range []*model.User{alice, bob, carol} {
		assert.NoError(t, users.Create(ctx, user))
	}
	policies := storage.NewPermissionPolicyStorage(db)
	for _, path := range []string{"note/view", "note/edit"} {
		assert.NoError(t, policies.Put(ctx, permission.KindRole, model.RoleViewer, path, permission.Allow))
	}
	board := &note.Board{Notes: storage.NewNoteStorage(db), Chats: storage.NewChatStorage(db), Users: users}
	routes := api.New(api.Options{Users: users, Board: board}).Routes()

	rec := ask(t, alice, routes, http.MethodPost, "/threads", `{"to":"bob","body":"hello"}`)
	assert.Equal(t, http.StatusCreated, rec.Code)
	var root model.Note
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &root))
	assert.Equal(t, model.ChannelChat, root.Channel)

	rec = ask(t, bob, routes, http.MethodGet, "/threads", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	var list struct {
		Unread int64 `json:"unread"`
		Chats  []struct {
			Peer   string `json:"peer"`
			Unread int64  `json:"unread"`
		} `json:"chats"`
	}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	assert.Equal(t, int64(1), list.Unread)
	assert.Len(t, list.Chats, 1)
	assert.Equal(t, "alice", list.Chats[0].Peer)
	assert.Equal(t, int64(1), list.Chats[0].Unread)

	rec = ask(t, bob, routes, http.MethodGet, "/threads/"+root.ID, "")
	assert.Equal(t, http.StatusOK, rec.Code)
	var thread struct {
		With     string       `json:"with"`
		Unread   int64        `json:"unread"`
		Messages []model.Note `json:"messages"`
	}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &thread))
	assert.Equal(t, "alice", thread.With)
	assert.Equal(t, int64(1), thread.Unread)
	assert.Len(t, thread.Messages, 1)

	rec = ask(t, carol, routes, http.MethodGet, "/threads/"+root.ID, "")
	assert.Equal(t, http.StatusNotFound, rec.Code)
	rec = ask(t, carol, routes, http.MethodPost, "/threads", `{"parent_id":"`+root.ID+`","body":"hi"}`)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	rec = ask(t, bob, routes, http.MethodPost, "/threads/"+root.ID+"/read", "")
	assert.Equal(t, http.StatusNoContent, rec.Code)
	rec = ask(t, bob, routes, http.MethodGet, "/threads", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	assert.Equal(t, int64(0), list.Unread)
}
