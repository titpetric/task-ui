package server

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
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

// crumb is one breadcrumb; an empty Href is the current page.
type crumb struct {
	Text string
	Href string
}

// pageData is what every template renders from.
type pageData struct {
	Title string
	Path  string

	User          *access.User
	LoginRequired bool

	Folders []*folder.Folder
	Folder  *folder.Folder
	Job     *folder.Job

	Runs    []*run.Run
	History []history.Record
	Recent  []history.Record

	HistoryEnabled bool

	Keys    []sshkey.Key
	KeysDir string

	Secrets  []secret.Secret
	Secret   *secret.Secret
	EnvFiles []envfile.File
	EnvFile  *envfile.File
	EnvPath  string

	Git     *git.Info
	Entries []folder.Entry
	Root    string
	Form    map[string]string

	Back     string
	Username string
	Error    string
	Notice   string

	running map[string]*run.Run
	last    func(folderID, job string) (history.Record, bool)
}

func (svc *Server) pageData(r *http.Request, title string) *pageData {
	data := &pageData{
		Title:          title,
		Path:           r.URL.Path,
		User:           currentUser(r),
		LoginRequired:  svc.loginRequired(),
		HistoryEnabled: svc.config.History,
		running:        map[string]*run.Run{},
		last:           svc.history.Last,
	}
	if data.User != nil {
		for _, f := range svc.folders().List() {
			if data.User.CanSee(f.ID) {
				data.Folders = append(data.Folders, f)
			}
		}
		for _, active := range svc.runs.List() {
			if data.User.CanSee(active.Folder) {
				data.Runs = append(data.Runs, active)
				data.running[active.Key] = active
			}
		}
	}
	return data
}

func (d *pageData) funcs() template.FuncMap {
	return template.FuncMap{
		"canRun":  func(f *folder.Folder, j *folder.Job) bool { return d.User.CanRun(f, j) },
		"running": func(key string) *run.Run { return d.running[key] },
		"last": func(folderID, job string) *history.Record {
			if !d.HistoryEnabled {
				return nil
			}
			if record, ok := d.last(folderID, job); ok {
				return &record
			}
			return nil
		},
		"isPath": func(p string) bool { return d.Path == p },
		"folderRunning": func(id string) int {
			n := 0
			for _, active := range d.Runs {
				if active.Folder == id {
					n++
				}
			}
			return n
		},
		"add":   func(a, b int) int { return a + b },
		"list":  func(items ...any) []any { return items },
		"crumb": func(text, href string) crumb { return crumb{Text: text, Href: href} },
		"whoCanRun": func(f *folder.Folder, j *folder.Job) string {
			who := "admin, runner"
			if len(f.Allow) > 0 {
				who = "admin, " + strings.Join(f.Allow, ", ")
			}
			if len(j.Allow) > 0 {
				who += " (this job: admin, " + strings.Join(j.Allow, ", ") + ")"
			}
			return who
		},
		"ago":      ago,
		"duration": durationText,
		"stamp":    func(t time.Time) string { return t.Local().Format("2006-01-02 15:04:05") },
		"iso":      func(t time.Time) string { return t.UTC().Format(time.RFC3339) },
		"join":     strings.Join,
		"mask": func(value string) string {
			if value == "" {
				return ""
			}
			return "••••••••"
		},
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

func (svc *Server) notFound(w http.ResponseWriter, r *http.Request, data *pageData, message string) {
	data.Title = "Not found"
	data.Error = message
	w.WriteHeader(http.StatusNotFound)
	svc.render(w, r, "notfound.tpl", data)
}

// redirectWith sends the browser to a page with a notice or an error in
// the query, which the page shows as an alert.
func redirectWith(w http.ResponseWriter, r *http.Request, path, notice string, err error) {
	query := ""
	if err != nil {
		query = "?error=" + queryEscape(err.Error())
	} else if notice != "" {
		query = "?notice=" + queryEscape(notice)
	}
	http.Redirect(w, r, path+query, http.StatusSeeOther)
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
