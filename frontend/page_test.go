package frontend_test

import (
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// The chrome every page shares: the user, the folders they see, and the
// recent runs with who started them.
func TestPage_Chrome(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.run(t, "infra", "plan", "ops", model.StateOK)
	f.run(t, "private", "deploy", "admin", model.StateFailed)

	body := f.page(t, asAdmin, "/")
	assert.Contains(t, body, `class="avatar">AD<`)
	assert.Contains(t, body, "Settings")
	assert.Contains(t, body, "Logout")
	assert.Contains(t, body, `data-folder="infra"`)
	assert.Contains(t, body, `data-folder="private"`)
	assert.Contains(t, body, "Recents")
	assert.Contains(t, body, `class="avatar sm" title="ops">OP<`)
	assert.Contains(t, body, `class="avatar sm" title="admin">AD<`)

	limited := f.page(t, asLimited, "/")
	assert.NotContains(t, limited, "Settings</a>")
	assert.NotContains(t, limited, `data-folder="private"`)
	assert.NotContains(t, limited, `title="admin">AD<`)
	assert.Contains(t, limited, `title="ops">OP<`)
}

func TestPage_NoticeAndError(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	body := f.page(t, asAdmin, "/runs?notice=Saved&error=Oops")
	assert.Contains(t, body, `<div class="alert ok" role="status">Saved</div>`)
	assert.Contains(t, body, `<div class="alert bad" role="alert">Oops</div>`)
}

func TestPage_AdminOnly(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	for _, path := range []string{"/settings", "/settings/projects", "/settings/keys", "/settings/secrets", "/settings/users", "/f/infra/env/.env"} {
		res := f.get(t, asOps, path)
		assert.Equal(t, http.StatusForbidden, res.Status, path)
		assert.Contains(t, res.Body, "Settings are for admins.")
	}
}
