package server

import (
	"errors"
	"log/slog"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
)

// FileConfig is the shape of task-ui.yml: the users, and the folders the
// configuration declares or narrows, under the projects key the file has
// always used.
type FileConfig struct {
	Title    string          `yaml:"title"`
	Users    access.Users    `yaml:"users"`
	Projects []folder.Config `yaml:"projects"`
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
	if err := cfg.Users.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// configFile loads task-ui.yml and reloads it when it changes, so users can
// be added or removed without a restart.
type configFile struct {
	path string

	mu       sync.Mutex
	modified time.Time
	current  *FileConfig
}

// Get returns the current configuration, re-read when the file changed.
// A file that stopped parsing keeps the previous configuration: a typo
// must not lock everyone out, or let everyone in.
func (f *configFile) Get() *FileConfig {
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
		slog.Error("config file invalid, keeping previous", "path", f.path, "err", err)
		f.modified = info.ModTime()
		return f.current
	}
	f.current = loaded
	f.modified = info.ModTime()
	slog.Info("config loaded", "path", f.path, "users", len(loaded.Users), "folders", len(loaded.Projects))
	return f.current
}
