package view

import (
	"strings"
	"time"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/diskusage"
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
	// Roles are the role of each user by name, for the colour of an
	// avatar; the body carries them for the avatars the script draws.
	Roles map[string]string

	// Folders are the folders the user may see.
	Folders []*folder.Folder
	// SideFolders are the folders the sidebar lists: Folders less the
	// projects set not to show there.
	SideFolders []*folder.Folder
	// Workspaces are the folders of the workspaces, each with its rank,
	// which the sidebar lists in that order, after the projects at the
	// root, with their projects under them.
	Workspaces map[string]int
	// Recents are the latest runs of jobs and of workflows the user may
	// see, newest first, for the sidebar.
	Recents []RecentGroup
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
	// Browse is the directory open inside the folder, or in the artifact
	// tree: "" for the top, else a path ending in "/".
	Browse string
	// Archive is where the directory the page shows downloads as a .tgz.
	Archive string
	// Space is how much the directory the page shows takes, files and
	// history; nil leaves the bar out.
	Space *diskusage.Usage
	// Spaces are the same for each project of the projects page, by ID.
	Spaces map[string]diskusage.Usage
	// Threads are the notes of the notes board, newest first, or the chat,
	// oldest first; Agent is who a note names to hand out work.
	Threads []Thread
	Agent   string
	// Integrations are the forms of Settings, Integrations.
	Integrations []Integration
	// Hosts are the inventory of the Beszel integration; HostsOff says
	// the integration is not set up.
	Hosts    []beszel.System
	HostsOff bool
	// Logins are who task-ui signs in to each host as over ssh, user@host
	// by the host's address; nil leaves the SSH column out.
	Logins map[string]string
	// Summarize shows what Claude summarizes, Run and summarize on the
	// Inventory and Summary on runs and workflow runs: the Claude
	// integration has a token for the user.
	Summarize bool

	Runs []Run
	// Pager links the runs of the page to the newer and the older ones,
	// nil when they fit on one page.
	Pager *Pager
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
	// WorkspaceList are the workspaces, for their page and the clone
	// form; Workspace is the one a configure page edits, and the one a
	// folder page takes a prompt against.
	WorkspaceList []model.Workspace
	Workspace     *model.Workspace
	// MoveFrom is the workspace a configure page's project is in, which
	// its Workspace select starts on; empty for a folder that does not
	// move. MoveLeft names the rows that stay behind on a move.
	MoveFrom string
	MoveLeft []string
	// Nested are the project rows inside each workspace, by its folder.
	Nested  map[string][]string
	Keys    []sshkey.Key
	KeysDir string
	// HomeKeys are the key pairs in ~/.ssh of the user task-ui runs as.
	HomeKeys    []sshkey.Key
	HomeKeysDir string
	// KeyFolders are the projects' .ssh/ folders with keys in them.
	KeyFolders []KeyFolder
	// KeyPolicies are the roles whose runs get each instance key, by
	// key name; Key is the key an edit form changes.
	KeyPolicies map[string][]string
	// KeyUsers are the users, by name, whose runs get each instance key.
	KeyUsers map[string][]string
	// SubFolders are the folders right under the folder a page shows.
	SubFolders []*folder.Folder
	// KeyHolders are the roles and the users, and the keys each gets.
	KeyHolders []KeyHolder
	Key        *sshkey.Key
	// Encryption names the keys that encrypt the secrets, and
	// EncryptKey is a workspace's encryption key, on its pages.
	Encryption []string
	EncryptKey *sshkey.Key
	Users      []model.User
	// Account is the user a form edits.
	Account *model.User
	// AllowChoices are the boxes of a project's "who may run", and
	// JobAllowChoices those of each job, by job ID.
	AllowChoices    []WhoChoice
	JobAllowChoices map[string][]WhoChoice
	// FolderChoices are the boxes of the user form's folders.
	FolderChoices []FolderChoice
	// CanAddUsers shows Create user: the server writes a password file.
	CanAddUsers bool
	Root        string

	// Form holds the values of a form the page shows again.
	Form   map[string]string
	Error  string
	Notice string
	Back   string
	Crumbs []Crumb
}

// Groups are the sidebar's projects, and the workspaces with theirs.
func (p *Page) Groups() []Group {
	return Groups(p.SideFolders, p.Workspaces)
}

// Href returns the path of a folder's page, as FolderHref places it
// among the page's workspaces.
func (p *Page) Href(id string) string {
	return FolderHref(id, p.Workspaces)
}

// Links turns steps into links, as Chain does among the page's
// workspaces.
func (p *Page) Links(steps []Step) []Link {
	return Chain(steps, p.Workspaces)
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
