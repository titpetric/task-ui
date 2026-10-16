package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/pflag"

	"github.com/titpetric/task-ui/folder"
)

// Config is the instance configuration. Every field comes from a TASK_UI_*
// environment variable, and a flag of the same meaning overrides it. The
// users and folder overrides come from the configuration file.
type Config struct {
	// Listen is the listen address. TASK_UI_LISTEN, -l.
	Listen string
	// Title names the instance. TASK_UI_TITLE.
	Title string
	// Root is the directory holding the folders. TASK_UI_ROOT, --root.
	Root string
	// File serves a single pipeline file as the only folder. TASK_UI_FILE, -f.
	File string
	// Discover lists every directory under Root with something to run.
	// TASK_UI_DISCOVER, default true.
	Discover bool
	// ConfigFile holds the users and folder overrides. TASK_UI_CONFIG,
	// --config, default task-ui.yml (or task-ui.yaml) in Root.
	ConfigFile string

	// History records every run. TASK_UI_HISTORY, default true.
	History bool
	// HistoryPath is the recordings directory. TASK_UI_HISTORY_PATH,
	// --history-output, default ./history.
	HistoryPath string

	// SSHDir is the directory the key settings manage. TASK_UI_SSH_DIR,
	// default $HOME/.ssh.
	SSHDir string
	// EnvPath holds the secrets and the managed env files.
	// TASK_UI_ENV_PATH, default ./env.
	EnvPath string

	// Secret signs the session cookies. TASK_UI_SECRET. Empty generates
	// one per process, so sessions do not survive a restart.
	Secret string
	// SessionTTL is the session lifetime. TASK_UI_SESSION_TTL, default 12h.
	SessionTTL time.Duration
	// AutoLogin signs every request in as this configured user, without
	// a password. TASK_UI_AUTOLOGIN. It is for development and for the
	// screenshot scripts, never for a deployment.
	AutoLogin string

	// Columns and Rows are the initial terminal size.
	Columns, Rows int

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
	if value, ok := os.LookupEnv(name); ok && value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return fallback
}

// NewConfig reads the environment and the command line flags.
func NewConfig(args []string) (*Config, error) {
	home, _ := os.UserHomeDir()

	cfg := &Config{
		Listen:      env("TASK_UI_LISTEN", ":3000"),
		Title:       env("TASK_UI_TITLE", ""),
		Root:        env("TASK_UI_ROOT", "."),
		File:        env("TASK_UI_FILE", ""),
		Discover:    envBool("TASK_UI_DISCOVER", true),
		ConfigFile:  env("TASK_UI_CONFIG", ""),
		History:     envBool("TASK_UI_HISTORY", true),
		HistoryPath: env("TASK_UI_HISTORY_PATH", "history"),
		SSHDir:      env("TASK_UI_SSH_DIR", filepath.Join(home, ".ssh")),
		EnvPath:     env("TASK_UI_ENV_PATH", "env"),
		Secret:      env("TASK_UI_SECRET", ""),
		AutoLogin:   env("TASK_UI_AUTOLOGIN", ""),
		Columns:     120,
		Rows:        30,
	}

	ttl, err := time.ParseDuration(env("TASK_UI_SESSION_TTL", "12h"))
	if err != nil {
		return nil, fmt.Errorf("TASK_UI_SESSION_TTL: %w", err)
	}
	cfg.SessionTTL = ttl

	flags := pflag.NewFlagSet("task-ui", pflag.ContinueOnError)
	flags.StringVarP(&cfg.Listen, "listen", "l", cfg.Listen, "Listen address (TASK_UI_LISTEN)")
	flags.StringVar(&cfg.Root, "root", cfg.Root, "Root directory (TASK_UI_ROOT)")
	flags.StringVarP(&cfg.File, "file", "f", cfg.File, "Serve a single pipeline file (TASK_UI_FILE)")
	flags.StringVar(&cfg.ConfigFile, "config", cfg.ConfigFile, "Users and folders config (TASK_UI_CONFIG)")
	flags.BoolVar(&cfg.History, "history-enable", cfg.History, "Record runs (TASK_UI_HISTORY)")
	flags.StringVar(&cfg.HistoryPath, "history-output", cfg.HistoryPath, "History directory (TASK_UI_HISTORY_PATH)")
	flags.BoolVarP(&cfg.PrintBuildInfo, "version", "v", false, "Print build info")
	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	return cfg, cfg.Validate()
}

// Validate checks the configuration and fills computed defaults.
func (c *Config) Validate() error {
	if c.ConfigFile == "" {
		for _, name := range []string{"task-ui.yml", "task-ui.yaml"} {
			path := filepath.Join(c.Root, name)
			if _, err := os.Stat(path); err == nil {
				c.ConfigFile = path
				break
			}
		}
	}
	if c.Columns <= 0 {
		c.Columns = 120
	}
	if c.Rows <= 0 {
		c.Rows = 30
	}
	if c.SessionTTL <= 0 {
		c.SessionTTL = 12 * time.Hour
	}
	return nil
}

// SecretsPath is where the secrets live under EnvPath.
func (c *Config) SecretsPath() string {
	return filepath.Join(c.EnvPath, "secrets")
}

// FilesPath is where the managed env files live under EnvPath.
func (c *Config) FilesPath() string {
	return filepath.Join(c.EnvPath, "files")
}

// discovery returns the catalogue for this configuration.
func (c *Config) discovery(file *FileConfig) *folder.Discovery {
	if c.File != "" {
		id := folder.SafeID(filepath.Base(filepath.Dir(mustAbs(c.File))))
		return &folder.Discovery{
			Root:   ".",
			Config: []folder.Config{{ID: id, Path: mustAbs(c.File)}},
		}
	}
	return &folder.Discovery{
		Root:     c.Root,
		Discover: c.Discover,
		Config:   file.Projects,
	}
}

func mustAbs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}
