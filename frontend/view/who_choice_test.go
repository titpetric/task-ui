package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// The roles that run, then the users who are no viewers, checked as the
// list names them; a name nobody holds stays, checked.
func TestWhoChoice(t *testing.T) {
	t.Parallel()
	users := []model.User{{Name: "ops", Role: model.RoleRunner}, {Name: "guest", Role: model.RoleViewer}, {Name: "black", Role: model.RoleAdmin}}
	var got []view.WhoChoice = view.WhoChoices([]string{"runner", "black", "gone"}, users)
	assert.Equal(t, []view.WhoChoice{
		{Name: "admin", Role: true},
		{Name: "runner", Role: true, Checked: true},
		{Name: "ops"},
		{Name: "black", Checked: true},
		{Name: "gone", Checked: true},
	}, got)
}
