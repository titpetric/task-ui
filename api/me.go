package api

import (
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/config"
)

// getMe answers what the user of the request may do, a MeView. Anyone
// signed in reads their own.
func (h *Handlers) getMe(w http.ResponseWriter, r *http.Request) {
	user := access.UserFrom(r.Context())
	set := h.permissions(r)
	p, effective := config.Default(), set.Effective()
	view := MeView{Name: user.Name, Role: user.Role, Admin: user.CanAdmin(), Allowed: set.Allows(), Summary: set.Summary(), Folders: []string{}}
	if view.Allowed == nil {
		view.Allowed = []string{}
	}
	for _, section := range p.Sections {
		area := AreaView{Key: section.Key, Title: section.Title, About: config.About(section.Key)}
		for _, rule := range p.Rules {
			if path := section.Key + "/" + rule.Key; effective[path] {
				area.Rules = append(area.Rules, rule.Key)
			}
		}
		area.Allowed = len(area.Rules) > 0
		view.Sections = append(view.Sections, area)
	}
	for _, permission := range p.Permissions {
		view.Permissions = append(view.Permissions, AreaView{Key: permission.Key, Title: permission.Title, About: config.About(permission.Key), Allowed: effective[permission.Key]})
	}
	for _, f := range h.opts.Catalogue.List() {
		if !user.CanSee(f.ID) {
			continue
		}
		view.FolderCount++
		if len(view.Folders) < MaxFolders {
			view.Folders = append(view.Folders, f.ID)
		}
	}
	writeJSON(w, http.StatusOK, view)
}
