package server

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/folder"
)

// projectsPage lists the directories under the root with their sources.
func (svc *Server) projectsPage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "Projects")
	data.Entries = folder.Entries(svc.config.Root, svc.git)
	data.Root = mustAbs(svc.config.Root)
	data.Notice = r.URL.Query().Get("notice")
	data.Error = r.URL.Query().Get("error")
	svc.render(w, r, "projects.tpl", data)
}

// projectNewPage renders the clone form, with what was typed when the
// last attempt failed.
func (svc *Server) projectNewPage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "Add project")
	query := r.URL.Query()
	data.Form = map[string]string{
		"url":    query.Get("url"),
		"id":     query.Get("id"),
		"branch": query.Get("branch"),
	}
	data.Error = query.Get("error")
	data.Notice = query.Get("output")
	svc.render(w, r, "project_new.tpl", data)
}

func queryEscape(s string) string {
	return url.QueryEscape(s)
}

// folderIDFromURL derives a folder name from a repository URL.
func folderIDFromURL(repo string) string {
	base := strings.TrimSuffix(strings.TrimRight(repo, "/"), ".git")
	if i := strings.LastIndexAny(base, "/:"); i >= 0 {
		base = base[i+1:]
	}
	return folder.SafeID(base)
}

// projectClone clones a repository into the root. The clone is the test:
// a URL that cannot be cloned is reported with git's output, and leaves
// nothing behind.
func (svc *Server) projectClone(w http.ResponseWriter, r *http.Request) {
	repo := strings.TrimSpace(r.PostFormValue("url"))
	id := strings.TrimSpace(r.PostFormValue("id"))
	branch := strings.TrimSpace(r.PostFormValue("branch"))
	if id == "" {
		id = folderIDFromURL(repo)
	}

	back := "/settings/projects/new?url=" + queryEscape(repo) + "&id=" + queryEscape(id) + "&branch=" + queryEscape(branch)
	fail := func(err error, output string) {
		http.Redirect(w, r, back+"&error="+queryEscape(err.Error())+"&output="+queryEscape(output), http.StatusSeeOther)
	}
	if !svc.git.ValidURL(repo) {
		fail(errors.New("enter a git URL: https://, ssh://, git@host:path or file://"), "")
		return
	}
	if !folder.ValidID(id) {
		fail(errors.New("the folder name may hold letters, digits, dots, dashes and underscores"), "")
		return
	}

	dir := filepath.Join(mustAbs(svc.config.Root), id)
	output, err := svc.git.Clone(r.Context(), repo, dir, branch)
	if err != nil {
		slog.Warn("project clone failed", "url", repo, "folder", id, "by", currentUser(r).Name, "err", err)
		fail(err, output)
		return
	}
	slog.Info("project cloned", "url", repo, "folder", id, "by", currentUser(r).Name)

	notice := "Cloned " + repo + " into " + id + "."
	if !detector.Any(dir) {
		notice += " Nothing to run was found in it: no pipeline file, Makefile, package scripts or scripts folder."
	}
	redirectWith(w, r, "/settings/projects", notice, nil)
}

// projectDelete removes a cloned folder.
func (svc *Server) projectDelete(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PostFormValue("id"))
	if !folder.ValidID(id) {
		redirectWith(w, r, "/settings/projects", "", errors.New("invalid folder name"))
		return
	}
	dir := filepath.Join(mustAbs(svc.config.Root), id)
	if err := svc.git.Remove(dir); err != nil {
		redirectWith(w, r, "/settings/projects", "", err)
		return
	}
	slog.Info("project removed", "folder", id, "by", currentUser(r).Name)
	redirectWith(w, r, "/settings/projects", "Removed "+id+".", nil)
}
