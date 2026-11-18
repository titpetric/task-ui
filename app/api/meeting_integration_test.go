//go:build integration

package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/app/api"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/feature/meeting"
	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/storage"

	_ "modernc.org/sqlite"
)

// meetingAPI is the room routes over an in-memory database and a temporary channel root.
type meetingAPI struct {
	t       *testing.T
	svc     *meeting.Service
	user    *model.User
	handler http.Handler
}

// newMeetingAPI returns the routes, a user, and a room service with no Claude.
func newMeetingAPI(t *testing.T) *meetingAPI {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:", app.Migrations())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	user := &model.User{ID: model.NewID(), Name: "root", Role: model.RoleAdmin}
	svc := &meeting.Service{
		Store:    storage.NewMeetingStorage(db),
		Users:    storage.NewUserStorage(db),
		Channels: &meeting.Channels{Root: t.TempDir()},
		Hub:      meeting.NewHub(),
	}
	return &meetingAPI{t: t, svc: svc, user: user, handler: api.New(api.Options{Meetings: svc}).Routes()}
}

// call sends a request as the user and returns its recorder.
func (m *meetingAPI) call(method, path, body string) *httptest.ResponseRecorder {
	m.t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req = req.WithContext(access.WithUser(req.Context(), m.user))
	rec := httptest.NewRecorder()
	m.handler.ServeHTTP(rec, req)
	return rec
}

// decode reads a JSON answer into a map.
func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	body := map[string]any{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("%s: %v", rec.Body.String(), err)
	}
	return body
}

// A room opens, takes a participant, serves their channel, ends and is deleted.
func TestMeetingAPI(t *testing.T) {
	m := newMeetingAPI(t)

	rec := m.call(http.MethodPost, "/meetings", `{"title":"Standup"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d: %s", rec.Code, rec.Body.String())
	}
	room, _ := decode(t, rec)["meeting"].(map[string]any)
	id, _ := room["id"].(string)
	if id == "" || room["title"] != "Standup" {
		t.Fatalf("create answered %s", rec.Body.String())
	}

	if rec := m.call(http.MethodGet, "/meetings", ""); rec.Code != http.StatusOK || len(decode(t, rec)["meetings"].([]any)) != 1 {
		t.Fatalf("list: %d: %s", rec.Code, rec.Body.String())
	}

	if rec := m.call(http.MethodPost, "/meetings", `{"title":"`+strings.Repeat("a", 121)+`"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("long title: %d: %s", rec.Code, rec.Body.String())
	}

	if rec := m.call(http.MethodPost, "/meetings/"+id+"/summary", ""); rec.Code != http.StatusConflict {
		t.Fatalf("summary of an open room: %d: %s", rec.Code, rec.Body.String())
	}

	rec = m.call(http.MethodPost, "/meetings/"+id+"/join", "")
	if rec.Code != http.StatusOK || decode(t, rec)["line"] == nil {
		t.Fatalf("join: %d: %s", rec.Code, rec.Body.String())
	}

	if _, err := m.svc.Channels.Append(id, m.user.ID, make([]byte, 64)); err != nil {
		t.Fatal(err)
	}
	rec = m.call(http.MethodGet, "/meetings/"+id+"/channels/"+m.user.ID, "")
	if rec.Code != http.StatusOK || rec.Body.Len() != 44+64 || !strings.HasPrefix(rec.Body.String(), "RIFF") {
		t.Fatalf("channel: %d, %d bytes", rec.Code, rec.Body.Len())
	}

	rec = m.call(http.MethodPost, "/meetings/"+id+"/finish", "")
	if rec.Code != http.StatusOK || decode(t, rec)["meeting"].(map[string]any)["state"] != model.MeetingEnded {
		t.Fatalf("finish: %d: %s", rec.Code, rec.Body.String())
	}
	rec = m.call(http.MethodGet, "/meetings/"+id+"/summary", "")
	if rec.Code != http.StatusOK || decode(t, rec)["state"] != model.SummaryFailed {
		t.Fatalf("summary: %d: %s", rec.Code, rec.Body.String())
	}

	if rec := m.call(http.MethodDelete, "/meetings/"+id, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("delete: %d: %s", rec.Code, rec.Body.String())
	}
	if rec := m.call(http.MethodGet, "/meetings/"+id, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("get deleted: %d: %s", rec.Code, rec.Body.String())
	}
}
