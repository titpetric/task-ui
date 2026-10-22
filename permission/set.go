package permission

import (
	"github.com/titpetric/task-ui/model"
)

// Set is what a user may do: the user, the policy of the user's role and
// the user's own.
type Set struct {
	User *model.User
	Role Policy
	Own  Policy
}

// Allowed reports whether the user may do what a path names. An admin
// may do everything; anyone else what the role or the user allows,
// unless either denies it. Nobody may do anything.
func (s Set) Allowed(path string) bool {
	if s.User == nil {
		return false
	}
	if s.User.CanAdmin() {
		return true
	}
	role, own := s.Role.Effect(path), s.Own.Effect(path)
	if role == Deny || own == Deny {
		return false
	}
	return role == Allow || own == Allow
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
