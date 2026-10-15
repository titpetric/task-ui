package model

import (
	"slices"
	"time"
)

// Project is a folder holding a task file. Its runbooks are the tasks the
// file declares that are meant to be run from the UI.
type Project struct {
	// ID is the URL safe name of the project, the folder name by default.
	ID string `json:"id"`
	// Title is shown in the navigation, falls back to ID.
	Title string `json:"title"`
	// Description is shown above the runbook table.
	Description string `json:"description,omitempty"`

	// Dir is the folder the runbooks run in.
	Dir string `json:"-"`
	// File is the task file the runbooks were read from.
	File string `json:"file"`
	// Kind is "atkins" or "task", which picks the default command.
	Kind string `json:"kind"`
	// Command runs a runbook: the runbook name is appended as the last
	// argument. Defaults to `atkins -f <file>` or `task -t <file>`.
	Command []string `json:"-"`

	// Allow lists the users and roles that may run this project's runbooks.
	// Empty allows anyone with the run permission.
	Allow []string `json:"allow,omitempty"`

	Runbooks []*Runbook `json:"runbooks"`

	// Error holds a parse error. The project is listed so the problem is
	// visible, but it has no runbooks.
	Error string `json:"error,omitempty"`
}

// Runbook returns the runbook by name or alias, or nil.
func (p *Project) Runbook(name string) *Runbook {
	if p == nil {
		return nil
	}
	for _, r := range p.Runbooks {
		if r.Name == name || slices.Contains(r.Aliases, name) {
			return r
		}
	}
	return nil
}

// Runbook is a single runnable task of a project.
type Runbook struct {
	Project     string   `json:"project"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Summary     string   `json:"summary,omitempty"`
	Aliases     []string `json:"aliases,omitempty"`
	Interactive bool     `json:"interactive,omitempty"`

	// Allow narrows who may run this runbook, on top of the project Allow.
	Allow []string `json:"allow,omitempty"`
}

// Key is the identity used for the run lock: one run per runbook per
// project, across the instance.
func (r *Runbook) Key() string {
	return r.Project + "/" + r.Name
}

// Role names the permission sets a user can hold.
type Role string

const (
	// RoleAdmin may run anything and manage the instance SSH keys.
	RoleAdmin Role = "admin"
	// RoleRunner may run runbooks and read history.
	RoleRunner Role = "runner"
	// RoleViewer may read runbooks and history, and run nothing.
	RoleViewer Role = "viewer"
)

// Valid reports whether the role is one of the known roles.
func (r Role) Valid() bool {
	switch r {
	case RoleAdmin, RoleRunner, RoleViewer:
		return true
	}
	return false
}

// User is a configured login.
type User struct {
	Name string `yaml:"name" json:"name"`
	// Password is a bcrypt hash, generate one with `task-ui hash`.
	Password string `yaml:"password" json:"-"`
	Role     Role   `yaml:"role" json:"role"`
	// Projects limits the projects the user sees. Empty means all.
	Projects []string `yaml:"projects,omitempty" json:"projects,omitempty"`
}

// CanSee reports whether the user may see a project.
func (u *User) CanSee(project string) bool {
	if u == nil {
		return false
	}
	return len(u.Projects) == 0 || slices.Contains(u.Projects, project)
}

// CanRun reports whether the user may run the runbook of a project. Admins
// may run everything they see; runners may run what the allow lists permit.
func (u *User) CanRun(p *Project, r *Runbook) bool {
	if u == nil || p == nil || r == nil || !u.CanSee(p.ID) {
		return false
	}
	switch u.Role {
	case RoleAdmin:
		return true
	case RoleRunner:
		return u.allowed(p.Allow) && u.allowed(r.Allow)
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

// HistoryRecord describes one recorded run.
type HistoryRecord struct {
	ID       string        `json:"id"`
	Project  string        `json:"project"`
	Runbook  string        `json:"runbook"`
	User     string        `json:"user"`
	Started  time.Time     `json:"started"`
	Duration time.Duration `json:"duration"`
	// State is how the run ended: ok, failed, stopped, killed, or running
	// when the process is still going or the instance went away mid run.
	State    string `json:"state"`
	ExitCode int    `json:"exit_code"`
	Forced   bool   `json:"forced,omitempty"`
}

// Run states.
const (
	StateRunning = "running"
	StateOK      = "ok"
	StateFailed  = "failed"
	StateStopped = "stopped"
	StateKilled  = "killed"
)

// SSHKey describes a key pair in the instance SSH folder.
type SSHKey struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Bits        int       `json:"bits,omitempty"`
	Fingerprint string    `json:"fingerprint"`
	Comment     string    `json:"comment"`
	PublicKey   string    `json:"public_key"`
	Modified    time.Time `json:"modified"`
	// Backup is set for keys moved aside by a rotation.
	Backup bool `json:"backup,omitempty"`
}
