package api

import (
	"bytes"
	"io"
	"net/http"
	"slices"
	"strings"

	"filippo.io/age"
	"filippo.io/age/armor"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
)

// KeyReadView is a private key handed to an agent: the key's name,
// algorithm and fingerprint, and its private half armored-encrypted to
// the age recipient the request named. Never the key in clear.
type KeyReadView struct {
	Name        string `json:"name"`
	Algo        string `json:"algo"`
	Bits        int    `json:"bits,omitempty"`
	Fingerprint string `json:"fingerprint"`
	Ciphertext  string `json:"ciphertext"`
	Note        string `json:"note"`
}

// keyReadNote is how an agent uses what it got.
const keyReadNote = "decrypt with `age -d -i <identity>`; use it from /tmp for the length of the task and delete it"

// mayRead reports whether a user may read an SSH key's private half: an
// admin any, anyone else one linked to them or to their role.
func mayRead(user *model.User, key *keystore.Key) bool {
	if key.Encrypts() {
		return false
	}
	return user.CanAdmin() || slices.Contains(key.Users, user.Name) || slices.Contains(key.Roles, user.EffectiveRole())
}

// readableKeys returns the SSH keys the user of a request may read, by
// name: none without MCP agent user or key/read.
func (h *Handlers) readableKeys(r *http.Request) []string {
	user := access.UserFrom(r.Context())
	names := []string{}
	if h.opts.Keys == nil || !user.Mcp || !h.permissions(r).Allowed("key/read") {
		return names
	}
	keys, _ := h.opts.Keys.List(r.Context())
	for i := range keys {
		if mayRead(user, &keys[i]) {
			names = append(names, keys[i].Name)
		}
	}
	return names
}

// readKey answers an SSH key's private half, encrypted to the age
// recipient of ?recipient=, for an agent that uses an instance key for
// the length of a task. It takes key/read and MCP agent user, and the
// key linked to the user or their role, any SSH key for an admin; an
// encryption key never leaves. Each read goes to the key read log.
func (h *Handlers) readKey(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "key/read") || h.keysOff(w) {
		return
	}
	user := access.UserFrom(r.Context())
	if !user.Mcp {
		fail(w, http.StatusForbidden, "reading a private key takes MCP agent user")
		return
	}
	key, err := h.opts.Keys.Get(r.Context(), param(r, "name"))
	if err != nil {
		failErr(w, err)
		return
	}
	switch {
	case key.Encrypts():
		fail(w, http.StatusForbidden, "an encryption key never leaves task-ui")
		return
	case !mayRead(user, key):
		fail(w, http.StatusForbidden, "the key is linked neither to you nor to your role")
		return
	}
	recipient, err := age.ParseX25519Recipient(strings.TrimSpace(r.URL.Query().Get("recipient")))
	if err != nil {
		fail(w, http.StatusBadRequest, "recipient: an age X25519 recipient, age1..., made for this request with age-keygen, is needed")
		return
	}
	private, err := h.opts.Keys.Private(r.Context(), key.Name)
	if err != nil {
		failErr(w, err)
		return
	}
	sealed, err := seal(recipient, private)
	if err != nil {
		failErr(w, err)
		return
	}
	if h.opts.KeyReads != nil {
		read := &model.KeyRead{KeyName: key.Name, UserName: user.Name, TokenID: access.TokenFrom(r.Context()), Via: access.Via(r.Context())}
		if err := h.opts.KeyReads.Record(r.Context(), read); err != nil {
			failErr(w, err)
			return
		}
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, KeyReadView{Name: key.Name, Algo: key.Algo, Bits: key.Bits, Fingerprint: key.Fingerprint, Ciphertext: sealed, Note: keyReadNote})
}

// seal encrypts data to an age recipient, armored.
func seal(recipient age.Recipient, data []byte) (string, error) {
	var buf bytes.Buffer
	arm := armor.NewWriter(&buf)
	enc, err := age.Encrypt(arm, recipient)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(enc, bytes.NewReader(data)); err != nil {
		return "", err
	}
	if err := enc.Close(); err != nil {
		return "", err
	}
	if err := arm.Close(); err != nil {
		return "", err
	}
	return buf.String(), nil
}
