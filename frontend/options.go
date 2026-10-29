package frontend

import (
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/hostinfo"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/phpscript"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/retention"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/sshprobe"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/tooling"
	"github.com/titpetric/task-ui/workflow"
)

// Options are what the pages read from.
type Options struct {
	Guard *access.Guard
	// Auth checks the sign in form.
	Auth *access.Authenticator

	Users    *storage.UserStorage
	Projects *storage.ProjectStorage
	// Workspaces are the directories the projects may clone inside.
	Workspaces *storage.WorkspaceStorage
	Runs       *storage.RunStorage
	Artifacts  *storage.ArtifactStorage
	Schedules  *storage.ScheduleStorage
	// FolderRefs find the rows that name a folder, which a move leaves.
	FolderRefs *storage.FolderRefStorage
	Secrets    *secret.Store
	// SecretReads log which run got which secret. Nil shows no reads.
	SecretReads *storage.SecretReadStorage
	// Pins keep what each user pins to the sidebar; nil pins nothing.
	Pins *storage.PinStorage
	// Sidebar is the sidebar's top block, in the order it lists them.
	// Nil falls back to the fixed list the block held before the table.
	Sidebar *storage.SidebarStorage
	// Modules serves the plugins written in PHP, under /m. Nil is an
	// instance with none, and every module path answers 404.
	Modules *phpscript.Manager
	// Telemetry is the tracer and its settings, for /settings/telemetry.
	// Nil leaves that page saying telemetry is off.
	Telemetry Telemetry

	Workflows    *storage.WorkflowStorage
	WorkflowRuns *storage.WorkflowRunStorage
	// Flows runs the workflows.
	Flows *workflow.Runner

	Runner    *runner.Manager
	Catalogue folder.Catalogue
	// Steps lists the folders a workflow step picks, every folder with a
	// job; nil picks from Catalogue.
	Steps folder.Catalogue
	Git   *git.Git
	// Keys are the key folder: every key of the instance and its
	// workspaces.
	Keys *keystore.Local
	// Recordings holds the runs' recordings, for the last screen of a
	// run whose session has ended.
	Recordings *recording.Store
	// Tooling asks atkins what the tree has been running, for the panel
	// on the runs page. Nil shows no panel.
	Tooling *tooling.Prober
	// Root is where the folders are, for the projects page.
	Root    string
	Version string
	// Now is the clock the pages read for the hours. Nil is time.Now.
	Now func() time.Time
	// ArtifactStore is the tree of every run's artifacts the artifacts
	// page lists and serves. Nil shows the records only.
	ArtifactStore *artifactstore.Store
	// Disk measures the space the artifacts and the checkouts take. Nil
	// leaves the space bars out.
	Disk *diskusage.Meter
	// Storage measures what the instance keeps on disk for the Storage
	// table of the Instance page, an admin's. Nil leaves the table out.
	Storage *diskusage.Inventory
	// ComposeProject is the docker compose project task-ui runs in: up
	// and down of it on the actions bar ask first.
	ComposeProject string
	// BaseURL is the public base link, TASK_UI_URL; empty derives it
	// from each request.
	BaseURL string
	// AddUsers lets an admin make users: the server writes a password
	// file.
	AddUsers bool
	// Board is the notes board and the chat; nil turns them off.
	Board *note.Board
	// Integrations keep the integrations' settings, and Hosts reads the
	// host inventory through the Beszel one.
	Integrations *integration.Store
	Hosts        *beszel.Inventory
	// SSH finds who task-ui signs in to each host as, for the Hosts page;
	// nil leaves the column out.
	SSH *sshprobe.Prober
	// Facts reads the hosts' filesystems and containers over their ssh
	// logins, for the Inventory page; nil leaves them to Beszel.
	Facts *hostinfo.Reader
	// Instance is what the server runs with, for the Instance step of
	// the settings; nil shows an empty page.
	Instance *view.Instance
	// Tools are the tools the instance found in its PATH at start, for
	// the Tooling section of the Instance page; nil leaves it empty.
	Tools *storage.ToolStorage
	// Cleaner measures what the instance keeps and holds the Data
	// retention rules; nil shows an empty page.
	Cleaner *retention.Cleaner
}
