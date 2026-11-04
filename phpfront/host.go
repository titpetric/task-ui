package phpfront

import (
	"context"
)

// Host answers what is not a row of the instance's database: what is derived
// from the filesystem, held in memory by a running process, or owned by a Go
// package for a reason. Every row is ported into code/Task instead.
//
// A nil Host answers every call empty.
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

// Place is a folder of the catalogue, or a directory inside one, in the depth
// the page about it draws.
//
// Found is false where the catalogue holds no such folder, or the directory
// inside it is not there to detect.
//
// ID is the folder's, which a directory inside it shares, and Rel is that
// directory under the folder, empty for the folder itself. Title is what the
// catalogue titles the folder, or folder/dir for a directory inside one.
type Place struct {
	Found       bool      `json:"found"`
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Dir         string    `json:"dir"`
	Rel         string    `json:"rel"`
	Allow       []string  `json:"allow"`
	Hours       string    `json:"hours"`
	Error       string    `json:"error"`
	Errors      []Finding `json:"errors"`
	Sources     []Source  `json:"sources"`

	// Includes are the files a place's pipeline reads besides its sources,
	// each with the lines it holds.
	Includes []File `json:"includes"`
}

// Source is one pipeline file a place was detected in: what convention it
// follows, the file, the command a target of it runs with, and the lines the
// file holds, which the button that shows it says.
type Source struct {
	Kind    string   `json:"kind"`
	File    string   `json:"file"`
	Command []string `json:"command"`
	Lines   int      `json:"lines"`
}

// File is a file of a place by name, with the lines it holds; a file the
// folder cannot read holds none.
type File struct {
	Name  string `json:"name"`
	Lines int    `json:"lines"`
}

// Group is one row of the Jobs table of a place: a pipeline file's targets,
// the files at the top of the folder that run, a skill's jobs, the terminals,
// or the targets of a directory inside.
//
// Folder is the folder the row's jobs run as, which gates every one of them,
// and is not always the place: the files run as the folder above a directory
// being browsed, and a directory inside with targets of its own runs as
// itself.
type Group struct {
	// Kind heads the row: the source's convention, files, @ and the skill's
	// ID, or built-in.
	Kind string `json:"kind"`

	// Note is what the row says on a hover after Kind: the file and the
	// command it runs with, or the skill's title.
	Note string `json:"note"`

	// Source is the pipeline file of a row of targets, as its folder names
	// it, and empty for every other row.
	Source string `json:"source"`

	// Dir is the directory inside the place a row of its targets runs in,
	// which the row's label links to; empty for the place's own.
	Dir string `json:"dir"`

	// Empty is what a row of a source with no targets says.
	Empty string `json:"empty"`

	// Targets marks a row of a source's targets, which is the one kind of
	// row whose label is a link.
	Targets bool  `json:"targets"`
	Folder  Where `json:"folder"`
	Jobs    []Job `json:"jobs"`
}

// Where is the folder a row's jobs run as: its ID and allow list, which gate
// them, the directory inside it they run in, and its hours, which decide
// whether a run needs forcing.
type Where struct {
	ID    string   `json:"id"`
	Rel   string   `json:"rel"`
	Allow []string `json:"allow"`
	Hours string   `json:"hours"`
}

// Entry is one entry of a directory the Files panel lists.
type Entry struct {
	Name string `json:"name"`
	Dir  bool   `json:"dir"`
	Size int64  `json:"size"`

	// ModTime is when the entry last changed, as Task\Clock reads a stamp.
	ModTime string `json:"mod_time"`

	// Path is the entry under its folder, and empty for one the folder
	// refuses to open.
	Path string `json:"path"`

	// Folder is the ID of the catalogue folder a directory entry is in its
	// own right, and empty for every other entry.
	Folder string `json:"folder"`

	// Sealed is an entry task-ui lists and does not open, and Exec a file
	// that runs: executable, or a shebang on its first line.
	Sealed bool `json:"sealed"`
	Exec   bool `json:"exec"`
}

// Status is what git reports of the checkout a place is in.
//
// Found is false where the place is in no checkout, or the instance has no
// root to name one against. Checkout is where the checkout starts under the
// root, empty for the checkout at the root itself.
type Status struct {
	Found    bool     `json:"found"`
	Checkout string   `json:"checkout"`
	Branch   string   `json:"branch"`
	Detached string   `json:"detached"`
	Upstream string   `json:"upstream"`
	Ahead    int      `json:"ahead"`
	Behind   int      `json:"behind"`
	Dirty    int      `json:"dirty"`
	Changes  []Change `json:"changes"`
}

// Change is one file a checkout has changed: the two letter code git writes
// it as, what that says in words, the path, and where a rename came from.
type Change struct {
	Code  string `json:"code"`
	State string `json:"state"`
	Path  string `json:"path"`
	From  string `json:"from"`
}

// Checkout is the commit a folder's checkout is at and the space it takes.
//
// Found is false for a folder in no checkout. When is the commit's moment as
// Task\Clock reads a stamp, and Space is empty where nothing measured it.
type Checkout struct {
	Found   bool   `json:"found"`
	Remote  string `json:"remote"`
	Branch  string `json:"branch"`
	Commit  string `json:"commit"`
	Subject string `json:"subject"`
	When    string `json:"when"`
	Space   Space  `json:"space"`
}

// Hours is an hours rule read: the one line a page prints it as, and whether
// it is open at the moment of the reading. A rule the parser refuses prints
// as itself and is always open, as a run of it is never refused.
type Hours struct {
	Line string `json:"line"`
	Open bool   `json:"open"`
}

// ToolingScope is one database atkins records to, as atkins reports it. The
// rollups and the charts over it are PHP's.
//
// From and To bound the window as dates, and To is the day atkins made the
// report, which the window of days counts back from.
type ToolingScope struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	Database   string `json:"database"`
	Days       int    `json:"days"`
	From       string `json:"from"`
	To         string `json:"to"`

	// Runs is how many commands ran in the window and TotalMs how long they
	// took between them.
	Runs    int64 `json:"runs"`
	TotalMs int64 `json:"total_ms"`

	// Commands are the command keys by total time, Usage a row per tool per
	// day and CommandUsage a row per command per day.
	Commands     []ToolingCommand `json:"commands"`
	Usage        []ToolingUsage   `json:"usage"`
	CommandUsage []ToolingUsage   `json:"command_usage"`

	// Samples are the commands of the window one by one, oldest first, which
	// a chart buckets finer than a day. An atkins that lists none leaves the
	// charts to the daily series.
	Samples []ToolingSample `json:"samples"`
}

// ToolingCommand is one command key and what it has cost, which the table
// behind the panel's toggle lists.
type ToolingCommand struct {
	Command   string `json:"command"`
	Tool      string `json:"tool"`
	Runs      int64  `json:"runs"`
	TotalMs   int64  `json:"total_ms"`
	AverageMs int64  `json:"average_ms"`
	MaxMs     int64  `json:"max_ms"`
	Failures  int64  `json:"failures"`

	// LastRunAt is when it last ran, as Task\Clock reads a stamp, and empty
	// for a command that has not run in the window.
	LastRunAt string `json:"last_run_at"`
}

// ToolingUsage is one name on one day: a tool of the usage series, or a
// command of the command series, which an older atkins does not report. A day
// the name did not run has no row.
type ToolingUsage struct {
	Date    string `json:"date"`
	Command string `json:"command"`
	Tool    string `json:"tool"`
	Runs    int64  `json:"runs"`
	TotalMs int64  `json:"total_ms"`
}

// ToolingSample is one command atkins recorded: when it started, its key and
// the job it ran under, how long it took and how it ended. It is named by its
// key and never by its command line, which can carry a token.
type ToolingSample struct {
	// At is when it started, as Task\Clock reads a stamp.
	At         string `json:"at"`
	Command    string `json:"command"`
	Tool       string `json:"tool"`
	Job        string `json:"job"`
	DurationMs int64  `json:"duration_ms"`
	ExitCode   int64  `json:"exit_code"`
	Failed     bool   `json:"failed"`
}

// TreeItem is one entry of a directory of the artifact tree.
type TreeItem struct {
	Name string `json:"name"`
	Dir  bool   `json:"dir"`
	Size int64  `json:"size"`

	// ModTime is when the entry last changed, as Task\Clock reads a stamp,
	// and empty for an entry with none.
	ModTime string `json:"mod_time"`
}

// Space is what a tree takes on disk: the files, the history behind them, and
// what git ignores, which the total leaves out.
//
// Measured is false on an instance with no meter, where a page says nothing
// about space rather than saying it takes none.
type Space struct {
	Measured bool  `json:"measured"`
	Files    int64 `json:"files"`
	History  int64 `json:"history"`
	Ignored  int64 `json:"ignored"`
}

// Inventory is the hosts page's facts.
//
// Off is an instance with no Beszel integration, which the page says instead
// of listing nothing, and Error what the hub answered when it answered badly.
type Inventory struct {
	Off    bool     `json:"off"`
	Error  string   `json:"error"`
	Hosts  []System `json:"hosts"`
	Logins []Login  `json:"logins"`

	// Cluster is the hosts summed into pools, from hostinfo.
	Cluster Cluster `json:"cluster"`

	// History is Beszel's records of each host, by host ID, which the
	// cluster's charts draw.
	History []Series `json:"history"`
}

// Login is the ssh login task-ui signs in to a host as.
type Login struct {
	Host  string `json:"host"`
	Login string `json:"login"`
}

// System is one host as Beszel reports it: the shares its line meters, and the
// sizes behind the memory and the disk.
//
// Measured is false where the hub holds no minute of measures, and then every
// size is zero rather than a reading.
type System struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Host   string `json:"host"`
	Status string `json:"status"`
	Up     bool   `json:"up"`
	CPUs   int    `json:"cpus"`
	Agent  string `json:"agent"`

	// Uptime is how long the host has run, as the host's page words it.
	Uptime string `json:"uptime"`

	// CPU, Memory and Disk are shares, 0 to 100.
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	Disk   float64 `json:"disk"`

	Measured    bool    `json:"measured"`
	MemoryTotal float64 `json:"memory_total"`
	MemoryUsed  float64 `json:"memory_used"`
	DiskTotal   float64 `json:"disk_total"`
	DiskUsed    float64 `json:"disk_used"`
}

// Cluster is the hosts summed into pools, as hostinfo computes it.
type Cluster struct {
	Up    int           `json:"up"`
	Down  int           `json:"down"`
	Left  int           `json:"left"`
	Pools []Pool        `json:"pools"`
	Hosts []ClusterHost `json:"hosts"`
}

// Pool is one resource the hosts share: the cores, the memory, the disks and
// the running containers.
type Pool struct {
	Name     string    `json:"name"`
	Unit     string    `json:"unit"`
	Figure   string    `json:"figure"`
	Used     float64   `json:"used"`
	Total    float64   `json:"total"`
	Percent  float64   `json:"percent"`
	Segments []Segment `json:"segments"`
}

// Segment is one host's share of a pool, as the pool's bar draws it.
type Segment struct {
	Host  string  `json:"host"`
	Color int     `json:"color"`
	Width float64 `json:"width"`
	Fill  float64 `json:"fill"`
	Title string  `json:"title"`
}

// ClusterHost is a host of the pools' legend. Left is a host the pools leave
// out, which is down or sends no measures.
type ClusterHost struct {
	Name  string `json:"name"`
	Color int    `json:"color"`
	Left  bool   `json:"left"`
}

// Series is Beszel's records of one host.
type Series struct {
	ID      string   `json:"id"`
	Samples []Sample `json:"samples"`
}

// Sample is one record of a host: the minute Beszel kept, as the charts read
// it. Sent and Recv are bytes a second.
type Sample struct {
	At         string  `json:"at"`
	CPU        float64 `json:"cpu"`
	Memory     float64 `json:"memory"`
	MemoryUsed float64 `json:"memory_used"`
	Disk       float64 `json:"disk"`
	DiskUsed   float64 `json:"disk_used"`
	Sent       float64 `json:"sent"`
	Recv       float64 `json:"recv"`

	// Extra are the filesystems the agent watches beside the root.
	Extra []Mount `json:"extra"`
}

// Mount is a filesystem of a record, by the name the agent set for it.
type Mount struct {
	Name  string  `json:"name"`
	Used  float64 `json:"used"`
	Total float64 `json:"total"`
}

// HostFacts is one host of the inventory in full.
//
// Found is false where Beszel monitors no such host, and Error what the hub
// answered when it refused.
type HostFacts struct {
	Found  bool   `json:"found"`
	Error  string `json:"error"`
	System System `json:"system"`
	Login  string `json:"login"`

	// Containers are the host's, busiest first. Measured says Beszel
	// reported them and Sized that docker ps answered over ssh; Unsized
	// says why it did not.
	Containers []Container `json:"containers"`
	Measured   bool        `json:"measured"`
	Sized      bool        `json:"sized"`
	Unsized    string      `json:"unsized"`

	// Filesystems are the durable ones; FromDF says they come from df over
	// ssh rather than from Beszel.
	Filesystems []Filesystem `json:"filesystems"`
	FromDF      bool         `json:"from_df"`

	// Read is when the host last answered over ssh, empty for never.
	Read string `json:"read"`

	// Samples are Beszel's records of the host, and Containers' of its
	// containers, which the charts draw.
	Samples          []Sample          `json:"samples"`
	ContainerSamples []ContainerSample `json:"container_samples"`
}

// Container is one container of a host, Beszel's measures and docker's sizes
// put together.
type Container struct {
	Name        string  `json:"name"`
	Image       string  `json:"image"`
	State       string  `json:"state"`
	Status      string  `json:"status"`
	CPU         float64 `json:"cpu"`
	Memory      float64 `json:"memory"`
	MemoryShare float64 `json:"memory_share"`
	Net         float64 `json:"net"`
	SizeRw      int64   `json:"size_rw"`
	SizeVirtual int64   `json:"size_virtual"`
	Measured    bool    `json:"measured"`
	Sized       bool    `json:"sized"`
}

// Filesystem is one durable filesystem of a host.
type Filesystem struct {
	Mount   string  `json:"mount"`
	Device  string  `json:"device"`
	Type    string  `json:"type"`
	Size    int64   `json:"size"`
	Used    int64   `json:"used"`
	Free    int64   `json:"free"`
	Percent float64 `json:"percent"`
}

// ContainerSample is one record of a host's containers.
type ContainerSample struct {
	At         string          `json:"at"`
	Containers []ContainerStat `json:"containers"`
}

// ContainerStat is one container of a record.
type ContainerStat struct {
	Name   string  `json:"name"`
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
}

// Folder is one folder of the catalogue.
type Folder struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Dir   string `json:"dir"`

	// Allow lists the users and roles that may run the folder's jobs, from
	// the configuration. Empty allows every runner. A gate takes it beside
	// the job's own, so neither stands alone.
	Allow []string `json:"allow"`

	// Sources are the files the conventions were detected in, relative to
	// the folder and in detection order, which a row names itself by.
	Sources []string `json:"sources"`

	// Error is what the sources failed to parse with, so a folder with a
	// broken file still lists and says what is wrong, and Errors the
	// findings of atkins' linter in its pipelines.
	Error  string    `json:"error"`
	Errors []Finding `json:"errors"`

	// Jobs are the jobs the folder declares, in source order.
	Jobs []Job `json:"jobs"`
}

// Finding is one thing atkins' linter said about a pipeline, which is why
// the pipeline lists no jobs.
type Finding struct {
	Path   string `json:"path"`
	Job    string `json:"job"`
	Issue  string `json:"issue"`
	Detail string `json:"detail"`
}

// Job is one job as a gate reads it: who may run it, and the two kinds that
// are gated on a permission of their own rather than on an allow list.
//
// Found is false where the catalogue resolves no such job, which is how a
// page learns the folder is gone, the directory inside it is gone, or the
// job left the Taskfile since the run - all three refuse in Go, where the
// job is a nil pointer.
//
// FolderAllow is the allow list of the folder the job was resolved in, and
// is answered by Host.Job alone: the catalogue carries a folder's own under
// the folder, and a run names a directory inside a folder whose folder the
// page that gates it does not hold.
// The fields past FolderAllow are the job as its own page and the Jobs table
// draw it, which the catalogue does not carry: a gate reads none of them and
// jobs() leaves them out, so a page that lists a catalogue costs nothing for
// them.
type Job struct {
	Found       bool     `json:"found"`
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Dir         string   `json:"dir"`
	Shell       bool     `json:"shell"`
	Host        bool     `json:"host"`
	Allow       []string `json:"allow"`
	FolderAllow []string `json:"folder_allow"`

	// Kind and Source name the convention the job came from and the file.
	Kind   string `json:"kind"`
	Source string `json:"source"`

	// SourceLines is the lines that file holds, which the button that shows
	// it says. A job's own page draws the button and the Jobs table does
	// not, so Host.Job answers it and Groups leaves it at zero.
	SourceLines int `json:"source_lines"`

	Description string `json:"description"`
	Summary     string `json:"summary"`

	// Command is the whole argv a run starts, and Steps the lines of a
	// target that runs several commands.
	Command []string `json:"command"`
	Steps   []string `json:"steps"`
	Aliases []string `json:"aliases"`

	// Depends are the targets this one runs first, each resolved to the job
	// of the same place where there is one.
	Depends []Depends `json:"depends"`

	// Interactive is a job that takes keyboard input, which runs in a
	// terminal, and TakesArgs one a run may add arguments to, which
	// ArgsCommand is the command they follow.
	Interactive bool   `json:"interactive"`
	TakesArgs   bool   `json:"takes_args"`
	ArgsCommand string `json:"args_command"`

	// Key is the identity of the run lock and the history: one run per job
	// per folder and directory.
	Key string `json:"key"`

	// Confirm is what a run of the job asks before it starts, which compose
	// up and down of the project task-ui itself runs in do. It is answered
	// by Groups alone, where the skill that holds the job is known.
	Confirm string `json:"confirm"`
}

// Depends is one target a job runs first: the name as the pipeline writes it,
// and the job of the same place it names where there is one, which the page
// links to.
type Depends struct {
	Text  string `json:"text"`
	Found bool   `json:"found"`
	ID    string `json:"id"`
	Name  string `json:"name"`
	Dir   string `json:"dir"`
}

// Running is one run in progress, by the key its job is listed under.
type Running struct {
	Key   string `json:"key"`
	ID    string `json:"id"`
	State string `json:"state"`
}

// Flow is one workflow run in progress, by the workflow it belongs to.
type Flow struct {
	WorkflowID string `json:"workflow_id"`
	RunID      string `json:"run_id"`
}

// Session is what the runner knows about a run that the table does not yet:
// the state of a transition it has not written, the status it reports, the
// windows attached to it and its place in the queue.
//
// Live is false where the runner holds no session for the ID, and then the
// row is the whole answer. Message is empty where the status says nothing and
// the run waits for nothing, so a page falls back to the column.
type Session struct {
	Live       bool   `json:"live"`
	Attached   int    `json:"attached"`
	Message    string `json:"message"`
	State      string `json:"state"`
	ExitCode   int64  `json:"exit_code"`
	Title      string `json:"title"`
	FinishedAt string `json:"finished_at"`
}

// Screen is the last frame of a run's output as the SVG a page embeds, with
// the lines of its whole output and whether the frame already shows them all.
//
// Lines is 0 and Complete false for a run still going, whose output may still
// scroll; Complete is true for a recording the retention sweep has removed,
// where the frame says so and there is nothing left to expand.
type Screen struct {
	SVG      string `json:"svg"`
	Lines    int    `json:"lines"`
	Complete bool   `json:"complete"`
}

// Rules is config/permissions.yml: the titles in file order, and the path
// lists the file decides. Paths and Checks come over already built, so the
// decision logic is the only part of permission/ that is ported.
type Rules struct {
	Rules       []Rule   `json:"rules"`
	Sections    []Rule   `json:"sections"`
	Permissions []Rule   `json:"permissions"`
	Paths       []string `json:"paths"`
	Checks      []string `json:"checks"`

	// Extras are the rules a section added of its own, which the section's
	// all does not stand for. The key is the path, <section>/<rule>, because
	// that is what a check names and what a policy sets; the title is the
	// rule's own, which the matrix composes with the section's title.
	Extras []Rule `json:"extras"`
}

// Rule is one titled key of the permission file.
type Rule struct {
	Key   string `json:"key"`
	Title string `json:"title"`
}

// Keystore is the key folder as the pages that list keys read it: the
// directory the keys are in, the instance's encryption key, and every key of
// the instance and its workspaces.
//
// Instance is the name of the encryption key, empty for an instance with
// none, and the key itself is in Keys: a page that draws the row draws it the
// way every other key row is drawn, off one list.
type Keystore struct {
	Dir      string `json:"dir"`
	Instance string `json:"instance"`
	Keys     []Key  `json:"keys"`
}

// KeyFolder is the .ssh folder of one project and the names of the private
// keys in it. Runs are given none of them, which is what the warning says.
type KeyFolder struct {
	Project string   `json:"project"`
	Dir     string   `json:"dir"`
	Keys    []string `json:"keys"`
}

// Key is one key of the key folder. Nothing here is material a run signs
// with: PublicKey is the authorized_keys line, which is public by purpose,
// and the private half stays in the keystore.
//
// Roles and Users are who gets the key in their runs, and Git marks the one
// git pulls and clones with. Workspace is the ID of the workspace a key
// belongs to, empty for an instance key, and Encrypts says the key stores
// secrets rather than signing in anywhere.
type Key struct {
	Name        string   `json:"name"`
	Purpose     string   `json:"purpose"`
	Scope       string   `json:"scope"`
	Algo        string   `json:"algo"`
	Bits        int      `json:"bits"`
	Size        int      `json:"size"`
	Fingerprint string   `json:"fingerprint"`
	Comment     string   `json:"comment"`
	PublicKey   string   `json:"public_key"`
	CreatedAt   string   `json:"created_at"`
	Roles       []string `json:"roles"`
	Users       []string `json:"users"`
	Git         bool     `json:"git"`
	Encrypts    bool     `json:"encrypts"`
	Workspace   string   `json:"workspace"`
	// Backup is the backup of the previous material a rotation kept,
	// backup/<name>, empty for none; until it is deleted the key serves both
	// and rotates no more.
	Backup            string `json:"backup"`
	BackupFingerprint string `json:"backup_fingerprint"`
	RotatedAt         string `json:"rotated_at"`
}

// Instance is what the server runs with, as the Instance page draws it: the
// configuration from the environment and the command line, less anything
// secret. A path shows; a token never does.
type Instance struct {
	URL      string `json:"url"`
	Listen   string `json:"listen"`
	Root     string `json:"root"`
	DataDir  string `json:"data_dir"`
	SSHDir   string `json:"ssh_dir"`
	Htpasswd string `json:"htpasswd"`
	// HostData is the data directory as the host's docker daemon sees it,
	// empty where it is the same.
	HostData string `json:"host_data"`
	Discover bool   `json:"discover"`
	// AutoLogin is the user every request signs in as, if any.
	AutoLogin string `json:"auto_login"`
	// TimeZone is the zone of the server's clock, which the hours and the
	// schedules read.
	TimeZone string `json:"time_zone"`
	// SessionTTL, Keep and Grace are the durations read, as they word.
	SessionTTL string `json:"session_ttl"`
	Keep       string `json:"keep"`
	Grace      string `json:"grace"`
	// ArtifactHistory is how many versions the history keeps, 0 for every
	// version.
	ArtifactHistory int `json:"artifact_history"`
	// ComposeProject is the compose project task-ui runs in, if any.
	ComposeProject string `json:"compose_project"`
	// Beszel says the inventory integration has a hub and a token.
	Beszel bool `json:"beszel"`
	// Backlog is how the slots and the line for them stand; Running is false
	// for an instance with no runner, whose page leaves the rows out.
	Running bool    `json:"running"`
	Backlog Backlog `json:"backlog"`
	// ParentEnv is the parent environment mode in effect, empty without a
	// runner, and Env the names of the server's own variables, never their
	// values, for an admin alone.
	ParentEnv string    `json:"parent_env"`
	Env       []EnvName `json:"env"`
	// EnvBase are the variables the None mode passes and EnvSensitive the
	// words that mark a name as a credential's, both the runner's own lists.
	EnvBase      []string `json:"env_base"`
	EnvSensitive []string `json:"env_sensitive"`
	// Measured says the Storage table applies: an admin's, and left out
	// where the instance measures nothing.
	Measured bool    `json:"measured"`
	Storage  Storage `json:"storage"`
}

// Backlog is how many runs do work at once and how many wait: Parallel the
// CPUs, Going and Waiting the runs now, Auto the automatic length of the
// line, Configured the one an admin set and Length the one in effect.
type Backlog struct {
	Parallel   int `json:"parallel"`
	Going      int `json:"going"`
	Waiting    int `json:"waiting"`
	Auto       int `json:"auto"`
	Configured int `json:"configured"`
	Length     int `json:"length"`
}

// EnvName is one variable of the server's own environment as the Parent
// environment table lists it: its name, whether the name reads as a
// credential's, and whether each mode forwards it.
type EnvName struct {
	Name      string `json:"name"`
	Sensitive bool   `json:"sensitive"`
	None      bool   `json:"none"`
	Auto      bool   `json:"auto"`
}

// Storage is what the instance keeps on disk, by item, as the Storage table
// of the Instance page draws it. Pending says a measure is under way and the
// rows are the last one's, which the page prints as an ellipsis.
type Storage struct {
	Rows    []StorageRow `json:"rows"`
	Total   StorageRow   `json:"total"`
	Pending bool         `json:"pending"`
}

// StorageRow is one item of the Storage table: what it is, where it is, and
// the space and the files it holds.
type StorageRow struct {
	Name    string `json:"name"`
	Where   string `json:"where"`
	Size    int64  `json:"size"`
	Files   int64  `json:"files"`
	Pending bool   `json:"pending"`
}

// Retention is what the areas of the Data retention table hold now.
//
// Off is an instance with no retention sweep, whose page draws no table at
// all: the areas are the sweep's and an instance that does not clean has none.
// Error is what the walk answered when it failed, which the page says instead
// of a table of nothing: counting the areas reads the recordings, the artifact
// history and the tool databases off disk, and any of those can refuse.
type Retention struct {
	Off   bool        `json:"off"`
	Error string      `json:"error"`
	Areas []AreaUsage `json:"areas"`
}

// AreaUsage is what one area of the Data retention table holds: the kind, as
// the rules name it, how many items, their bytes where the items have a size,
// and when the oldest of them is from, empty for an area with none.
type AreaUsage struct {
	Kind   string `json:"kind"`
	Count  int    `json:"count"`
	Bytes  int64  `json:"bytes"`
	Oldest string `json:"oldest"`
}

// Telemetry is what the instance records of its own work: the options in
// effect, the stored ones over the defaults, and where the dashboard is
// mounted.
//
// Wired says the instance has telemetry at all. Without it the page shows
// these same fields, which are what it would record, and saves nothing.
//
// Path is the mount the router read at startup, which is what the link on the
// page opens; OptionPath is the one the options hold, which a save changes
// and a restart applies.
//
// Error is what reading the stored options answered when it failed, which the
// page says instead of a form of defaults it would save over.
type Telemetry struct {
	Error            string   `json:"error"`
	Wired            bool     `json:"wired"`
	Path             string   `json:"path"`
	Enabled          bool     `json:"enabled"`
	ServiceName      string   `json:"service_name"`
	SampleRate       float64  `json:"sample_rate"`
	RingBufferSize   int      `json:"ring_buffer_size"`
	TopRequests      int      `json:"top_requests"`
	MaxSpansPerTrace int      `json:"max_spans_per_trace"`
	IgnorePaths      []string `json:"ignore_paths"`
	LiveStream       bool     `json:"live_stream"`
	RefreshInterval  int      `json:"refresh_interval"`
	AllowedNetworks  []string `json:"allowed_networks"`
	OptionPath       string   `json:"option_path"`
	CaptureLogs      bool     `json:"capture_logs"`
	TrackMemoryUse   bool     `json:"track_memory_use"`
	TrustRequestID   bool     `json:"trust_request_id"`
	// DebugSpanMinDuration is spelled as a duration, "1ms" or "500us",
	// which is what the form posts and what the stored row holds. The page
	// writes it into a text field and no PHP parses it.
	DebugSpanMinDuration string `json:"debug_span_min_duration"`
}

// Integration is one service of the Integrations settings as its form draws
// it: what the registry says of it, and what its settings hold now.
//
// Values are the plain settings as they stand, by field name, and a secret is
// never one of them: Set says a secret has a value, which is all a form shows
// of one, and an empty field on save keeps it.
type Integration struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	// Credential names what the steps lead to, for their heading, and is
	// empty for a service that brings none.
	Credential string `json:"credential"`
	// Steps are how to get that credential, in order, each a sentence or
	// two of markdown.
	Steps  []string           `json:"steps"`
	Fields []IntegrationField `json:"fields"`
	// Configured says every setting that is not optional has a value.
	Configured bool              `json:"configured"`
	Values     map[string]string `json:"values"`
	Set        map[string]bool   `json:"set"`
	// Checked are the roles and users a Who setting names, in the order it
	// names them.
	Checked []string `json:"checked"`
	// Enabled is the state of the switch of each setting that has one.
	Enabled map[string]bool `json:"enabled"`
	// Error is what reading the settings said, where that failed.
	Error string `json:"error"`
}

// IntegrationField is one setting of a service, as integration.Field holds it:
// a form draws the fields in this order.
type IntegrationField struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Help        string `json:"help"`
	Placeholder string `json:"placeholder"`
	// Secret keeps the value out of every page, Optional leaves the service
	// set up without one, and Who makes the value a list of roles and users
	// the form draws a box each for.
	Secret   bool `json:"secret"`
	Optional bool `json:"optional"`
	Who      bool `json:"who"`
	// Options are the values the setting takes, the first its default, which
	// the form shows as a list to pick from.
	Options []string `json:"options"`
	// Generated fills an empty secret on the first save, so its steps are
	// not the heading's credential.
	Generated bool `json:"generated"`
	// Env is the variable the setting sets in a run, and Default the value
	// it has while the field is empty.
	Env     string `json:"env"`
	Default string `json:"default"`
	// Toggle gives the setting a switch, DefaultOn starts it on, Empty sends
	// the variable with no value and Source takes it from another field; the
	// last two draw the switch alone.
	Toggle    bool   `json:"toggle"`
	DefaultOn bool   `json:"default_on"`
	Empty     bool   `json:"empty"`
	Source    string `json:"source"`
}

// Variables is what the Secrets page prints of the vault: the value of every
// plain env variable, by row ID, and what the walk of them said.
//
// A secret is not in Values at all. The page draws six dots for one, and the
// value it withholds does not cross a boundary to be withheld on the other
// side of it.
type Variables struct {
	Values map[string]string `json:"values"`
	// Error is what the walk said where the vault would not open a
	// variable, which is the first one it would not: the page prints it
	// above the table and lists every variable all the same.
	Error string `json:"error"`
}

// Variable is one variable as its form draws it: the value, and the error
// opening it gave.
//
// Content is empty for a secret, whose value the form never draws, and for a
// variable that holds nothing. Error is set where the vault would not open it,
// and the form answers the shared 404 instead of drawing a value it does not
// have.
type Variable struct {
	Content string `json:"content"`
	Error   string `json:"error"`
}

// EnvVar is one variable the integrations add to a run, as the Secrets page
// lists it under the stored ones.
//
// Redacted is a name that holds a credential: the page prints REDACTED for it
// and the value is empty here.
type EnvVar struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Redacted bool   `json:"redacted"`
}

// Plugin is one module the instance serves under /m, as the Sidebar settings
// list it beside the pages: its name in the path, the row's title, the line
// under its heading and the sprite symbol of its icon.
type Plugin struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// Project is one project directory under the src of a workspace, as the
// Projects settings list it: the ID a page of it answers under, the directory
// itself, the pipeline files its conventions were detected in, how many jobs
// it and the folders inside it hold, and what git knows of it.
//
// Checkout is false for a directory that is no checkout, and then the four
// fields of the commit hold nothing. Error is the parse error of a source
// that does not parse.
type Project struct {
	ID       string   `json:"id"`
	Dir      string   `json:"dir"`
	Sources  []string `json:"sources"`
	Jobs     int      `json:"jobs"`
	Error    string   `json:"error"`
	Checkout bool     `json:"checkout"`
	Remote   string   `json:"remote"`
	Branch   string   `json:"branch"`
	Commit   string   `json:"commit"`
	Subject  string   `json:"subject"`
	When     string   `json:"when"`
}
