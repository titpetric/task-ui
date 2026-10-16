package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

// maxKeyUpload bounds an uploaded private key. A 16k RSA key in PEM is
// around 12KB.
const maxKeyUpload = 64 << 10

// SettingsPage lists the settings sections.
func (svc *Server) SettingsPage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "Settings")
	data.Keys = svc.keys.List()
	data.Secrets = svc.env.Secrets()
	data.Folders = svc.git.ListFolders(svc.config.Root)
	data.KeysDir = svc.keys.Dir
	data.EnvPath = svc.env.Root
	svc.render(w, r, "settings.tpl", data)
}

// KeysPage lists the SSH keys. The forms that change them are pages of
// their own, so the list stays a list.
func (svc *Server) KeysPage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "SSH keys")
	data.Keys = svc.keys.List()
	data.KeysDir = svc.keys.Dir
	data.Notice = r.URL.Query().Get("notice")
	data.Error = r.URL.Query().Get("error")
	svc.render(w, r, "keys.tpl", data)
}

func redirectWith(w http.ResponseWriter, r *http.Request, path, notice string, err error) {
	query := url.Values{}
	if err != nil {
		query.Set("error", err.Error())
	} else if notice != "" {
		query.Set("notice", notice)
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	http.Redirect(w, r, path, http.StatusSeeOther)
}

func keyComment(r *http.Request) string {
	comment := strings.TrimSpace(r.PostFormValue("comment"))
	if comment == "" {
		comment = "task-ui"
	}
	return strings.Join(strings.Fields(comment), "-")
}

// KeysGeneratePage renders the generate form.
func (svc *Server) KeysGeneratePage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "Generate key")
	data.Keys = svc.keys.List()
	svc.render(w, r, "keys_generate.tpl", data)
}

// KeysGenerate creates or rotates a key pair.
func (svc *Server) KeysGenerate(w http.ResponseWriter, r *http.Request) {
	kind := r.PostFormValue("type")
	key, err := svc.keys.Generate(kind, keyComment(r))
	if err == nil {
		slog.Info("ssh key generated", "name", key.Name, "fingerprint", key.Fingerprint, "by", currentUser(r).Name)
	}
	redirectWith(w, r, "/settings/keys", "Generated "+key.Name+" "+key.Fingerprint, err)
}

// KeysUploadPage renders the upload form.
func (svc *Server) KeysUploadPage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "Upload key")
	svc.render(w, r, "keys_upload.tpl", data)
}

// KeysUpload stores an uploaded private key, from the file field or the
// pasted text.
func (svc *Server) KeysUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxKeyUpload)
	if err := r.ParseMultipartForm(maxKeyUpload); err != nil && err != http.ErrNotMultipart {
		redirectWith(w, r, "/settings/keys/upload", "", err)
		return
	}

	data := []byte(r.PostFormValue("private_key"))
	if file, _, err := r.FormFile("key_file"); err == nil {
		defer file.Close()
		if uploaded, err := io.ReadAll(file); err == nil && len(strings.TrimSpace(string(uploaded))) > 0 {
			data = uploaded
		}
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		redirectWith(w, r, "/settings/keys/upload", "", errEmptyKey)
		return
	}

	key, err := svc.keys.Upload(data, keyComment(r))
	if err != nil {
		redirectWith(w, r, "/settings/keys/upload", "", err)
		return
	}
	slog.Info("ssh key uploaded", "name", key.Name, "fingerprint", key.Fingerprint, "by", currentUser(r).Name)
	redirectWith(w, r, "/settings/keys", "Stored "+key.Name+" "+key.Fingerprint, nil)
}

type keyError string

func (e keyError) Error() string { return string(e) }

const errEmptyKey = keyError("paste a private key or choose a key file")

// KeysDelete removes a key pair or backup.
func (svc *Server) KeysDelete(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	err := svc.keys.Delete(name)
	if err == nil {
		slog.Info("ssh key deleted", "name", name, "by", currentUser(r).Name)
	}
	redirectWith(w, r, "/settings/keys", "Deleted "+name, err)
}
