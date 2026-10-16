package access

import (
	"slices"
	"strings"

	"github.com/titpetric/task-ui/folder"
)

// User is a configured login.
type User struct {
	Name string `yaml:"name" json:"name"`
	// Password is a bcrypt hash, generate one with `task-ui hash`.
	Password string `yaml:"password" json:"-"`
	Role     Role   `yaml:"role" json:"role"`
	// Folders limits the folders the user sees. Empty means all. The yaml
	// key is projects, the word task-ui.yml has always used.
	Folders []string `yaml:"projects,omitempty" json:"folders,omitempty"`
}

// Initials returns one or two letters that stand for the user in the
// header: the first letters of the first two words, or the first two
// letters of a single word.
func (u *User) Initials() string {
	if u == nil || u.Name == "" {
		return "?"
	}
	words := strings.FieldsFunc(u.Name, func(r rune) bool {
		return r == ' ' || r == '.' || r == '-' || r == '_' || r == '@'
	})
	if len(words) == 0 {
		return "?"
	}
	var initials []rune
	if len(words) >= 2 {
		initials = []rune{[]rune(words[0])[0], []rune(words[1])[0]}
	} else {
		initials = []rune(words[0])
		if len(initials) > 2 {
			initials = initials[:2]
		}
	}
	return strings.ToUpper(string(initials))
}

// CanSee reports whether the user may see a folder.
func (u *User) CanSee(id string) bool {
	if u == nil {
		return false
	}
	return len(u.Folders) == 0 || slices.Contains(u.Folders, id)
}

// CanRun reports whether the user may run the job of a folder. Admins may
// run everything they see; runners may run what the allow lists permit.
func (u *User) CanRun(f *folder.Folder, j *folder.Job) bool {
	if u == nil || f == nil || j == nil || !u.CanSee(f.ID) {
		return false
	}
	switch u.Role {
	case RoleAdmin:
		return true
	case RoleRunner:
		return u.allowed(f.Allow) && u.allowed(j.Allow)
	}
	return false
}

// CanAdmin reports whether the user may manage the instance.
func (u *User) CanAdmin() bool {
	return u != nil && u.Role == RoleAdmin
}

func (u *User) allowed(list []string) bool {
	if len(list) == 0 {
		return true
	}
	return slices.Contains(list, u.Name) || slices.Contains(list, string(u.Role))
}
