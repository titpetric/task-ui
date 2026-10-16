package model

import (
	"slices"
	"time"
)

// Project is a folder: a working directory and the targets found in it.
// Each convention the folder follows is a Source, and every target of
// every source is a Job.
type Project struct {
	// ID is the URL safe name of the folder.
	ID string `json:"id"`
	// Title is the name the first source declares, or the folder name.
	Title string `json:"title"`
	// Description is what the first source says about itself.
	Description string `json:"description,omitempty"`

	// Dir is the working directory the jobs run in.
	Dir string `json:"-"`

	// Sources are the conventions detected in the folder, in detection
	// order: pipeline files first, then build files, then scripts.
	Sources []*Source `json:"sources"`

	// Includes are the files the sources read into their variables.
	Includes []string `json:"includes,omitempty"`
	// EnvFiles are the dotenv files the sources read, relative to Dir.
	// These are the files task-ui can manage.
	EnvFiles []string `json:"env_files,omitempty"`

	// Allow lists the users and roles that may run this folder's jobs,
	// from task-ui.yml. Empty allows every runner.
	Allow []string `json:"allow,omitempty"`

	// Jobs are every source's jobs, in source order.
	Jobs []*Job `json:"jobs"`

	// Error holds the parse errors of the sources, so a folder with a
	// broken file is listed and says what is wrong.
	Error string `json:"error,omitempty"`
}

// Source is one convention found in a folder: the file it was read from,
// the command that runs its targets, and the targets.
type Source struct {
	// Kind names the convention: atkins, task, make, npm, composer, script.
	Kind string `json:"kind"`
	// File is the file the targets were read from, relative to the folder.
	File string `json:"file"`
	// Name and Description are what the file says about itself, when the
	// convention has a place for that.
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	// Command is the argv prefix that runs a target, such as
	// `atkins -f atkins.yml` or `npm run`. A job's Args follow it.
	Command []string `json:"command"`

	Includes []string `json:"includes,omitempty"`
	EnvFiles []string `json:"env_files,omitempty"`

	Jobs []*Job `json:"jobs"`

	Error string `json:"error,omitempty"`
}

// Job returns the job by ID, name or alias, or nil. The ID is what URLs
// and history use; the name is the target as its file spells it.
func (p *Project) Job(id string) *Job {
	if p == nil {
		return nil
	}
	for _, j := range p.Jobs {
		if j.ID == id {
			return j
		}
	}
	for _, j := range p.Jobs {
		if j.Name == id || slices.Contains(j.Aliases, id) {
			return j
		}
	}
	return nil
}

// Job is one runnable target of a folder.
type Job struct {
	Project string `json:"project"`
	// ID is the job's name within the folder: the target name when the
	// folder follows one convention, kind:name when it follows several,
	// so `make build` and `npm run build` stay apart.
	ID   string `json:"id"`
	Name string `json:"name"`
	// Kind and Source name the convention and file the job came from.
	Kind   string `json:"kind"`
	Source string `json:"source"`

	Description string   `json:"description,omitempty"`
	Summary     string   `json:"summary,omitempty"`
	Aliases     []string `json:"aliases,omitempty"`
	Interactive bool     `json:"interactive,omitempty"`

	// Args follow the source command: the target name for most
	// conventions, the script path for a script.
	Args []string `json:"args"`
	// Command is the full argv, the source command and Args.
	Command []string `json:"command"`

	// Steps are the step labels, the way `atkins -l` prints them, or the
	// recipe lines of a make target, or the script of an npm target.
	Steps []string `json:"steps,omitempty"`
	// DependsOn names the jobs that run first.
	DependsOn []string `json:"depends_on,omitempty"`
	// EnvFiles are the dotenv files this job reads on top of the source's.
	EnvFiles []string `json:"env_files,omitempty"`

	// Allow narrows who may run this job, on top of the folder Allow,
	// from task-ui.yml.
	Allow []string `json:"allow,omitempty"`
}

// Key is the identity used for the run lock and the history: one run per
// job per folder, across the instance.
func (j *Job) Key() string {
	return j.Project + "/" + j.ID
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
