package phpfront

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
