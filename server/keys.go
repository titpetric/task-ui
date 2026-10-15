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

// KeysPage renders the SSH key admin.
func (svc *Server) KeysPage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "SSH keys")
	data.Keys = svc.keys.List()
	data.KeysDir = svc.keys.Dir
	data.Notice = r.URL.Query().Get("notice")
	data.Error = r.URL.Query().Get("error")
	svc.render(w, r, "keys.tpl", data)
}

func keysRedirect(w http.ResponseWriter, r *http.Request, notice string, err error) {
	query := url.Values{}
	if err != nil {
		query.Set("error", err.Error())
	} else {
		query.Set("notice", notice)
	}
	http.Redirect(w, r, "/admin/keys?"+query.Encode(), http.StatusSeeOther)
}

func keyComment(r *http.Request) string {
	comment := strings.TrimSpace(r.PostFormValue("comment"))
	if comment == "" {
		comment = "task-ui"
	}
	return strings.Join(strings.Fields(comment), "-")
}

// KeysGenerate creates or rotates a key pair.
func (svc *Server) KeysGenerate(w http.ResponseWriter, r *http.Request) {
	kind := r.PostFormValue("type")
	key, err := svc.keys.Generate(kind, keyComment(r))
	if err == nil {
		slog.Info("ssh key generated", "name", key.Name, "fingerprint", key.Fingerprint, "by", currentUser(r).Name)
	}
	keysRedirect(w, r, "Generated "+key.Name+" "+key.Fingerprint, err)
}

// KeysUpload stores an uploaded private key, from the file field or the
// pasted text.
func (svc *Server) KeysUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxKeyUpload)
	if err := r.ParseMultipartForm(maxKeyUpload); err != nil && err != http.ErrNotMultipart {
		keysRedirect(w, r, "", err)
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
		keysRedirect(w, r, "", errEmptyKey)
		return
	}

	key, err := svc.keys.Upload(data, keyComment(r))
	if err == nil {
		slog.Info("ssh key uploaded", "name", key.Name, "fingerprint", key.Fingerprint, "by", currentUser(r).Name)
	}
	keysRedirect(w, r, "Stored "+key.Name+" "+key.Fingerprint, err)
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
	keysRedirect(w, r, "Deleted "+name, err)
}
