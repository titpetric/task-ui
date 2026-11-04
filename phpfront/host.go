package phpfront

import (
	"context"
)

// Host answers what is not a row of the instance's database: what is derived
// from the filesystem, held in memory by a running process, or owned by a Go
// package for a reason. Every row is ported into code/Task instead.
//
// A nil Host answers every call empty.
//
// The shapes it answers with are grouped the way the bindings and the
// converters are: a folder's and the catalogue's in host_folder.go, the
// Inventory's in host_inventory.go, the tooling statistics' in
// host_tooling.go, the settings' in host_settings.go and the live runs' in
// host_run.go.
type Host interface {
	// Whoami is the name the Guard resolved, "" when nobody is signed in.
	Whoami(ctx context.Context) string

	// Folders is the catalogue: folder.Catalogue parses Taskfiles to find jobs.
	Folders(ctx context.Context) []Folder

	// Job is the job a run names, as folder.Folder resolves it: the jobs of
	// a directory inside a folder are not in the catalogue, and the four a
	// folder has without declaring them are not either.
	Job(ctx context.Context, folder, dir, id string) Job

	// Projects is the folder IDs of the directories in a named workspace's
	// src, which the sidebar lists beside the catalogue's own folders: a
	// checkout with no pipeline file is in no catalogue and still has a page.
	Projects(ctx context.Context, workspace string) []string

	// FolderDir is the directory a folder ID names: the one the catalogue
	// holds for it, else where the layout puts the project. Empty on an
	// instance with no root. The named workspaces say which first segment
	// is a workspace and which is a project of the root, the same list
	// Task\Route::folderHref reads, because they are rows and PHP holds
	// them.
	FolderDir(ctx context.Context, id string, workspaces []string) string

	// FolderDirs is FolderDir of many IDs at once, keyed by the ID: an ID
	// the layout places nowhere is absent from the answer rather than
	// present and empty.
	//
	// The folder tree of the chrome asks about every row it draws, which is
	// every pin and every project of every workspace, and one call each was
	// 102ms of the 151ms the tree cost on every page: the crossing converts
	// the workspace list again per call, for an answer that is a map lookup
	// behind it.
	FolderDirs(ctx context.Context, ids []string, workspaces []string) map[string]string

	// IsDir reports whether a directory inside a folder is there, which is
	// what a pin of a directory is resolved against. The folders sit outside
	// the directory the PHP may open.
	IsDir(ctx context.Context, folder, rel string) bool

	// Commit is the time of the last commit in a directory, in the shape a
	// timestamp column holds, in UTC, and empty for a directory in no
	// checkout or one whose checkout has no commit. The shape is the one
	// Task\Clock calls STORED, because the sidebar orders a workspace's
	// projects by this and compares two of them as text. Both front ends
	// read it through one cache, so a project costs one git process a
	// minute between them.
	Commit(ctx context.Context, dir string) string

	// Running is the runs in progress, which are in no table until they end.
	Running(ctx context.Context) []Running

	// Flows is the workflow runs in progress, which the workflow runner holds
	// in memory: a row left `running` by a process that died is not one, and
	// the sidebar and the workflows table mark neither.
	Flows(ctx context.Context) []Flow

	// Dirty counts changed and untracked files under a checkout directory.
	Dirty(ctx context.Context, dir string) int

	// Rules is config/permissions.yml, which is embedded in the binary.
	Rules(ctx context.Context) Rules

	// Publish sends an event on the stream every live label reads.
	Publish(ctx context.Context, topic string, payload map[string]any) error

	// Session is a run as the runner holds it, which is ahead of the row.
	Session(ctx context.Context, id string) Session

	// Frame is the last screen of a live run, as the SVG a page embeds.
	Frame(ctx context.Context, id string) string

	// Screen is the last screen of a run that has ended, replayed from its
	// recording, which a pipeline draws under every step. castPath and cols
	// are the run's columns, because the row is PHP's and the reader of the
	// cast file is not.
	Screen(ctx context.Context, id, castPath string, cols int) Screen

	// Expired reports whether a run names a recording that is not on disk.
	Expired(ctx context.Context, castPath, state string) bool

	// Claude reports whether the reader has a token for Claude, which is
	// what a page needs before it offers a summary.
	Claude(ctx context.Context) bool

	// Tree is one directory of the artifact tree, which sits outside the
	// directory the PHP may open. A directory's size is the space the files
	// under it take, measured the way the Space of the whole tree is.
	Tree(ctx context.Context, dir string) []TreeItem

	// Space is the space the artifact tree takes, from the meter both front
	// ends read, so neither walks the tree a second time.
	Space(ctx context.Context) Space

	// Spaces is the space each of the directories takes, by the directory,
	// through the same meter. The Projects settings measure every project
	// they list, a checkout and a plain directory alike, which Checkout
	// answers for neither: it measures the checkout a directory sits in and
	// a directory in none takes nothing.
	Spaces(ctx context.Context, dirs []string) map[string]Space

	// Inventory is the hosts page's facts: what Beszel monitors, the ssh
	// logins, the pools the hosts sum into and the month the charts draw.
	// hosts says the reader has the hosts permission, which is what makes
	// task-ui read a host rather than answer from what it kept.
	Inventory(ctx context.Context, hosts bool) Inventory

	// HostFacts is one host of the inventory in full: its containers as
	// Beszel and docker ps report them together, its durable filesystems,
	// and the records its charts draw.
	HostFacts(ctx context.Context, name string, hosts bool) HostFacts

	// Reachable is the hosts a command can be run on: the ones Beszel
	// monitors that task-ui has an ssh login to, largest capacity first.
	// A job picker offers one job per host and one over all of them, and
	// that is the whole of what it needs, which is why this is not the
	// Inventory: the page of a host reads the hub and the host, a picker
	// reads neither.
	Reachable(ctx context.Context) []string

	// Tooling is what atkins has recorded for the tree, a scope per database
	// it records to and the project's first. task-ui reads it by running
	// atkins, which is why it is not a row: a run of a job is one row here
	// and the commands inside it are rows there.
	Tooling(ctx context.Context) []ToolingScope

	// Place is a folder of the catalogue, or a directory inside one, in the
	// depth the page about it draws: what it says of itself, the pipeline
	// files it was detected in with the lines each holds, the hours it runs
	// in and what atkins' linter said of it. The catalogue carries none of
	// it, and a directory inside a folder is a detection of that directory
	// rather than an entry of the catalogue at all.
	Place(ctx context.Context, folder, rel string) Place

	// Groups is the Jobs table of a place: a row per pipeline file of it,
	// the files at its top that run, the skills atkins applies there, the
	// terminals, and a row per directory inside it that runs jobs of its
	// own. It is one walk of the directory and up to a couple of hundred
	// detections of the directories in it, behind the cache both front ends
	// share, which is why the whole table is one answer and not a row at a
	// time.
	//
	// seen is the folders the reader sees, shell and claude the two
	// permissions a terminal is gated on: all three are decided in PHP. A
	// directory that is a folder nobody sees runs here under the folder
	// above it, which is what Go draws for the same reader.
	Groups(ctx context.Context, folder, rel string, seen []string, shell, claude bool) []Group

	// Listing is a directory as the Files panel draws it: a folder's own, or
	// a directory inside it. The folders sit outside the directory the PHP
	// may open.
	//
	// Path is the entry under the folder, empty for one the folder refuses
	// to open: a sealed entry, or a symlink that leads out of it. Folder is
	// the ID of the catalogue folder a directory entry is in its own right,
	// of the ones seen names, because which directory is which folder is
	// decided by the directories themselves and not by their names.
	Listing(ctx context.Context, folder, rel string, seen []string) []Entry

	// RootTree is the root's own directories as the Files panel of the root's
	// page draws them, which no folder of the catalogue answers: the root is
	// where the catalogue was read from and is not in it.
	//
	// named says a workspace row names the root, which moves what it lists
	// from the tree itself down into the src of it; the binding joins that
	// itself, because where a workspace keeps its checkouts is the Go layout's
	// and not a path for the PHP to spell.
	//
	// Folder is read as Listing reads it, of the ones seen names. Path is
	// always empty: nothing of the root's own listing opens a file.
	RootTree(ctx context.Context, named bool, seen []string) []Entry

	// ModTime is when a directory last changed, as Task\Clock reads a stamp:
	// for a folder under a folder that the listing of the one above does not
	// hold, which the Files panel lists as a row of its own.
	ModTime(ctx context.Context, dir string) string

	// Status is what git reports of the checkout a place is in: its branch,
	// how far it is from the branch it tracks and every file it has changed,
	// which is the actions bar of a folder's page and of a directory inside
	// one.
	Status(ctx context.Context, dir string) Status

	// Checkout is the commit a folder's checkout is at and the space that
	// checkout takes, in one answer: the Folder panel draws the two
	// together, and the space is measured from the top of the checkout the
	// commit was read from.
	Checkout(ctx context.Context, dir string) Checkout

	// Hours is an hours rule read: the one line a page prints it as, and
	// whether it is open at this moment. The grammar is a Go parser and the
	// rule is a field of a pipeline file, so neither end of it is a row.
	Hours(ctx context.Context, text string) Hours

	// Keys is the key folder: where it is, the instance's encryption key and
	// every key in it, each in the depth a keys table draws.
	//
	// The public half of a key is a file beside the private one and the
	// private one never leaves the keystore, so this is the one way a page
	// reads a key at all: the rows name a key and hold no material.
	Keys(ctx context.Context) Keystore

	// KeyFolders is the .ssh folder of every project that holds a private
	// key of its own. It is a walk of the checkouts and no row of any
	// table, so a page that warns about one reads it here.
	KeyFolders(ctx context.Context) []KeyFolder

	// Instance is what the server runs with, as the Instance step of the
	// settings draws it: the environment and the command line, the runner's
	// line for a slot, the parent environment in effect and the integrations
	// that have a hub.
	//
	// admin is the reader's flag, which PHP decides: the names of the
	// server's own variables and the Storage table are an admin's, and an
	// instance answering a reader who is not one holds neither.
	Instance(ctx context.Context, admin bool) Instance

	// Usage is what each area of the Data retention table holds now: the
	// items, their bytes where they have a size, and the oldest of them.
	// The rules are rows and PHP reads them; what the rules apply to is a
	// walk of the recordings, the artifact history and the tool databases.
	Usage(ctx context.Context) Retention

	// Telemetry is the options the tracer runs on, the stored ones over the
	// defaults, and where the dashboard is mounted. The defaults read the
	// OIDA_* environment and the mount was read at startup, so neither is a
	// row; Wired says the instance has telemetry at all, which is what makes
	// the page editable.
	Telemetry(ctx context.Context) Telemetry

	// Services are the integrations as the Integrations settings draw them,
	// in the order the page lists them, which Task\Host::integrations
	// answers: each service of the Go registry, and its settings as they
	// stand.
	//
	// Neither half is a row PHP reads. The registry is Go data, down to
	// every field's label and the steps to its credential; the settings are
	// secrets encrypted with the instance key, and the key does not enter
	// PHP. No secret crosses: a form shows whether one has a value.
	Services(ctx context.Context) []Integration

	// Variables are the values the Secrets page prints, by row ID, and what
	// reading them said, which Task\Host::variables answers.
	//
	// The rows are PHP's: the scope, the name and whether a variable is a
	// secret are columns. The value is not. Every row's content is age
	// armour encrypted to the instance key, the key is the keystore's and
	// the keystore does not enter PHP, so what a variable holds is a fact
	// of the vault and crosses from here.
	//
	// A secret's value is withheld, as the page never prints one. The walk
	// stops at the first variable the vault will not open and says so in
	// Error, which the page prints above the table; a variable after that
	// one holds the armour the row holds, which the page then prints,
	// because that is what the Go page prints.
	Variables(ctx context.Context) Variables

	// Variable is one variable opened on its own, the value the form of it
	// draws and the error opening it gave, which Task\Host::variable
	// answers.
	//
	// It is Variables for one row and not a slice of it: the form opens the
	// row it edits, so a variable the list could not reach past still opens
	// here. A secret's value is withheld, as the form never draws one; an
	// error here is the shared 404, as a missing row is.
	Variable(ctx context.Context, id string) Variable

	// IntegrationEnv is what the integrations add to a run of the reader,
	// name and value in name order, which the Secrets page lists under the
	// variables.
	//
	// It is no row at all: it is the registry's env mapping over the
	// settings of every service that is set up and reaches this reader, by
	// their name and their role. A name that holds a credential is
	// redacted, which runner.Sensitive decides.
	IntegrationEnv(ctx context.Context) []EnvVar

	// Modules are the plugins the instance serves under /m, which the
	// Sidebar settings list beside the pages: a module is a tree on disk
	// that registered itself at startup and is in no table.
	Modules(ctx context.Context) []Plugin

	// Entries are the project directories under the src of each workspace,
	// as the Projects settings list them: with or without something to run,
	// with or without a remote. Which directories are inside a workspace,
	// and what each of them holds, is a walk of the root.
	//
	// workspaces is the folder of each workspace row, in the order the
	// settings list them, as Task\Workspace holds them: the rows are PHP's,
	// here as in FolderDir, and the walk is the part that is not.
	Entries(ctx context.Context, workspaces []string) []Project

	// Dispatch starts a Claude run over a comment and the action file
	// saying what to do with it, and answers the note that holds the
	// thread. The action is a path inside the front end's own tree, so a
	// page ships the markdown it dispatches beside itself.
	//
	// The folder is where in the artifact tree the run writes. A page that
	// gives each of its threads a folder gets a scratch space per thread;
	// empty is the top of the tree, which every run shares.
	//
	// A run is the runner's and not a row, and the prompt reads a file
	// outside the directory the PHP may open, which is why both halves are
	// here. The permissions are the reader's: a viewer's comment starts
	// nothing.
	Dispatch(ctx context.Context, comment, action, folder string) (string, error)

	// RunOf is the run a dispatch turned into, "" where there is none yet.
	//
	// A dispatch opens the note and starts the run behind it, in that
	// order, so the run does not exist when the dispatch answers. The page
	// records the note and asks here when it draws, which is later; by
	// then there is one. The empty answer draws a label for a run that has
	// not started.
	RunOf(ctx context.Context, note string) string
}
