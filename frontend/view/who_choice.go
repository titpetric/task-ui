package view

import (
	"slices"

	"github.com/titpetric/task-ui/model"
)

// WhoChoice is a box of a "who may run" field: a role or a user, and
// whether the list names it. Viewers run nothing, so they get no box; a name the list holds that is
// neither a role nor a user keeps a box of its own, checked, so saving keeps it.
type WhoChoice struct {
	Name    string
	Role    bool
	Checked bool
}

// WhoChoices lists a box for admin, runner and each user who is no viewer, checked as selected names them.
func WhoChoices(selected []string, users []model.User) []WhoChoice {
	var choices []WhoChoice
	known := map[string]bool{}
	for _, role := range []string{model.RoleAdmin, model.RoleRunner} {
		known[role] = true
		choices = append(choices, WhoChoice{Name: role, Role: true, Checked: slices.Contains(selected, role)})
	}
	for _, user := range users {
		if user.EffectiveRole() == model.RoleViewer || known[user.Name] {
			continue
		}
		known[user.Name] = true
		choices = append(choices, WhoChoice{Name: user.Name, Checked: slices.Contains(selected, user.Name)})
	}
	for _, name := range selected {
		if !known[name] {
			known[name] = true
			choices = append(choices, WhoChoice{Name: name, Checked: true})
		}
	}
	return choices
}
