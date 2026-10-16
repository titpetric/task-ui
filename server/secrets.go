package server

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/secret"
)

// maxFormBody bounds a saved env file or secret value.
const maxFormBody = 256 << 10

// secretsPage lists the secrets, values masked.
func (svc *Server) secretsPage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "Secrets")
	data.Secrets = svc.secrets.List()
	data.EnvPath = svc.config.EnvPath
	data.Notice = r.URL.Query().Get("notice")
	data.Error = r.URL.Query().Get("error")
	svc.render(w, r, "secrets.tpl", data)
}

// secretForm renders the add or edit form. Edit names the secret in the
// query; a secret that is not there falls back to the add form.
func (svc *Server) secretForm(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "Add secret")
	data.EnvPath = svc.config.EnvPath
	ref := secret.Secret{Scope: r.URL.Query().Get("scope"), Key: r.URL.Query().Get("key")}
	if ref.Scope != "" && ref.Key != "" {
		if found, ok := svc.secrets.Get(ref); ok {
			data.Title = "Edit secret"
			data.Secret = &found
		}
	}
	if data.Secret == nil {
		data.Secret = &secret.Secret{Scope: secret.ScopeAll}
		if secret.ValidScope(ref.Scope) {
			data.Secret.Scope = ref.Scope
		}
	}
	data.Error = r.URL.Query().Get("error")
	svc.render(w, r, "secret_form.tpl", data)
}

// secretSave stores a secret.
func (svc *Server) secretSave(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBody)
	item := secret.Secret{
		Scope: strings.TrimSpace(r.PostFormValue("scope")),
		Key:   strings.TrimSpace(r.PostFormValue("key")),
		Value: strings.TrimRight(strings.ReplaceAll(r.PostFormValue("value"), "\r\n", "\n"), "\n"),
	}
	if item.Scope != secret.ScopeAll && svc.folders().Find(item.Scope) == nil {
		redirectWith(w, r, "/settings/secrets/new", "", errors.New("unknown folder "+item.Scope))
		return
	}
	if err := svc.secrets.Set(item); err != nil {
		redirectWith(w, r, "/settings/secrets/new?scope="+queryEscape(item.Scope), "", err)
		return
	}
	slog.Info("secret saved", "scope", item.Scope, "key", item.Key, "by", currentUser(r).Name)
	redirectWith(w, r, "/settings/secrets", "Saved "+item.Key+" for "+item.Scope, nil)
}

// secretDelete removes a secret.
func (svc *Server) secretDelete(w http.ResponseWriter, r *http.Request) {
	ref := secret.Secret{Scope: r.PostFormValue("scope"), Key: r.PostFormValue("key")}
	err := svc.secrets.Delete(ref)
	if err == nil {
		slog.Info("secret deleted", "scope", ref.Scope, "key", ref.Key, "by", currentUser(r).Name)
	}
	redirectWith(w, r, "/settings/secrets", "Deleted "+ref.Key+" from "+ref.Scope, err)
}
