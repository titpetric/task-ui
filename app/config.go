// Package app is the composition root: it reads the configuration, opens
// the database, builds the runner, the scheduler, the API and the front
// end, and serves them. Nothing else imports it.
package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/titpetric/task-ui/model"
)

// Config is the instance configuration. Every field comes from a TASK_UI_*
// environment variable, and a flag of the same meaning overrides it.
type Config struct {
	// Listen is the listen address. TASK_UI_LISTEN, -l.
	Listen string
	// Root is the workspace folder, the directory holding the folders
	// with jobs and the default workspace, _. TASK_UI_WORKSPACE,
	// --workspace, default <data>/workspaces.
	Root string
	// Discover lists every directory under Root with something to run.
	// TASK_UI_DISCOVER, default true.
	Discover bool
	// Htpasswd is the password file, htpasswd style, read for sign in:
	// .htpasswd in the data directory, set by no variable.
	Htpasswd string
	// Import holds the role lists of an older configuration,
	// TASK_UI_ADMINS, TASK_UI_VIEWERS and TASK_UI_AGENT, which the first
	// start over a database applies to its users once; nil without any of them
	// set. Storage holds the roles after.
	Import *RoleImport
	// DataDir holds the database, the recordings, the artifacts, the
	// password file and the workspace folder. TASK_UI_DATA, --data,
	// default /data when it exists, else ./data.
	DataDir string
	// DB is the database DSN. TASK_UI_DB, default <data>/task-ui.db.
	DB string
	// SSHDir is the key folder, where every key of the instance lives.
	// TASK_UI_SSH_DIR, default /keys when it exists, else ./keys.
	SSHDir string
	// HostData is the data directory as the host's docker daemon sees
	// it, for a run's docker run -v. TASK_UI_HOST_DATA; empty reads the
	// mounts of the container task-ui runs in.
	HostData string
	// SessionTTL is the session lifetime. TASK_UI_SESSION_TTL, default 12h.
	SessionTTL time.Duration
	// AutoLogin signs every request in as this user, without a password.
	// TASK_UI_AUTOLOGIN. For development and the screenshot scripts.
	AutoLogin string
	// Keep is how long a finished run stays attachable in memory.
	// TASK_UI_KEEP, default 15m.
	Keep time.Duration
	// Grace is how long a stopped process gets before the kill.
	// TASK_UI_GRACE, default 10s.
	Grace time.Duration
	// ComposeProject is the docker compose project task-ui runs in, whose
	// compose up and down ask first on the actions bar.
	// TASK_UI_COMPOSE_PROJECT; empty reads the compose label of the
	// container task-ui runs in, if it runs in one.
	ComposeProject string
	// URL is the public base link of the instance, such as
	// https://task.example.com, without a trailing slash: the links to
	// copy on a page start with it. TASK_UI_URL; empty derives it
	// from each request, through the proxy's forwarded headers.
	URL string
	// ArtifactHistory is how many versions the artifact history keeps
	// before it squashes to one. TASK_UI_ARTIFACT_HISTORY, default 50; 0
	// keeps every version.
	ArtifactHistory int
	// Now pins the clock the hours gate and the pages read, for the
	// screenshot scripts. TASK_UI_NOW, RFC 3339. Empty is the real time.
	Now time.Time
	// PrintBuildInfo prints the build and exits. -v.
	PrintBuildInfo bool
}

// Clock returns the clock: the pinned time, or the real one.
func (c *Config) Clock() func() time.Time {
	if c.Now.IsZero() {
		return time.Now
	}
	now := c.Now
	return func() time.Time { return now }
}

// timeZone names the zone of a time as the Instance page shows it: the
// location, or the zone's abbreviation for the local one, with its
// offset.
func timeZone(t time.Time) string {
	name, _ := t.Zone()
	if loc := t.Location().String(); loc != "Local" && loc != name {
		name = loc + ", " + name
	}
	if name == "UTC" {
		return name
	}
	return name + " (UTC" + t.Format("-07:00") + ")"
}

func env(name, fallback string) string {
	if value, ok := os.LookupEnv(name); ok && value != "" {
		return value
	}
	return fallback
}

// csv splits a comma separated list, trimmed, empties dropped.
func csv(text string) []string {
	var result []string
	for _, item := range strings.Split(text, ",") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}

func envBool(name string, fallback bool) bool {
	if parsed, err := strconv.ParseBool(os.Getenv(name)); err == nil {
		return parsed
	}
	return fallback
}

func envDuration(name string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	return d, nil
}

// gone are the variables of an older layout, with what took their
// place: task-ui reads none of them, and one set stops it at start.
var gone = map[string]string{
	"TASK_UI_ROOT":     "the folders live in TASK_UI_WORKSPACE, by default <data>/workspaces: move them there and unset TASK_UI_ROOT",
	"TASK_UI_HTPASSWD": "the password file is <data>/.htpasswd: move it there and unset TASK_UI_HTPASSWD",
}

// NewConfig reads the environment and the command line flags.
func NewConfig(args []string) (*Config, error) {
	return NewConfigAt("/", args)
}

// NewConfigAt reads the configuration as NewConfig does, with the
// default data and key folders under base in place of /: /data and
// /keys when they exist there, else ./data and ./keys.
func NewConfigAt(base string, args []string) (*Config, error) {
	for _, name := range model.SortedKeys(gone) {
		if _, ok := os.LookupEnv(name); ok {
			return nil, fmt.Errorf("%s is not read: %s", name, gone[name])
		}
	}
	cfg := &Config{
		Listen:    env("TASK_UI_LISTEN", ":3000"),
		Root:      env("TASK_UI_WORKSPACE", ""),
		Discover:  envBool("TASK_UI_DISCOVER", true),
		Import:    roleImport(),
		DataDir:   env("TASK_UI_DATA", existing(filepath.Join(base, "data"), "data")),
		DB:        env("TASK_UI_DB", ""),
		SSHDir:    env("TASK_UI_SSH_DIR", existing(filepath.Join(base, "keys"), "keys")),
		HostData:  env("TASK_UI_HOST_DATA", ""),
		AutoLogin: env("TASK_UI_AUTOLOGIN", ""),

		ComposeProject: env("TASK_UI_COMPOSE_PROJECT", ""),
		URL:            strings.TrimRight(strings.TrimSpace(env("TASK_UI_URL", "")), "/"),
	}
	var err error
	if cfg.SessionTTL, err = envDuration("TASK_UI_SESSION_TTL", 12*time.Hour); err != nil {
		return nil, err
	}
	if cfg.Keep, err = envDuration("TASK_UI_KEEP", 15*time.Minute); err != nil {
		return nil, err
	}
	if cfg.Grace, err = envDuration("TASK_UI_GRACE", 10*time.Second); err != nil {
		return nil, err
	}
	if cfg.ArtifactHistory, err = envInt("TASK_UI_ARTIFACT_HISTORY", 50); err != nil {
		return nil, err
	}
	if value := os.Getenv("TASK_UI_NOW"); value != "" {
		if cfg.Now, err = time.Parse(time.RFC3339, value); err != nil {
			return nil, fmt.Errorf("TASK_UI_NOW: %w", err)
		}
	}

	flags := pflag.NewFlagSet("task-ui", pflag.ContinueOnError)
	flags.StringVarP(&cfg.Listen, "listen", "l", cfg.Listen, "Listen address (TASK_UI_LISTEN)")
	flags.StringVar(&cfg.Root, "workspace", cfg.Root, "Workspace folder holding the folders, default <data>/workspaces (TASK_UI_WORKSPACE)")
	flags.StringVar(&cfg.DataDir, "data", cfg.DataDir, "Data directory: database, password file, workspaces (TASK_UI_DATA)")
	flags.BoolVarP(&cfg.PrintBuildInfo, "version", "v", false, "Print build info")
	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	return cfg, cfg.Validate()
}

// existing returns dir for a directory, else the fallback.
func existing(dir, fallback string) string {
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir
	}
	return fallback
}

// Validate fills the derived defaults and checks the workspace folder:
// the default one, under the data directory, it makes.
func (c *Config) Validate() error {
	if c.Root == "" {
		if c.DataDir == "" {
			return errors.New("TASK_UI_DATA: a data directory is needed for the workspace folder")
		}
		c.Root = c.WorkspacesDir()
		if err := os.MkdirAll(c.Root, 0o755); err != nil {
			return fmt.Errorf("TASK_UI_WORKSPACE: %w", err)
		}
	}
	if info, err := os.Stat(c.Root); err != nil || !info.IsDir() {
		return fmt.Errorf("TASK_UI_WORKSPACE: %s is not a directory", c.Root)
	}
	if c.Htpasswd == "" {
		c.Htpasswd = filepath.Join(c.DataDir, ".htpasswd")
	}
	if c.DB == "" {
		c.DB = filepath.Join(c.DataDir, "task-ui.db")
	}
	if c.SessionTTL <= 0 {
		return errors.New("TASK_UI_SESSION_TTL: must be positive")
	}
	return nil
}

// WorkspacesDir is the workspace folder by default, under the data
// directory.
func (c *Config) WorkspacesDir() string {
	return filepath.Join(c.DataDir, "workspaces")
}

// RecordingsDir is where the recordings go.
func (c *Config) RecordingsDir() string {
	return filepath.Join(c.DataDir, "recordings")
}

// ArtifactTree is the tree every run's artifacts go into, a git repository.
func (c *Config) ArtifactTree() string {
	return filepath.Join(c.ArtifactsDir(), "tree")
}

// ArtifactsDir is where the runs write their artifacts while they run.
func (c *Config) ArtifactsDir() string {
	return filepath.Join(c.DataDir, "artifacts")
}

// envInt reads a whole number from the environment, or the fallback.
func envInt(name string, fallback int) (int, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("%s: %q is not a whole number", name, value)
	}
	return n, nil
}
