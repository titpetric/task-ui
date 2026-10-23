package view

import (
	"sort"
	"strings"
	"time"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/workspace"
)

// Page is what a template binds to: the chrome every page shares, and
// the data of the page itself. A handler fills what its page shows and
// leaves the rest empty.
type Page struct {
	Title string
	// Path is the request path, for the active entry of the sidebar.
	Path string
	User *model.User
	// Permissions are what User may do, for Can.
	Permissions   permission.Set
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

	Folder *folder.Folder
	// Place is the folder of the directory browsed inside Folder, as
	// Folder.Sub reads it, when it has targets of its own.
	Place *folder.Folder
	Job   *folder.Job
	// Actions are the actions bar of the folder or the directory browsed.
	Actions *Actions
	// Prompt is the prompt box of the folder or the directory browsed,
	// for a user with prompt/edit and a Claude token.
	Prompt *Prompt
	// Summaries says the user may Run and summarize: prompt/edit, a
	// Claude token and prompts on.
	Summaries bool
	Checkout  *git.Info
	// Tree is what the folder, or the directory browsed in it, holds.
	Tree []folder.Item
	// Browse is the directory open inside the folder, or in the artifact
	// tree: "" for the top, else a path ending in "/".
	Browse string
	// Up is where the Files panel's .. goes: the directory or the folder
	// above. Empty shows no .. row.
	Up string
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
	// HostJobs are the jobs of the root that run a command on hosts, for
	// the job pickers of schedules and workflow steps: @host:* and
	// @host:<host> a host with an ssh login, for a user with hosts.
	HostJobs []*folder.Job
	// Summarize shows what Claude summarizes, Run and summarize on the
	// Inventory and Summary on runs and workflow runs: the Claude
	// integration has a token for the user.
	Summarize bool
	// Batch are the runs of the user's last Run on the inventory, one a
	// host, oldest first.
	Batch []Run

	Runs []Run
	// Pager links the runs of the page to the newer and the older ones,
	// nil when they fit on one page.
	Pager *Pager
	// Activity charts when the runs of the page ran, nil without any.
	Activity *Activity
	// Tooling is what atkins has recorded for the tree, nil without a
	// record of anything.
	Tooling   *Tooling
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
	// SecretGroups are the workspaces on the secrets pages, each with
	// its secrets and its folders.
	SecretGroups []SecretGroup
	// SecretWorkspace is the workspace the secrets page shows alone,
	// empty for every one.
	SecretWorkspace string
	// SecretReadCounts and SecretLastReads are how many runs read each
	// secret and the last of them, by secret ID, on the secrets page.
	SecretReadCounts map[string]int
	SecretLastReads  map[string]model.SecretRead
	// SecretReads are a secret's read log on its page, and on a run's
	// page the secrets the run got, by name.
	SecretReads []model.SecretRead
	// KeyWorkspaces are the rows of the keys page, one per workspace.
	KeyWorkspaces []KeyWorkspace
	Entries       []folder.Entry
	Projects      []model.Project
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
	// TreeFolders are the folders of the directories Tree lists, by the
	// directory's name, whose jobs the list holds beside them.
	TreeFolders map[string]*folder.Folder
	// TreeMenus are the Run menus of the rows Tree lists, by the entry's
	// name: a directory that is a folder or holds a job source, and a
	// file that is a job source.
	TreeMenus map[string]*RunMenu
	// KeyHolders are the roles and the users, and the keys each gets.
	KeyHolders []KeyHolder
	Key        *sshkey.Key
	// InstanceKey is the instance's encryption key, the services'
	// secrets', such as the integrations' settings, in InstanceKeyDir;
	// EncryptKey is a workspace's encryption key, on its pages.
	InstanceKey    *sshkey.Key
	InstanceKeyDir string
	EncryptKey     *sshkey.Key
	// KeyOverlaps are the keys held both as an encryption key and as an
	// SSH key, which runs and git leave out.
	KeyOverlaps []workspace.Overlap
	Users       []model.User
	// Account is the user a form or a profile shows, and AccountMCP
	// their MCP token.
	Account    *model.User
	AccountMCP MCPAccess
	// Profile is the rest of a user's profile.
	Profile *Profile
	// Matrices are the roles' policies on the permissions page.
	Matrices []Matrix
	// AllowChoices are the boxes of a project's "who may run", and
	// JobAllowChoices those of each job, by job ID.
	AllowChoices    []WhoChoice
	JobAllowChoices map[string][]WhoChoice
	// FolderChoices are the boxes of the user form's folders.
	FolderChoices []FolderChoice
	// CanAddUsers shows Create user: the server writes a password file.
	CanAddUsers bool
	Root        string
	// BaseURL is the public base link of the instance, without a
	// trailing slash, that the links to copy start with.
	BaseURL string

	// Form holds the values of a form the page shows again.
	Form   map[string]string
	Error  string
	Notice string
	Back   string
	Crumbs []Crumb
	// Step is the step of the settings the page belongs to, by name,
	// for the footer that links the steps; empty for a page of none.
	Step string
	// StepCounts say how much each step holds, by name, on the
	// Settings page.
	StepCounts map[string]string
	// Instance is what the server runs with, for the Instance step.
	Instance *Instance
	// Retention are the kinds of data the Data retention step keeps,
	// with what each holds and its rule.
	Retention []RetentionBlock
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

// RunFolderHref returns the page of the place a run ran in, its folder
// or the directory inside it, or "" for the internal folder of older
// runs, which has none.
func (p *Page) RunFolderHref(folderID, dir string) string {
	if model.IsInternal(folderID) {
		return ""
	}
	if dir != "" {
		return p.Href(folderID) + "/" + dir
	}
	return p.Href(folderID)
}

// RunJobHref returns the page of a run's job, or "" for a job with none:
// a prompt's, the Notes agent's, or one of the internal folder.
func (p *Page) RunJobHref(folderID, dir, job string) string {
	run := model.Run{Folder: folderID, Job: job}
	if run.IsPrompt() || run.IsNote() || model.IsInternal(folderID) {
		return ""
	}
	return p.RunFolderHref(folderID, dir) + "/" + job
}

// PlaceHref returns the page of the place a prompt ran in, as a note's
// Dir holds it: under the folder that holds it among the page's
// folders, or the root's for _. Empty for a place in no such folder.
func (p *Page) PlaceHref(place string) string {
	if place == model.Default {
		return "/f/" + model.Default
	}
	best := ""
	for _, f := range p.Folders {
		if (place == f.ID || strings.HasPrefix(place, f.ID+"/")) && len(f.ID) > len(best) {
			best = f.ID
		}
	}
	if best == "" {
		return ""
	}
	return p.Href(best) + strings.TrimPrefix(place, best)
}

// Links turns steps into links, as Chain does among the page's
// workspaces.
func (p *Page) Links(steps []Step) []Link {
	return Chain(steps, p.Workspaces)
}

// Can reports whether the signed in user may do what a path of
// `config/permissions.yml` names: {{ if $.Can "secret/edit" }}.
func (p *Page) Can(path string) bool {
	return p.permissions().Allowed(path)
}

// permissions returns Permissions when they are the user's, and nothing
// more than the user's role grants without a policy otherwise.
func (p *Page) permissions() permission.Set {
	if p.Permissions.User != p.User {
		return permission.Set{User: p.User}
	}
	return p.Permissions
}

// CanSettings reports whether the signed in user may see a page of the
// settings: any of the sections they list.
func (p *Page) CanSettings() bool {
	return p.Can("workspace/view") || p.Can("project/view") || p.Can("key/view") || p.Can("secret/view") || p.Can("integration/view") || p.Can("user/view")
}

// CanProfile reports whether the signed in user may open the profile
// of a user, by name: their own, or anyone's with user/view. A name no
// user holds, such as the agent a note names, has none.
func (p *Page) CanProfile(name string) bool {
	if _, ok := p.Roles[name]; !ok || p.User == nil {
		return false
	}
	return p.User.Name == name || p.Can("user/view")
}

// CanPermissions reports whether the menu of a user, by name, links
// their permissions: a user, for a signed in user with user/edit.
func (p *Page) CanPermissions(name string) bool {
	_, ok := p.Roles[name]
	return ok && p.Can("user/edit")
}

// UserNames returns the names of the users, sorted, for a filter.
func (p *Page) UserNames() []string {
	names := make([]string, 0, len(p.Roles))
	for name := range p.Roles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// CanRun reports whether the signed in user may run a job: a shell job
// takes the shell permission besides, and a command on a host the hosts
// permission alone.
func (p *Page) CanRun(f *folder.Folder, job *folder.Job) bool {
	if f == nil || job == nil || (job.IsShell() && !p.Can("shell")) {
		return false
	}
	if job.IsHost() {
		return p.Can("hosts")
	}
	return p.permissions().Run(f.ID, f.Allow, job.Allow)
}

// Summarizes reports whether a job's Run has |>summarize beside it: for
// a user with Summaries, on a job that runs to an end of its own, no
// shell, prompt, Notes agent or interactive job.
func (p *Page) Summarizes(job *folder.Job) bool {
	return p.Summaries && job != nil && !job.IsShell() && !job.IsPrompt() && !job.IsNote() && !job.Interactive
}

// Deletes reports whether the signed in user may delete a note: their
// own, or anyone's for an admin, with note/delete, or prompt/delete for
// a prompt.
func (p *Page) Deletes(note model.Note) bool {
	allowed := p.Can("note/delete")
	if note.IsPrompt() && note.ParentID == "" {
		allowed = p.Can("prompt/delete")
	}
	return allowed && p.User != nil && (note.UserName == p.User.Name || p.User.CanAdmin())
}

// Edits reports whether the signed in user may edit a note of a thread,
// with note/edit, as Thread.Edits says.
func (p *Page) Edits(note model.Note, t Thread) bool {
	return p.Can("note/edit") && t.Edits(p.User, note)
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
// the IDs of its folder and itself and the directory inside the folder
// it runs in, as a run names them.
func (p *Page) CanRunJob(folderID, dir, jobID string) bool {
	for _, f := range p.Folders {
		if f.ID == folderID {
			return p.CanRun(f, f.Sub(dir).Job(jobID))
		}
	}
	return false
}

// JobHref returns the path of a job's page: under its folder's, after
// the directory it runs in for a job of a directory inside the folder.
func (p *Page) JobHref(f *folder.Folder, job *folder.Job) string {
	if f == nil || job == nil {
		return ""
	}
	href := p.Href(f.ID) + "/"
	if job.Dir != "" {
		href += job.Dir + "/"
	}
	return href + job.ID
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

// SummaryMenu returns the menu of |>summarize: the jobs of m that Claude
// summarizes, without the interactive ones; Main the same when Claude
// summarizes it, else the only job remaining. Nil for none.
func (p *Page) SummaryMenu(m *RunMenu) *RunMenu {
	if m == nil {
		return nil
	}
	var jobs []*folder.Job
	for _, job := range m.Jobs {
		if p.Summarizes(job) {
			jobs = append(jobs, job)
		}
	}
	main := m.Main
	if !p.Summarizes(main) {
		main = nil
	}
	if main == nil && len(jobs) == 1 {
		main = jobs[0]
	}
	if main == nil && len(jobs) == 0 {
		return nil
	}
	return &RunMenu{Folder: m.Folder, Main: main, Jobs: jobs}
}

// CanRunMenu reports whether the user may run any job of a menu.
func (p *Page) CanRunMenu(m *RunMenu) bool {
	if m == nil {
		return false
	}
	if p.CanRun(m.Folder, m.Main) {
		return true
	}
	for _, job := range m.Jobs {
		if p.CanRun(m.Folder, job) {
			return true
		}
	}
	return false
}

// LastState is the class of a job's dot: how its last run ended, pending
// for none.
func (p *Page) LastState(job *folder.Job) string {
	if job == nil {
		return "pending"
	}
	if run := p.LastRun(job.Key()); run != nil {
		return run.State
	}
	return "pending"
}
