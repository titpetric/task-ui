package view

import (
	"strings"
	"time"

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
	// Now is the moment the page is for: the hours read against it.
	Now time.Time

	// Folders are the folders the user may see.
	Folders []*folder.Folder
	// SideFolders are the folders the sidebar lists: Folders less the
	// projects set not to show there.
	SideFolders []*folder.Folder
	// Recents are the latest runs of jobs and of workflows the user may
	// see, newest first, for the sidebar.
	Recents []Recent
	// SideWorkflows are the workflows the user may see, for the sidebar.
	SideWorkflows []Workflow
	// StepWorkflows are the workflows a step of the one edited may run:
	// those the user sees, less the one edited and those that run it.
	StepWorkflows []Workflow
	// Running holds the runs in progress by job key.
	Running map[string]Run
	// Last holds the latest run of each job the page lists, by job key.
	Last map[string]Run

	Folder   *folder.Folder
	Job      *folder.Job
	Checkout *git.Info
	// Tree is what the folder, or the directory browsed in it, holds.
	Tree []folder.Item
	// Browse is the directory open inside the folder, "" for the folder
	// itself, else a path ending in "/".
	Browse string

	Runs []Run
	// Activity charts when the runs of the page ran, nil without any.
	Activity  *Activity
	Run       *Run
	Artifacts []model.Artifact
	Schedules []model.Schedule
	Workflows []Workflow
	// Pipeline is a workflow's steps top to bottom, with the latest run's
	// states.
	Pipeline []Link
	// Control is the dashboard's control centre.
	Control  Control
	Workflow *Workflow
	// WorkflowRuns are a workflow's runs, newest first.
	WorkflowRuns []WorkflowRun
	EnvFiles     []envfile.File
	EnvFile      *envfile.File
	// SecretNames are the secrets that reach a folder, by name.
	SecretNames []string
	Secrets     []model.Secret
	Secret      *model.Secret
	Entries     []folder.Entry
	Projects    []model.Project
	// Project is the row a configure page edits, nil for a fresh folder.
	Project *model.Project
	Keys    []sshkey.Key
	KeysDir string
	// KeyFolders are the projects' .ssh/ folders with keys in them.
	KeyFolders []KeyFolder
	// KeyPolicies are the roles whose runs get each instance key, by
	// key name; Key is the key an edit form changes.
	KeyPolicies map[string][]string
	// KeyUsers are the users, by name, whose runs get each instance key.
	KeyUsers map[string][]string
	Key      *sshkey.Key
	// Encryption names the keys that encrypt the secrets.
	Encryption []string
	Users      []model.User
	// Account is the user a form edits.
	Account *model.User
	Root    string

	// Form holds the values of a form the page shows again.
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

// CanRunAny reports whether the signed in user may run a job of a folder.
func (p *Page) CanRunAny(f *folder.Folder) bool {
	for _, job := range f.Jobs {
		if p.CanRun(f, job) {
			return true
		}
	}
	return false
}

// CanRunJob reports whether the signed in user may run a job named by
// the IDs of its folder and itself, as a run names them.
func (p *Page) CanRunJob(folderID, jobID string) bool {
	for _, f := range p.Folders {
		if f.ID == folderID {
			return p.CanRun(f, f.Job(jobID))
		}
	}
	return false
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

// InHours reports whether the hours allow a run at the page's moment.
func (p *Page) InHours(text string) bool {
	now := p.Now
	if now.IsZero() {
		now = time.Now()
	}
	return InHours(text, now)
}

// WhoCanRun says who may run a job, from the allow lists.
func (p *Page) WhoCanRun(f *folder.Folder, job *folder.Job) string {
	allow := append(append([]string{}, f.Allow...), job.Allow...)
	if len(allow) == 0 {
		return "admin, runner"
	}
	return "admin, " + strings.Join(allow, ", ")
}
