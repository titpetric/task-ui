package model

import (
	"slices"
	"strings"
)

// Initials returns one or two letters that stand for the user in the
// header: the first letters of the first two words of the name, or the
// first two letters of a single word.
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
	initials := []rune(words[0])[:1]
	if len(words) >= 2 {
		initials = append(initials, []rune(words[1])[0])
	} else if runes := []rune(words[0]); len(runes) > 1 {
		initials = runes[:2]
	}
	return strings.ToUpper(string(initials))
}

// FolderList returns the folders the user may see. Empty means all.
func (u *User) FolderList() []string {
	if u == nil {
		return nil
	}
	return DecodeList(u.Folders)
}

// CanSee reports whether the user may see a folder.
func (u *User) CanSee(folder string) bool {
	if u == nil {
		return false
	}
	list := u.FolderList()
	return len(list) == 0 || slices.Contains(list, folder)
}

// CanRun reports whether the user may run a job, given the allow list of
// its folder and the allow list of the job. Admins may run everything
// they see; runners may run what the allow lists permit; viewers run
// nothing.
func (u *User) CanRun(folder string, folderAllow, jobAllow []string) bool {
	if u == nil || !u.CanSee(folder) {
		return false
	}
	switch u.Role {
	case RoleAdmin:
		return true
	case RoleRunner:
		return u.allowed(folderAllow) && u.allowed(jobAllow)
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
	return slices.Contains(list, u.Name) || slices.Contains(list, u.Role)
}
