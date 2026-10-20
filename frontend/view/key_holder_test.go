package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/sshkey"
)

// A role row lists the keys of the role; a user row the keys given to
// them by name and, apart, those their role gets. A backup is nobody's.
func TestKeyHolder(t *testing.T) {
	t.Parallel()
	keys := []sshkey.Key{{Name: "id_rsa_deploy", Comment: "deploy"}, {Name: "id_ed25519_alice", Comment: "alice"}, {Name: "id_rsa_deploy.1", Backup: true}}
	byRole := map[string][]string{"id_rsa_deploy": {"admin"}, "id_rsa_deploy.1": {"admin"}}
	byUser := map[string][]string{"id_ed25519_alice": {"alice"}}
	users := []model.User{{Name: "alice", Role: model.RoleAdmin}, {Name: "bob", Role: model.RoleViewer}}

	var rows []view.KeyHolder = view.KeyHolders(keys, byRole, byUser, users)
	assert.Len(t, rows, len(model.Roles())+2)
	assert.Equal(t, "admin", rows[0].Role)
	assert.Len(t, rows[0].Direct, 1)
	alice := rows[len(model.Roles())]
	assert.Equal(t, "alice", alice.User.Name)
	assert.Equal(t, "alice", alice.Direct[0].Comment)
	assert.Equal(t, "deploy", alice.ViaRole[0].Comment)
	bob := rows[len(model.Roles())+1]
	assert.Len(t, bob.Direct, 0)
	assert.Len(t, bob.ViaRole, 0)
}
