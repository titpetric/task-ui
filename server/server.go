// Package server is the presentation: the HTTP routes, the pages and the
// websocket that drives a run. It composes the domain packages and holds
// no rule of its own beyond what a page needs.
package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime/debug"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/history"
	"github.com/titpetric/task-ui/run"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/sshkey"
)

// Server is the task-ui HTTP service.
type Server struct {
	config   *Config
	file     *configFile
	sessions *access.Sessions
	runs     *run.Registry
	history  *history.Store
	keys     *sshkey.Store
	secrets  *secret.Store
	env      *envfile.Store
	git      *git.Git

	assets    fs.FS
	templates fs.FS
}

// New builds the server from the environment and the command line.
func New(files fs.FS, args []string) (*Server, error) {
	config, err := NewConfig(args)
	if err != nil {
		return nil, err
	}

	if config.PrintBuildInfo {
		buildInfo, ok := debug.ReadBuildInfo()
		if !ok {
			return nil, errors.New("error when reading build info")
		}
		fmt.Println(buildInfo)
		os.Exit(0)
	}

	return NewWithConfig(files, config)
}

// NewWithConfig builds the server from a configuration.
func NewWithConfig(files fs.FS, config *Config) (*Server, error) {
	assets, err := fs.Sub(files, "public_html/static")
	if err != nil {
		return nil, err
	}
	templates, err := fs.Sub(files, "templates")
	if err != nil {
		return nil, err
	}

	secrets := &secret.Store{Root: config.SecretsPath()}
	svc := &Server{
		config:    config,
		file:      &configFile{path: config.ConfigFile},
		sessions:  access.NewSessions(config.Secret, config.SessionTTL),
		runs:      run.NewRegistry(),
		history:   &history.Store{Root: config.HistoryPath},
		keys:      &sshkey.Store{Dir: config.SSHDir},
		secrets:   secrets,
		env:       &envfile.Store{Root: config.FilesPath(), Secrets: secrets},
		git:       &git.Git{SSHDir: config.SSHDir},
		assets:    assets,
		templates: templates,
	}
	return svc, nil
}

// folders returns the catalogue for the current configuration.
func (svc *Server) folders() *folder.Discovery {
	return svc.config.discovery(svc.file.Get())
}

// Handler returns the routes.
func (svc *Server) Handler() http.Handler {
	return svc.routes()
}

// Start serves until the context is cancelled.
func (svc *Server) Start(ctx context.Context) error {
	_, port, err := net.SplitHostPort(svc.config.Listen)
	if err != nil {
		return fmt.Errorf("invalid listen address %q: %w", svc.config.Listen, err)
	}
	slog.Info("listening", "url", "http://localhost:"+port)
	svc.checkEnvironment()

	listener, err := net.Listen("tcp", svc.config.Listen)
	if err != nil {
		return err
	}
	server := &http.Server{
		Handler:           svc.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, active := range svc.runs.List() {
			active.Kill()
		}
		_ = server.Shutdown(shutdown)
	}()
	if err := server.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// checkEnvironment says at startup what a deployment would otherwise find
// out on the first click: folders that do not parse, runners missing from
// PATH, no users, the auto login.
func (svc *Server) checkEnvironment() {
	folders := svc.folders().List()
	slog.Info("folders loaded", "root", svc.config.Root, "count", len(folders))
	seen := map[string]bool{}
	for _, f := range folders {
		if f.Error != "" {
			slog.Warn("folder has errors", "folder", f.ID, "err", f.Error)
		}
		for _, source := range f.Sources {
			if len(source.Command) == 0 || seen[source.Command[0]] {
				continue
			}
			seen[source.Command[0]] = true
			if _, err := exec.LookPath(source.Command[0]); err != nil {
				slog.Warn("runner not found in PATH", "command", source.Command[0], "folder", f.ID, "source", source.File)
			}
		}
	}
	if svc.config.AutoLogin != "" {
		slog.Warn("TASK_UI_AUTOLOGIN is set, every request is signed in without a password", "user", svc.config.AutoLogin)
	}
	if !svc.hasUsers() {
		slog.Error("no users configured, nobody can sign in: add users to task-ui.yml with passwords from `task-ui hash`", "config", svc.config.ConfigFile)
	}
	if _, err := exec.LookPath("git"); err != nil {
		slog.Warn("git not found in PATH, projects cannot be cloned or pulled")
	}
}
