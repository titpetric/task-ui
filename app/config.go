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
	// UsersFile lists the users to seed. TASK_UI_CONFIG, --config, default
	// task-ui.yml in Root when it exists.
	UsersFile string
	// DataDir holds the database, the recordings and the artifacts.
	// TASK_UI_DATA, --data, default ./data.
	DataDir string
	// DB is the database DSN. TASK_UI_DB, default <data>/task-ui.db.
	DB string
	// SSHDir is the directory the key settings manage. TASK_UI_SSH_DIR,
	// default $HOME/.ssh.
	SSHDir string
	// Secret signs the session cookies. TASK_UI_SECRET. Empty generates
	// one per process, so sessions do not survive a restart.
	Secret string
	// SessionTTL is the session lifetime. TASK_UI_SESSION_TTL, default 12h.
	SessionTTL time.Duration
	// AutoLogin signs every request in as this user, without a password.
	// TASK_UI_AUTOLOGIN. For development and the screenshot scripts.
	AutoLogin string
	// Keep is how long a finished run stays attachable in memory.
	// TASK_UI_KEEP, default 15m.
	Keep time.Duration
	// Grace is how long a stopped process gets before it is killed.
	// TASK_UI_GRACE, default 10s.
	Grace time.Duration
	// PrintBuildInfo prints the build and exits. -v.
	PrintBuildInfo bool
}

func env(name, fallback string) string {
	if value, ok := os.LookupEnv(name); ok && value != "" {
		return value
	}
	return fallback
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
		UsersFile: env("TASK_UI_CONFIG", ""),
		DataDir:   env("TASK_UI_DATA", "data"),
		DB:        env("TASK_UI_DB", ""),
		SSHDir:    env("TASK_UI_SSH_DIR", filepath.Join(home, ".ssh")),
		Secret:    env("TASK_UI_SECRET", ""),
		AutoLogin: env("TASK_UI_AUTOLOGIN", ""),
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

	flags := pflag.NewFlagSet("task-ui", pflag.ContinueOnError)
	flags.StringVarP(&cfg.Listen, "listen", "l", cfg.Listen, "Listen address (TASK_UI_LISTEN)")
	flags.StringVar(&cfg.Root, "root", cfg.Root, "Root directory holding the folders (TASK_UI_ROOT)")
	flags.StringVar(&cfg.UsersFile, "config", cfg.UsersFile, "Users file to seed from (TASK_UI_CONFIG)")
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
	if c.UsersFile == "" {
		for _, name := range []string{"task-ui.yml", "task-ui.yaml"} {
			if _, err := os.Stat(filepath.Join(c.Root, name)); err == nil {
				c.UsersFile = filepath.Join(c.Root, name)
				break
			}
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

// ArtifactsDir is where the artifacts go.
func (c *Config) ArtifactsDir() string {
	return filepath.Join(c.DataDir, "artifacts")
}
