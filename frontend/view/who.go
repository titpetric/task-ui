package view

import (
	"fmt"
	"slices"

	"github.com/titpetric/task-ui/model"
)

// Who is the one "who gets it" element of a form, which the whofield
// template draws: a row of the roles, each with how many users act with
// it, a list of the users, each with the role a run checks for them and
// the state the script keeps, and a summary line. The same element
// picks whose runs get an SSH key, who may run a project or a job, and
// whose runs get the Claude token; what differs is the boxes' name, their
// data-type and the sentence for nothing checked.
type Who struct {
	// Name is the roles' boxes' name, and the users' too unless UserName
	// says otherwise.
	Name string
	// UserName is the users' boxes' name: a key's users beside its policies.
	UserName string
	// Type is the boxes' data-type: set for a list, joined for a comma
	// separated setting.
	Type string
	// Roles and Users are the names the setting holds, checked.
	Roles []string
	Users []string
	// All are the users to list, with their effective roles.
	All []model.User
	// Empty is what nothing checked means, "every run gets it"; empty when
	// no run gets it then, and every user then reads doesn't.
	Empty string
	// Compact draws the roles and the users on one wrapped line, with
	// no summary, for a table cell.
	Compact bool
}

// NewWho returns the element from the boxes' name and data-type, the roles and the users checked, the users to list and the
// sentence for nothing checked. roles and users take a list of strings, a list a template builds, a set as a map or nothing;
// anything else fails the render.
func NewWho(name, typ string, roles, users any, all []model.User, empty string) (Who, error) {
	w := Who{Name: name, UserName: name, Type: typ, All: all, Empty: empty}
	var err error
	if w.Roles, err = names(roles); err != nil {
		return w, fmt.Errorf("who %s: roles: %w", name, err)
	}
	if w.Users, err = names(users); err != nil {
		return w, fmt.Errorf("who %s: users: %w", name, err)
	}
	return w, nil
}

// names returns a list of names from what a template holds them in.
func names(v any) ([]string, error) {
	switch list := v.(type) {
	case nil:
		return nil, nil
	case []string:
		return list, nil
	case map[string]bool:
		out := make([]string, 0, len(list))
		for name, on := range list {
			if on {
				out = append(out, name)
			}
		}
		slices.Sort(out)
		return out, nil
	case []any:
		out := make([]string, 0, len(list))
		for _, item := range list {
			if item == nil {
				continue
			}
			name, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("%T is not a name", item)
			}
			if name != "" {
				out = append(out, name)
			}
		}
		return out, nil
	}
	return nil, fmt.Errorf("%T is not a list of names", v)
}

// UsersNamed returns the element with the users' boxes under another name.
func (w Who) UsersNamed(name string) Who {
	w.UserName = name
	return w
}

// AsCompact returns the element in its compact form.
func (w Who) AsCompact() Who {
	w.Compact = true
	return w
}

// RoleChecked reports whether the setting names the role.
func (w Who) RoleChecked(role model.Role) bool {
	return slices.Contains(w.Roles, string(role))
}

// UserChecked reports whether the setting names the user.
func (w Who) UserChecked(name string) bool {
	return slices.Contains(w.Users, name)
}

// RoleCount returns how many of the listed users act with the role.
func (w Who) RoleCount(role model.Role) int {
	count := 0
	for i := range w.All {
		if w.All[i].EffectiveRole() == role {
			count++
		}
	}
	return count
}

// Extra returns the checked names that are neither a role nor a listed user, so a save keeps them.
func (w Who) Extra() []string {
	var extra []string
	for _, name := range w.Users {
		if slices.ContainsFunc(w.All, func(u model.User) bool { return u.Name == name }) || slices.Contains(model.Roles(), model.Role(name)) || slices.Contains(extra, name) {
			continue
		}
		extra = append(extra, name)
	}
	return extra
}
