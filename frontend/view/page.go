package view

import (
	"strings"

	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/sshkey"
)

// Page is what a template binds to: the chrome every page shares, and
// the data of the page itself. A handler fills what its page shows and
// leaves the rest empty.
type Page struct {
	Title string
	// Path is the request path, for the active entry of the sidebar.
	Path          string
	User          *model.User
	LoginRequired bool
	Version       string

	// Folders are the folders the user may see, for the sidebar.
	Folders []*folder.Folder
	// Recents are the latest runs the user may see, for the sidebar.
	Recents []Run
	// Running holds the runs in progress by job key.
	Running map[string]Run
	// Last holds the latest run of each job the page lists, by job key.
	Last map[string]Run

	Folder   *folder.Folder
	Job      *folder.Job
	Checkout *git.Info

	Runs      []Run
	Run       *Run
	Artifacts []model.Artifact
	Schedules []model.Schedule
	EnvFiles  []envfile.File
	EnvFile   *envfile.File
	// SecretNames are the secrets that reach a folder, by name.
	SecretNames []string
	Secrets     []model.Secret
	Secret      *model.Secret
	Entries     []folder.Entry
	Projects    []model.Project
	Keys        []sshkey.Key
	KeysDir     string
	// Encryption names the keys the secrets are encrypted with.
	Encryption []string
	Users      []model.User
	// Account is the user a form edits.
	Account *model.User
	Root    string

	// Form holds the values of a form being shown again.
	Form   map[string]string
	Error  string
	Notice string
	Back   string
	Crumbs []Crumb
}

// CanRun reports whether the signed in user may run a job.
func (p *Page) CanRun(f *folder.Folder, job *folder.Job) bool {
	return f != nil && job != nil && p.User.CanRun(f.ID, f.Allow, job.Allow)
}

// RunningRun returns the run in progress of a job, or nil.
func (p *Page) RunningRun(key string) *Run {
	if run, ok := p.Running[key]; ok {
		return &run
	}
	return nil
}

// LastRun returns the latest run of a job, or nil.
func (p *Page) LastRun(key string) *Run {
	if run, ok := p.Last[key]; ok {
		return &run
	}
	return nil
}

// IsPath reports whether the page is at a path.
func (p *Page) IsPath(path string) bool {
	return p.Path == path
}

// Under reports whether the page is at or under a path.
func (p *Page) Under(prefix string) bool {
	return p.Path == prefix || strings.HasPrefix(p.Path, prefix+"/")
}

// WhoCanRun says who may run a job, from the allow lists.
func (p *Page) WhoCanRun(f *folder.Folder, job *folder.Job) string {
	allow := append(append([]string{}, f.Allow...), job.Allow...)
	if len(allow) == 0 {
		return "admin, runner"
	}
	return "admin, " + strings.Join(allow, ", ")
}
