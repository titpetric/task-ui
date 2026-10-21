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
)

// Config is the instance configuration. Every field comes from a TASK_UI_*
// environment variable, and a flag of the same meaning overrides it.
type Config struct {
	// Listen is the listen address. TASK_UI_LISTEN, -l.
	Listen string
	// Root is the directory holding the folders. TASK_UI_ROOT, --root.
	Root string
	// Discover lists every directory under Root with something to run.
	// TASK_UI_DISCOVER, default true.
	Discover bool
	// Htpasswd is the password file, htpasswd style, read for sign in.
	// TASK_UI_HTPASSWD, --htpasswd, default .htpasswd in Root when it
	// exists.
	Htpasswd string
	// Admins and Viewers name the users with those roles, comma
	// separated. TASK_UI_ADMINS, TASK_UI_VIEWERS. Everyone else in the
	// file is a runner.
	Admins  []string
	Viewers []string
	// DataDir holds the database, the recordings and the artifacts.
	// TASK_UI_DATA, --data, default ./data.
	DataDir string
	// DB is the database DSN. TASK_UI_DB, default <data>/task-ui.db.
	DB string
	// SSHDir is the directory the key settings manage. TASK_UI_SSH_DIR,
	// default $HOME/.ssh.
	SSHDir string
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
	// Agent is the user a note on the board names with @ to make it an
	// instruction to run. TASK_UI_AGENT; empty turns instructions off.
	Agent string
	// AgentJob is the job an instruction runs, as folder/job, as the
	// agent's user. TASK_UI_AGENT_JOB, default internal/claude-note; set
	// empty, it leaves instructions open for an agent that reads the
	// board over MCP.
	AgentJob string
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

func env(name, fallback string) string {
	if value, ok := os.LookupEnv(name); ok && value != "" {
		return value
	}
	return fallback
}

// envSet reads a variable whose empty value means something of its own:
// the fallback applies only to a variable not set at all.
func envSet(name, fallback string) string {
	if value, ok := os.LookupEnv(name); ok {
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

// NewConfig reads the environment and the command line flags.
func NewConfig(args []string) (*Config, error) {
	home, _ := os.UserHomeDir()
	cfg := &Config{
		Listen:    env("TASK_UI_LISTEN", ":3000"),
		Root:      env("TASK_UI_ROOT", "."),
		Discover:  envBool("TASK_UI_DISCOVER", true),
		Htpasswd:  env("TASK_UI_HTPASSWD", ""),
		Admins:    csv(env("TASK_UI_ADMINS", "")),
		Viewers:   csv(env("TASK_UI_VIEWERS", "")),
		DataDir:   env("TASK_UI_DATA", "data"),
		DB:        env("TASK_UI_DB", ""),
		SSHDir:    env("TASK_UI_SSH_DIR", filepath.Join(home, ".ssh")),
		AutoLogin: env("TASK_UI_AUTOLOGIN", ""),
		Agent:     env("TASK_UI_AGENT", ""),
		AgentJob:  envSet("TASK_UI_AGENT_JOB", "internal/claude-note"),
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
	flags.StringVar(&cfg.Root, "root", cfg.Root, "Root directory holding the folders (TASK_UI_ROOT)")
	flags.StringVar(&cfg.Htpasswd, "htpasswd", cfg.Htpasswd, "Password file, htpasswd style (TASK_UI_HTPASSWD)")
	flags.StringVar(&cfg.DataDir, "data", cfg.DataDir, "Data directory: database, recordings, artifacts (TASK_UI_DATA)")
	flags.BoolVarP(&cfg.PrintBuildInfo, "version", "v", false, "Print build info")
	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	return cfg, cfg.Validate()
}

// Validate fills the derived defaults and checks the root.
func (c *Config) Validate() error {
	if c.Root == "" {
		return errors.New("TASK_UI_ROOT: a root directory is needed")
	}
	if info, err := os.Stat(c.Root); err != nil || !info.IsDir() {
		return fmt.Errorf("TASK_UI_ROOT: %s is not a directory", c.Root)
	}
	if c.Htpasswd == "" {
		if _, err := os.Stat(filepath.Join(c.Root, ".htpasswd")); err == nil {
			c.Htpasswd = filepath.Join(c.Root, ".htpasswd")
		}
	}
	if c.DB == "" {
		c.DB = filepath.Join(c.DataDir, "task-ui.db")
	}
	if c.SessionTTL <= 0 {
		return errors.New("TASK_UI_SESSION_TTL: must be positive")
	}
	return nil
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
