package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

// memSettings are instance settings in memory.
type memSettings map[string]string

func (s memSettings) Values(context.Context) (map[string]string, error) {
	values := map[string]string{}
	for k, v := range s {
		values[k] = v
	}
	return values, nil
}

func (s memSettings) Set(_ context.Context, values map[string]string) error {
	for k, v := range values {
		s[k] = v
	}
	return nil
}

// instanceCall sends a request to /instance as a user.
func instanceCall(t *testing.T, h http.Handler, user *model.User, method, body string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, "/instance", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(access.WithUser(req.Context(), user))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	out := map[string]any{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func TestInstanceRequest(t *testing.T) {
	t.Parallel()
	for body, want := range map[string]string{`{"backlog_length": 12}`: "12", `{"backlog_length": "40"}`: "40", `{"backlog_length": null}`: ""} {
		var req api.InstanceRequest
		assert.NoError(t, json.Unmarshal([]byte(body), &req), body)
		if want == "" && req.BacklogLength == nil {
			continue
		}
		assert.NotNil(t, req.BacklogLength, body)
		assert.Equal(t, want, string(*req.BacklogLength), body)
	}
}

func TestInstanceRoutes(t *testing.T) {
	t.Parallel()
	values := memSettings{}
	manager := runner.New(runner.Options{Backlog: &runner.Backlog{Settings: values, CPUs: 4}})
	h := api.New(api.Options{Runner: manager}).Routes()
	admin := &model.User{Name: "root", Role: model.RoleAdmin}
	viewer := &model.User{Name: "ann", Role: model.RoleViewer}

	status, body := instanceCall(t, h, viewer, http.MethodGet, "")
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, float64(80), body["backlog_length_auto"])
	assert.Equal(t, float64(80), body["backlog_length_effective"])

	status, _ = instanceCall(t, h, viewer, http.MethodPut, `{"backlog_length": 5}`)
	assert.Equal(t, http.StatusForbidden, status, "only an admin sets it")

	status, body = instanceCall(t, h, admin, http.MethodPut, `{"backlog_length": 5}`)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, float64(5), body["backlog_length"])
	assert.Equal(t, float64(5), body["backlog_length_effective"])
	assert.Equal(t, "5", values[runner.BacklogSetting])

	status, body = instanceCall(t, h, admin, http.MethodPut, `{"backlog_length": "lots"}`)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Contains(t, body["error"], "whole number")

	status, body = instanceCall(t, h, admin, http.MethodPut, `{"backlog_length": ""}`)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, float64(80), body["backlog_length_effective"], "empty is the automatic length")
}
