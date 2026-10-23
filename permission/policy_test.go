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
	assert.Equal(t, permission.Deny, p.Effect("note/delete"), "a deny for the rule beats an allow of all")
	denied := permission.Policy{"note/all": permission.Deny, "note/view": permission.Allow}
	assert.Equal(t, permission.Deny, denied.Effect("note/view"), "a deny of all beats an allow for the rule")
	assert.Equal(t, permission.Allow, p.Effect("run"))
	assert.Equal(t, "", p.Effect("secret/view"))
	assert.Equal(t, "", p.Effect("shell"))
	assert.Equal(t, "", permission.Policy(nil).Effect("run"))
}
