package permission_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/permission"
)

func TestPolicy_Effect(t *testing.T) {
	t.Parallel()
	p := permission.Policy{"note/all": permission.Allow, "note/delete": permission.Deny, "run": permission.Allow}
	assert.Equal(t, permission.Allow, p.Effect("note/view"), "all stands for every rule")
	assert.Equal(t, permission.Deny, p.Effect("note/delete"), "an entry of its own overrides all")
	assert.Equal(t, permission.Allow, p.Effect("run"))
	assert.Equal(t, "", p.Effect("secret/view"))
	assert.Equal(t, "", p.Effect("shell"))
	assert.Equal(t, "", permission.Policy(nil).Effect("run"))
}
