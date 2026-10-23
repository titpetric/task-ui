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

// CanSee reports whether the user may see a folder. The internal folder
// of older runs, and those runs, are the admins'.
func (u *User) CanSee(folder string) bool {
	if u == nil {
		return false
	}
	if IsInternal(folder) {
		return u.CanAdmin()
	}
	list := u.FolderList()
	return len(list) == 0 || slices.Contains(list, folder)
}

// SeesArtifact reports whether the user may see an artifact: one a run
// of a folder the user sees left, or one a user wrote into the tree with
// no run, which belongs to no folder.
func (u *User) SeesArtifact(artifact *Artifact) bool {
	if u == nil || artifact == nil {
		return false
	}
	if artifact.RunID == "" && artifact.Folder == "" {
		return true
	}
	return u.CanSee(artifact.Folder)
}

// SeesRun reports whether the user may see a run: one of a folder the user sees, or the user's own prompt or note run.
// Another user's prompt run also takes prompts (prompt/view), and a Notes agent's run takes notes (note/view).
func (u *User) SeesRun(run *Run, prompts, notes bool) bool {
	if u == nil || run == nil {
		return false
	}
	if (run.IsPrompt() || run.IsNote()) && run.UserID == u.ID {
		return true
	}
	return u.CanSee(run.Folder) && (prompts || !run.IsPrompt()) && (notes || !run.IsNote())
}

// Listed reports whether the allow lists of a job's folder and of the job both let the user run it.
func (u *User) Listed(folderAllow, jobAllow []string) bool {
	return u != nil && u.allowed(folderAllow) && u.allowed(jobAllow)
}

// CanAdmin reports whether the user is an admin by the flag or the role, whatever the policies say.
func (u *User) CanAdmin() bool {
	return u != nil && (u.Admin || u.Role == RoleAdmin)
}

// EffectiveRole returns the role the user acts with: admin for an
// admin, else the role column, runner or viewer.
func (u *User) EffectiveRole() Role {
	if u == nil {
		return ""
	}
	if u.CanAdmin() {
		return RoleAdmin
	}
	return u.Role
}

// allowed reports whether an allow list is empty or names the user or the
// user's role. What the user may do besides is the permission package's.
func (u *User) allowed(list []string) bool {
	if len(list) == 0 {
		return true
	}
	return slices.Contains(list, u.Name) || slices.Contains(list, u.EffectiveRole())
}
