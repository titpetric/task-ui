package server

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"

	"github.com/titpetric/task-ui/server/model"
	"github.com/titpetric/task-ui/server/repository"
)

// Config is the instance configuration. Every field comes from a TASK_UI_*
// environment variable, and a flag of the same meaning overrides it. The
// users and project overrides come from the configuration file.
type Config struct {
	// Listen is the listen address. TASK_UI_LISTEN, -l.
	Listen string
	// Title names the instance in the header. TASK_UI_TITLE.
	Title string
	// Root is the folder holding the projects. TASK_UI_ROOT, --root.
	Root string
	// File runs a single task file as the only project. TASK_UI_FILE, -f.
	File string
	// Discover lists every folder under Root with a task file as a
	// project. TASK_UI_DISCOVER, default true.
	Discover bool
	// ConfigFile holds the users and project overrides. TASK_UI_CONFIG,
	// --config, default task-ui.yml (or task-ui.yaml) in Root.
	ConfigFile string

	// History records every run. TASK_UI_HISTORY, default true.
	History bool
	// HistoryPath is the recordings folder. TASK_UI_HISTORY_PATH,
	// --history-output, default ./history.
	HistoryPath string

	// SSHDir is the folder the key admin manages. TASK_UI_SSH_DIR,
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
	// Anonymous is the role of a visitor when no users are configured.
	// TASK_UI_ANONYMOUS, default admin, so a local instance works without
	// setup. Set it to viewer for a read only instance, or configure users.
	Anonymous model.Role

	// Columns and Rows are the initial terminal size.
	Columns, Rows int

	PrintBuildInfo bool
}

// FileConfig is the shape of task-ui.yml.
type FileConfig struct {
	Title    string                     `yaml:"title"`
	Users    []model.User               `yaml:"users"`
	Projects []repository.ProjectConfig `yaml:"projects"`
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
		Anonymous:   model.Role(env("TASK_UI_ANONYMOUS", string(model.RoleAdmin))),
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
	flags.StringVar(&cfg.Root, "root", cfg.Root, "Projects folder (TASK_UI_ROOT)")
	flags.StringVarP(&cfg.File, "file", "f", cfg.File, "Serve a single task file (TASK_UI_FILE)")
	flags.StringVar(&cfg.ConfigFile, "config", cfg.ConfigFile, "Users and projects config (TASK_UI_CONFIG)")
	flags.BoolVar(&cfg.History, "history-enable", cfg.History, "Record runs (TASK_UI_HISTORY)")
	flags.StringVar(&cfg.HistoryPath, "history-output", cfg.HistoryPath, "History folder (TASK_UI_HISTORY_PATH)")
	flags.BoolVarP(&cfg.PrintBuildInfo, "version", "v", false, "Print build info")
	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	return cfg, cfg.Validate()
}

// Validate checks the configuration and fills computed defaults.
func (c *Config) Validate() error {
	if c.Anonymous == "" {
		c.Anonymous = model.RoleViewer
	}
	if c.Anonymous != "none" && !c.Anonymous.Valid() {
		return fmt.Errorf("TASK_UI_ANONYMOUS: unknown role %q, use admin, runner, viewer or none", c.Anonymous)
	}
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
	return nil
}

// Projects returns the project repository for this configuration.
func (c *Config) Projects(file *FileConfig) *repository.Projects {
	if c.File != "" {
		id := repository.SafeID(filepath.Base(filepath.Dir(mustAbs(c.File))))
		return &repository.Projects{
			Root:   ".",
			Config: []repository.ProjectConfig{{ID: id, Path: mustAbs(c.File)}},
		}
	}
	return &repository.Projects{
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

// fileConfig loads task-ui.yml and reloads it when it changes, so users can
// be added or removed without a restart.
type fileConfig struct {
	path string

	mu       sync.Mutex
	modified time.Time
	current  *FileConfig
}

func (f *fileConfig) Get() *FileConfig {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.current == nil {
		f.current = &FileConfig{}
	}
	if f.path == "" {
		return f.current
	}

	info, err := os.Stat(f.path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) || !f.modified.IsZero() {
			slog.Warn("config file unavailable", "path", f.path, "err", err)
		}
		return f.current
	}
	if info.ModTime().Equal(f.modified) {
		return f.current
	}

	loaded, err := LoadFileConfig(f.path)
	if err != nil {
		// Keep the previous users: a typo in the file must not lock
		// everyone out, or let everyone in.
		slog.Error("config file invalid, keeping previous", "path", f.path, "err", err)
		f.modified = info.ModTime()
		return f.current
	}
	f.current = loaded
	f.modified = info.ModTime()
	slog.Info("config loaded", "path", f.path, "users", len(loaded.Users), "projects", len(loaded.Projects))
	return f.current
}

// LoadFileConfig reads and validates a task-ui.yml file.
func LoadFileConfig(path string) (*FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg FileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for i, user := range cfg.Users {
		if user.Name == "" {
			return nil, fmt.Errorf("users[%d]: name is required", i)
		}
		if seen[user.Name] {
			return nil, fmt.Errorf("users[%d]: duplicate user %q", i, user.Name)
		}
		seen[user.Name] = true
		if user.Role == "" {
			cfg.Users[i].Role = model.RoleViewer
		} else if !user.Role.Valid() {
			return nil, fmt.Errorf("users[%d]: unknown role %q", i, user.Role)
		}
		if !strings.HasPrefix(user.Password, "$2") {
			return nil, fmt.Errorf("users[%d]: password must be a bcrypt hash, generate one with `task-ui hash`", i)
		}
	}
	return &cfg, nil
}
