package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// No route answers a private key: the path of the former read answers
// 404 for an admin, on every method.
func TestKey_NoPrivateRoute(t *testing.T) {
	t.Parallel()
	h := api.New(api.Options{}).Routes()
	admin := &model.User{ID: "u1", Name: "root", Role: model.RoleAdmin, Admin: true, Mcp: true}
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		req := httptest.NewRequest(method, "/keys/deploy_ed25519/private?recipient=age1x", nil)
		req = req.WithContext(access.WithUser(req.Context(), admin))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code, method)
	}
}
