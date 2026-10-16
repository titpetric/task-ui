package server

import (
	"net/http"

	"github.com/titpetric/task-ui/folder"
)

// settingsPage lists the settings sections.
func (svc *Server) settingsPage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "Settings")
	data.Keys = svc.keys.List()
	data.Secrets = svc.secrets.List()
	data.Entries = folder.Entries(svc.config.Root, svc.git)
	data.KeysDir = svc.keys.Dir
	data.EnvPath = svc.config.EnvPath
	svc.render(w, r, "settings.tpl", data)
}
