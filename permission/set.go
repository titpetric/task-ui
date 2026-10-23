package permission

import (
	"strings"

	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/model"
)

// Set is what a user may do: the user, the policy of the user's role and
// the user's own.
type Set struct {
	User *model.User
	Role Policy
	Own  Policy
}

// Allowed reports whether the user may do what a path names, a string literal that rulelint reads.
func (s Set) Allowed(path string) bool {
	return s.decide(path)
}

// Effective is what the set allows by every path a check may name, for
// the permission matrix.
func (s Set) Effective() map[string]bool {
	out := map[string]bool{}
	p := config.Default()
	for _, path := range p.Paths() {
		if p.Checkable(path) {
			out[path] = s.decide(path)
		}
	}
	return out
}

// Allows lists the paths a check may name that the set allows, in the
// order of permissions.yml.
func (s Set) Allows() []string {
	var paths []string
	p := config.Default()
	for _, path := range p.Paths() {
		if p.Checkable(path) && s.decide(path) {
			paths = append(paths, path)
		}
	}
	return paths
}

// Summary says in a line what the set allows, for an agent: everything
// for an admin, else the paths it allows, the rest denied.
func (s Set) Summary() string {
	switch allows := s.Allows(); {
	case s.User.CanAdmin():
		return "everything, as an admin"
	case len(allows) == 0:
		return "nothing"
	default:
		return strings.Join(allows, ", ") + "; everything else is denied"
	}
}

// decide answers Allowed for any path: an admin may do everything, anyone
// else what the user's own entry says, or without one what the role's
// says, and nobody may do anything.
func (s Set) decide(path string) bool {
	if s.User == nil {
		return false
	}
	if s.User.CanAdmin() {
		return true
	}
	if own := s.Own.Effect(path); own != "" {
		return own == Allow
	}
	return s.Role.Effect(path) == Allow
}

// Run reports whether the user may run a job, given the allow lists of
// its folder and of the job: an admin runs everything the admin sees;
// anyone else needs the run permission, a folder they see and allow
// lists that name them or their role, or are empty.
func (s Set) Run(folder string, folderAllow, jobAllow []string) bool {
	if s.User == nil || !s.User.CanSee(folder) {
		return false
	}
	if s.User.CanAdmin() {
		return true
	}
	return s.Allowed("run") && s.User.Listed(folderAllow, jobAllow)
}
