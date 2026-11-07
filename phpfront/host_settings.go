package phpfront

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
	// SessionTTL, SessionMax, Keep and Grace are the durations read, as they word.
	SessionTTL string `json:"session_ttl"`
	SessionMax string `json:"session_max"`
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
