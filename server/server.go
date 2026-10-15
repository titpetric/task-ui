package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/titpetric/task-ui/server/internal"
	"github.com/titpetric/task-ui/server/model"
	"github.com/titpetric/task-ui/server/repository"
)

// Server is the task-ui HTTP service.
type Server struct {
	config   *Config
	file     *fileConfig
	sessions *sessions
	runs     *internal.Runs
	history  *repository.History
	keys     *repository.Keys

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

	svc := &Server{
		config:    config,
		file:      &fileConfig{path: config.ConfigFile},
		sessions:  newSessions(config.Secret, config.SessionTTL),
		runs:      internal.NewRuns(),
		history:   &repository.History{Root: config.HistoryPath},
		keys:      &repository.Keys{Dir: config.SSHDir},
		assets:    assets,
		templates: templates,
	}
	return svc, nil
}

// projects returns the project repository for the current configuration.
func (svc *Server) projects() *repository.Projects {
	return svc.config.Projects(svc.file.Get())
}

// Handler returns the routes.
func (svc *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(sameOrigin)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok\n")) })
	r.Handle("/static/*", http.StripPrefix("/static/", staticHandler(svc.assets)))
	r.Get("/login", svc.Login)
	r.Post("/login", svc.Login)
	r.Post("/logout", svc.Logout)

	r.Group(func(r chi.Router) {
		r.Use(svc.requireUser)

		r.Get("/", svc.Index)
		r.Get("/p/{project}", svc.ProjectPage)
		r.Get("/p/{project}/{runbook}", svc.RunbookPage)

		r.Get("/api/runs", svc.APIRuns)
		r.Post("/api/runs/{project}/{runbook}/stop", svc.APIStop)
		r.Get("/api/history/{project}/{runbook}/{id}", svc.APIHistory)
		r.Handle("/ws/run/{project}/{runbook}", svc.RunSocket())

		r.Group(func(r chi.Router) {
			r.Use(requireAdmin)
			r.Get("/admin/keys", svc.KeysPage)
			r.Post("/admin/keys/generate", svc.KeysGenerate)
			r.Post("/admin/keys/upload", svc.KeysUpload)
			r.Post("/admin/keys/delete", svc.KeysDelete)
		})
	})
	return r
}

func staticHandler(assets fs.FS) http.Handler {
	files := http.FileServer(http.FS(assets))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=300")
		files.ServeHTTP(w, r)
	})
}

// Start serves until the context is cancelled.
func (svc *Server) Start(ctx context.Context) error {
	if _, port, err := net.SplitHostPort(svc.config.Listen); err != nil {
		return fmt.Errorf("invalid listen address %q: %w", svc.config.Listen, err)
	} else {
		slog.Info("listening", "url", "http://localhost:"+port)
	}

	projects := svc.projects().List()
	slog.Info("projects loaded", "root", svc.config.Root, "count", len(projects))
	seen := map[string]bool{}
	for _, project := range projects {
		if project.Error != "" {
			slog.Warn("project has errors", "project", project.ID, "err", project.Error)
		}
		if len(project.Command) > 0 && !seen[project.Command[0]] {
			seen[project.Command[0]] = true
			if _, err := exec.LookPath(project.Command[0]); err != nil {
				slog.Warn("runner not found in PATH", "command", project.Command[0], "project", project.ID)
			}
		}
	}
	if !svc.loginRequired() {
		slog.Warn("no users configured, visitors get the anonymous role", "role", svc.config.Anonymous)
	}

	server := &http.Server{
		Addr:              svc.config.Listen,
		Handler:           svc.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, run := range svc.runs.List() {
			run.Kill()
		}
		_ = server.Shutdown(shutdown)
	}()
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !currentUser(r).CanAdmin() {
			http.Error(w, "forbidden: admin role required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// pageData is what every template renders from.
type pageData struct {
	Title    string
	Instance string
	Path     string

	User          *model.User
	LoginRequired bool

	Projects []*model.Project
	Project  *model.Project
	Runbook  *model.Runbook

	Runs    []*internal.Run
	History []model.HistoryRecord
	Recent  []model.HistoryRecord

	HistoryEnabled bool

	Keys    []model.SSHKey
	KeysDir string

	Back     string
	Username string
	Error    string
	Notice   string

	running map[string]*internal.Run
	last    func(project, runbook string) (model.HistoryRecord, bool)
}

func (svc *Server) pageData(r *http.Request, title string) *pageData {
	data := &pageData{
		Title:          title,
		Instance:       svc.config.Title,
		Path:           r.URL.Path,
		User:           currentUser(r),
		LoginRequired:  svc.loginRequired(),
		HistoryEnabled: svc.config.History,
		running:        map[string]*internal.Run{},
		last:           svc.history.Last,
	}
	if title := svc.file.Get().Title; title != "" && data.Instance == "" {
		data.Instance = title
	}
	if data.User != nil {
		for _, project := range svc.projects().List() {
			if data.User.CanSee(project.ID) {
				data.Projects = append(data.Projects, project)
			}
		}
		for _, run := range svc.runs.List() {
			if data.User.CanSee(run.Project) {
				data.Runs = append(data.Runs, run)
				data.running[run.Key] = run
			}
		}
	}
	return data
}

func (d *pageData) funcs() template.FuncMap {
	return template.FuncMap{
		"canRun": func(p *model.Project, rb *model.Runbook) bool { return d.User.CanRun(p, rb) },
		"running": func(key string) *internal.Run {
			return d.running[key]
		},
		"last": func(project, runbook string) *model.HistoryRecord {
			if !d.HistoryEnabled {
				return nil
			}
			if record, ok := d.last(project, runbook); ok {
				return &record
			}
			return nil
		},
		"isPath": func(p string) bool { return d.Path == p },
		"projectRunning": func(id string) int {
			n := 0
			for _, run := range d.Runs {
				if run.Project == id {
					n++
				}
			}
			return n
		},
		"add":  func(a, b int) int { return a + b },
		"list": func(items ...any) []any { return items },
		"whoCanRun": func(p *model.Project, rb *model.Runbook) string {
			who := "admin, runner"
			if len(p.Allow) > 0 {
				who = "admin, " + strings.Join(p.Allow, ", ")
			}
			if len(rb.Allow) > 0 {
				who += " (this runbook: admin, " + strings.Join(rb.Allow, ", ") + ")"
			}
			return who
		},
		"ago":      ago,
		"duration": durationText,
		"stamp":    func(t time.Time) string { return t.Local().Format("2006-01-02 15:04:05") },
		"iso":      func(t time.Time) string { return t.UTC().Format(time.RFC3339) },
	}
}

// render executes a page template inside the layout.
func (svc *Server) render(w http.ResponseWriter, r *http.Request, name string, data *pageData) {
	t, err := template.New(name).Funcs(data.funcs()).ParseFS(svc.templates, "_layout.tpl", name)
	if err != nil {
		slog.Error("template parse", "template", name, "err", err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	var out bytes.Buffer
	if err := t.ExecuteTemplate(&out, "layout", data); err != nil {
		slog.Error("template render", "template", name, "err", err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "same-origin")
	_, _ = out.WriteTo(w)
}

func ago(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}

func durationText(d time.Duration) string {
	switch {
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	case d < time.Minute:
		return fmt.Sprintf("%.1fs", d.Seconds())
	case d < time.Hour:
		return fmt.Sprintf("%dm%02ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%02dm", int(d.Hours()), int(d.Minutes())%60)
}

// Index renders the overview: projects, active runs, recent history.
func (svc *Server) Index(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "Overview")
	if svc.config.History {
		for _, record := range svc.history.Recent(50) {
			if data.User.CanSee(record.Project) && len(data.Recent) < 15 {
				data.Recent = append(data.Recent, record)
			}
		}
	}
	svc.render(w, r, "index.tpl", data)
}

// visibleProject finds a project the user may see.
func (svc *Server) visibleProject(r *http.Request, data *pageData) *model.Project {
	id := chi.URLParam(r, "project")
	for _, project := range data.Projects {
		if project.ID == id {
			return project
		}
	}
	return nil
}

// ProjectPage renders a project with its runbooks.
func (svc *Server) ProjectPage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "")
	project := svc.visibleProject(r, data)
	if project == nil {
		svc.notFound(w, r, data, "No such project.")
		return
	}
	data.Title = project.Title
	data.Project = project
	svc.render(w, r, "project.tpl", data)
}

// RunbookPage renders a runbook: its description, run actions and history.
func (svc *Server) RunbookPage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "")
	project := svc.visibleProject(r, data)
	runbook := project.Runbook(chi.URLParam(r, "runbook"))
	if runbook == nil {
		svc.notFound(w, r, data, "No such runbook.")
		return
	}
	data.Title = runbook.Name + " · " + project.Title
	data.Project = project
	data.Runbook = runbook
	if svc.config.History {
		data.History = svc.history.List(project.ID, runbook.Name)
	}
	svc.render(w, r, "runbook.tpl", data)
}

func (svc *Server) notFound(w http.ResponseWriter, r *http.Request, data *pageData, message string) {
	data.Title = "Not found"
	data.Error = message
	w.WriteHeader(http.StatusNotFound)
	svc.render(w, r, "notfound.tpl", data)
}

// APIHistory serves a recording.
func (svc *Server) APIHistory(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "")
	project := svc.visibleProject(r, data)
	runbook := project.Runbook(chi.URLParam(r, "runbook"))
	if runbook == nil || !svc.config.History {
		http.NotFound(w, r)
		return
	}
	file, err := svc.history.CastPath(project.ID, runbook.Name, chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/x-asciicast")
	w.Header().Set("Content-Disposition", "inline; filename="+path.Base(file))
	http.ServeFile(w, r, file)
}
