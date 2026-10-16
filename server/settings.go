package server

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/titpetric/task-ui/server/model"
	"github.com/titpetric/task-ui/server/repository"
)

// maxEnvFile bounds a saved env file or secret value.
const maxEnvFile = 256 << 10

// SecretsPage lists the secrets, values masked.
func (svc *Server) SecretsPage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "Secrets")
	data.Secrets = svc.env.Secrets()
	data.EnvPath = svc.env.Root
	data.Notice = r.URL.Query().Get("notice")
	data.Error = r.URL.Query().Get("error")
	svc.render(w, r, "secrets.tpl", data)
}

// SecretForm renders the add or edit form. Edit names the secret in the
// query; a secret that is not there falls back to the add form.
func (svc *Server) SecretForm(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "Add secret")
	scope := r.URL.Query().Get("scope")
	key := r.URL.Query().Get("key")
	if scope != "" && key != "" {
		if secret, ok := svc.env.Secret(scope, key); ok {
			data.Title = "Edit secret"
			data.Secret = &secret
		}
	}
	if data.Secret == nil {
		data.Secret = &model.Secret{Scope: repository.ScopeAll}
		if scope != "" && repository.ValidScope(scope) {
			data.Secret.Scope = scope
		}
	}
	data.Error = r.URL.Query().Get("error")
	svc.render(w, r, "secret_form.tpl", data)
}

// SecretSave stores a secret.
func (svc *Server) SecretSave(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxEnvFile)
	secret := model.Secret{
		Scope: strings.TrimSpace(r.PostFormValue("scope")),
		Key:   strings.TrimSpace(r.PostFormValue("key")),
		Value: strings.TrimRight(strings.ReplaceAll(r.PostFormValue("value"), "\r\n", "\n"), "\n"),
	}
	if secret.Scope != repository.ScopeAll && !svc.folderExists(secret.Scope) {
		redirectWith(w, r, "/settings/secrets/new", "", keyError("unknown folder "+secret.Scope))
		return
	}
	if err := svc.env.SetSecret(secret); err != nil {
		redirectWith(w, r, "/settings/secrets/new?scope="+secret.Scope, "", err)
		return
	}
	slog.Info("secret saved", "scope", secret.Scope, "key", secret.Key, "by", currentUser(r).Name)
	redirectWith(w, r, "/settings/secrets", "Saved "+secret.Key+" for "+secret.Scope, nil)
}

// SecretDelete removes a secret.
func (svc *Server) SecretDelete(w http.ResponseWriter, r *http.Request) {
	scope := r.PostFormValue("scope")
	key := r.PostFormValue("key")
	err := svc.env.DeleteSecret(scope, key)
	if err == nil {
		slog.Info("secret deleted", "scope", scope, "key", key, "by", currentUser(r).Name)
	}
	redirectWith(w, r, "/settings/secrets", "Deleted "+key+" from "+scope, err)
}

func (svc *Server) folderExists(id string) bool {
	return svc.projects().Find(id) != nil
}

// envFileOf resolves the folder and env file a request names. Only files
// the pipeline declares can be edited.
func (svc *Server) envFileOf(r *http.Request, data *pageData) (*model.Project, string, bool) {
	project := svc.visibleProject(r, data)
	if project == nil {
		return nil, "", false
	}
	path := chi.URLParam(r, "*")
	for _, declared := range project.EnvFiles {
		if declared == path {
			return project, path, true
		}
	}
	return project, path, false
}

// EnvFilePage renders the editor of a folder's env file.
func (svc *Server) EnvFilePage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "")
	project, path, ok := svc.envFileOf(r, data)
	if !ok {
		svc.notFound(w, r, data, "No such env file. Only files the pipeline reads can be edited.")
		return
	}
	file, err := svc.env.File(project, path)
	if err != nil {
		svc.notFound(w, r, data, err.Error())
		return
	}
	data.Title = path + " | " + project.Title
	data.Project = project
	data.EnvFile = &file
	data.Secrets = scopedSecrets(svc.env.Secrets(), project.ID)
	data.Notice = r.URL.Query().Get("notice")
	data.Error = r.URL.Query().Get("error")
	svc.render(w, r, "envfile.tpl", data)
}

func scopedSecrets(all []model.Secret, project string) []model.Secret {
	result := []model.Secret{}
	for _, secret := range all {
		if secret.Scope == repository.ScopeAll || secret.Scope == project {
			result = append(result, secret)
		}
	}
	return result
}

// EnvFileSave stores the content and regenerates the file in the folder.
func (svc *Server) EnvFileSave(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxEnvFile)
	data := svc.pageData(r, "")
	project, path, ok := svc.envFileOf(r, data)
	if !ok {
		http.NotFound(w, r)
		return
	}
	page := "/f/" + project.ID + "/env/" + path
	if err := svc.env.Save(project, path, r.PostFormValue("content")); err != nil {
		redirectWith(w, r, page, "", err)
		return
	}
	slog.Info("env file saved", "project", project.ID, "file", path, "by", currentUser(r).Name)
	redirectWith(w, r, page, "Saved "+path+" and generated it in the folder.", nil)
}
