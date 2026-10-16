package server

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// maxKeyUpload bounds an uploaded private key. A 16k RSA key in PEM is
// around 12KB.
const maxKeyUpload = 64 << 10

var errEmptyKey = errors.New("paste a private key or choose a key file")

// keysPage lists the SSH keys. The forms that change them are pages of
// their own, so the list stays a list.
func (svc *Server) keysPage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "SSH keys")
	data.Keys = svc.keys.List()
	data.KeysDir = svc.keys.Dir
	data.Notice = r.URL.Query().Get("notice")
	data.Error = r.URL.Query().Get("error")
	svc.render(w, r, "keys.tpl", data)
}

func keyComment(r *http.Request) string {
	comment := strings.TrimSpace(r.PostFormValue("comment"))
	if comment == "" {
		comment = "task-ui"
	}
	return strings.Join(strings.Fields(comment), "-")
}

// keysGeneratePage renders the generate form.
func (svc *Server) keysGeneratePage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "Generate key")
	data.KeysDir = svc.keys.Dir
	svc.render(w, r, "keys_generate.tpl", data)
}

// keysGenerate creates or rotates a key pair.
func (svc *Server) keysGenerate(w http.ResponseWriter, r *http.Request) {
	key, err := svc.keys.Generate(r.PostFormValue("type"), keyComment(r))
	if err == nil {
		slog.Info("ssh key generated", "name", key.Name, "fingerprint", key.Fingerprint, "by", currentUser(r).Name)
	}
	redirectWith(w, r, "/settings/keys", "Generated "+key.Name+" "+key.Fingerprint, err)
}

// keysUploadPage renders the upload form.
func (svc *Server) keysUploadPage(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "Upload key")
	data.Error = r.URL.Query().Get("error")
	svc.render(w, r, "keys_upload.tpl", data)
}

// keysUpload stores an uploaded private key, from the file field or the
// pasted text.
func (svc *Server) keysUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxKeyUpload)
	if err := r.ParseMultipartForm(maxKeyUpload); err != nil && !errors.Is(err, http.ErrNotMultipart) {
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

// keysDelete removes a key pair or backup.
func (svc *Server) keysDelete(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	err := svc.keys.Delete(name)
	if err == nil {
		slog.Info("ssh key deleted", "name", name, "by", currentUser(r).Name)
	}
	redirectWith(w, r, "/settings/keys", "Deleted "+name, err)
}
