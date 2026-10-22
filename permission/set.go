package permission

import (
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

// Allowed reports whether the user may do what a path names. An admin
// may do everything; anyone else what the role or the user allows,
// unless either denies it. Nobody may do anything. A check names its
// path as a string literal, which rulelint reads.
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

// decide answers as Allowed does, for any path.
func (s Set) decide(path string) bool {
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
