//go:build integration

package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/app/api"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/permission"
	"github.com/titpetric/task-ui/storage"
)

// A managed ci:daily row is the admin's to change; the first change makes it
// an ordinary schedule.
func TestScheduleSourceAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:", app.Migrations())
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	users := storage.NewUserStorage(db)
	alice := &model.User{Name: "alice", Role: model.RoleAdmin, Admin: true}
	bob := &model.User{Name: "bob", Role: model.RoleViewer, Folders: model.EncodeList([]string{"acme"})}
	for _, user := range []*model.User{alice, bob} {
		assert.NoError(t, users.Create(ctx, user))
	}
	policies := storage.NewPermissionPolicyStorage(db)
	for _, rule := range []string{"schedule/view", "schedule/edit"} {
		assert.NoError(t, policies.Put(ctx, permission.KindRole, model.RoleViewer, rule, permission.Allow))
	}

	schedules := storage.NewScheduleStorage(db)
	managed := &model.Schedule{Folder: "acme", Job: "ci:daily", Expression: "@daily", Source: "ci:daily", UserID: bob.ID, UserName: bob.Name}
	assert.NoError(t, schedules.Create(ctx, managed))
	router := api.New(api.Options{Users: users, Schedules: schedules}).Routes()

	rec := ask(t, bob, router, http.MethodPut, "/schedules/"+managed.ID, `{"enabled":true}`)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	rec = ask(t, alice, router, http.MethodPut, "/schedules/"+managed.ID, `{"enabled":true}`)
	assert.Equal(t, http.StatusOK, rec.Code)
	var changed model.Schedule
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &changed))
	assert.Equal(t, "", changed.Source)
	assert.Equal(t, int64(1), changed.Enabled)

	stored, err := schedules.Get(ctx, managed.ID)
	assert.NoError(t, err)
	assert.Equal(t, "", stored.Source)
	assert.Equal(t, int64(1), stored.Enabled)
	assert.True(t, stored.NextRunAt != nil, "enabling a row with no next run computes one")
}
