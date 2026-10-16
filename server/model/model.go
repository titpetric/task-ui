package model

import (
	"slices"
	"time"
)

// Project is a folder holding a pipeline file: a working directory and the
// atkins.yml (or Taskfile.yml) in it. Its jobs are what the file declares.
type Project struct {
	// ID is the URL safe name of the folder.
	ID string `json:"id"`
	// Title is the pipeline name, falls back to the folder name.
	Title string `json:"title"`
	// Description is the pipeline help text.
	Description string `json:"description,omitempty"`

	// Dir is the working directory the jobs run in.
	Dir string `json:"-"`
	// File is the pipeline file name inside Dir.
	File string `json:"file"`
	// Kind is "atkins" or "task", which picks the default command.
	Kind string `json:"kind"`
	// Command runs a job: the job name is appended as the last argument.
	// Defaults to `atkins -f <file>` or `task -t <file>`.
	Command []string `json:"-"`

	// Includes are the files the pipeline reads into its variables.
	Includes []string `json:"includes,omitempty"`
	// EnvFiles are the dotenv files the pipeline and its jobs read, relative
	// to Dir. These are the files task-ui can manage.
	EnvFiles []string `json:"env_files,omitempty"`

	// Allow lists the users and roles that may run this folder's jobs,
	// from task-ui.yml. Empty allows every runner.
	Allow []string `json:"allow,omitempty"`

	Jobs []*Job `json:"jobs"`

	// Error holds a parse error. The folder is listed so the problem is
	// visible, but it has no jobs.
	Error string `json:"error,omitempty"`
}

// Job returns the job by name or alias, or nil.
func (p *Project) Job(name string) *Job {
	if p == nil {
		return nil
	}
	for _, j := range p.Jobs {
		if j.Name == name || slices.Contains(j.Aliases, name) {
			return j
		}
	}
	return nil
}

// Job is one runnable job of a folder.
type Job struct {
	Project     string   `json:"project"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Summary     string   `json:"summary,omitempty"`
	Aliases     []string `json:"aliases,omitempty"`
	Interactive bool     `json:"interactive,omitempty"`

	// Steps are the step labels, the way `atkins -l` prints them.
	Steps []string `json:"steps,omitempty"`
	// DependsOn names the jobs that run first.
	DependsOn []string `json:"depends_on,omitempty"`
	// EnvFiles are the dotenv files this job reads on top of the pipeline's.
	EnvFiles []string `json:"env_files,omitempty"`

	// Allow narrows who may run this job, on top of the folder Allow,
	// from task-ui.yml.
	Allow []string `json:"allow,omitempty"`
}

// Key is the identity used for the run lock: one run per job per folder,
// across the instance.
func (j *Job) Key() string {
	return j.Project + "/" + j.Name
}

// Role names the permission sets a user can hold.
type Role string

const (
	// RoleAdmin may run anything and manage the instance settings.
	RoleAdmin Role = "admin"
	// RoleRunner may run jobs and read history.
	RoleRunner Role = "runner"
	// RoleViewer may read jobs and history, and run nothing.
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
	// Projects limits the folders the user sees. Empty means all.
	Projects []string `yaml:"projects,omitempty" json:"projects,omitempty"`
}

// Initials returns one or two letters that stand for the user in the
// header: the first letters of the first two words, or the first two
// letters of a single word.
func (u *User) Initials() string {
	if u == nil || u.Name == "" {
		return "?"
	}
	var words []string
	word := ""
	for _, r := range u.Name {
		switch {
		case r == ' ' || r == '.' || r == '-' || r == '_' || r == '@':
			if word != "" {
				words = append(words, word)
				word = ""
			}
		default:
			word += string(r)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	if len(words) == 0 {
		return "?"
	}
	initials := []rune{}
	if len(words) >= 2 {
		initials = append(initials, []rune(words[0])[0], []rune(words[1])[0])
	} else {
		runes := []rune(words[0])
		initials = append(initials, runes[0])
		if len(runes) > 1 {
			initials = append(initials, runes[1])
		}
	}
	out := ""
	for _, r := range initials {
		if r >= 'a' && r <= 'z' {
			r -= 'a' - 'A'
		}
		out += string(r)
	}
	return out
}

// CanSee reports whether the user may see a folder.
func (u *User) CanSee(project string) bool {
	if u == nil {
		return false
	}
	return len(u.Projects) == 0 || slices.Contains(u.Projects, project)
}

// CanRun reports whether the user may run the job of a folder. Admins may
// run everything they see; runners may run what the allow lists permit.
func (u *User) CanRun(p *Project, j *Job) bool {
	if u == nil || p == nil || j == nil || !u.CanSee(p.ID) {
		return false
	}
	switch u.Role {
	case RoleAdmin:
		return true
	case RoleRunner:
		return u.allowed(p.Allow) && u.allowed(j.Allow)
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
	Job      string        `json:"job"`
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

// GitInfo describes a folder that is a git checkout.
type GitInfo struct {
	Remote  string    `json:"remote"`
	Branch  string    `json:"branch"`
	Commit  string    `json:"commit"`
	Subject string    `json:"subject"`
	When    time.Time `json:"when"`
}

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

// Secret is one environment variable task-ui adds to runs: to every
// folder when Scope is "all", or to the folder Scope names.
type Secret struct {
	Scope string `json:"scope"`
	Key   string `json:"key"`
	Value string `json:"-"`
}

// EnvFile is a dotenv file a folder reads, as task-ui manages it: the
// content is what the UI edits, the file in the folder is generated from
// it, with the secrets, before each run.
type EnvFile struct {
	Project string `json:"project"`
	// Path is the file path relative to the folder, as the pipeline names it.
	Path string `json:"path"`
	// Managed is set once the content was saved from the UI.
	Managed bool `json:"managed"`
	// Exists is set when the file is present in the folder.
	Exists bool `json:"exists"`
	// Content is the managed content, or the folder's file when not managed.
	Content string `json:"-"`
	// Keys are the variable names the content declares.
	Keys []string `json:"keys"`
}
