package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// UserMCP is a user's MCP access: whether their settings hold a token,
// its creation, its expiry and its last use. The token itself shows
// once, in the POST that makes it.
type UserMCP struct {
	Enabled   bool       `json:"enabled"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

// mcpUser returns the user whose MCP token a request manages, when the
// caller may: user/edit manages anyone's, and a user may rotate or
// disable their own once on. It answers the request itself when not.
func (h *Handlers) mcpUser(w http.ResponseWriter, r *http.Request) *model.User {
	if h.opts.Tokens == nil {
		fail(w, http.StatusMethodNotAllowed, "tokens aren't available on this instance")
		return nil
	}
	user, err := h.opts.Users.Get(r.Context(), param(r, "id"))
	if err != nil {
		if h.require(w, r, "user/edit") {
			failErr(w, err)
		}
		return nil
	}
	if caller := access.UserFrom(r.Context()); caller != nil && caller.ID == user.ID {
		if _, err := h.opts.Tokens.Get(r.Context(), user.ID, access.MCPToken); err == nil {
			return user
		}
	}
	if !h.require(w, r, "user/edit") {
		return nil
	}
	return user
}

// enableMCP makes a user's MCP token, in place of the one they had,
// which stops signing in at once. Its 201 holds the token, the one
// place it shows; storage keeps its hash.
func (h *Handlers) enableMCP(w http.ResponseWriter, r *http.Request) {
	user := h.mcpUser(w, r)
	if user == nil {
		return
	}
	token, hash, err := access.NewToken()
	if err != nil {
		failErr(w, err)
		return
	}
	row, err := h.opts.Tokens.SetUntil(r.Context(), user.ID, access.MCPToken, hash, h.now().Add(access.MCPTokenTTL))
	if err != nil {
		failErr(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusCreated, map[string]any{"token": token, "expires_at": row.ExpiresAt})
}

// disableMCP removes a user's MCP token.
func (h *Handlers) disableMCP(w http.ResponseWriter, r *http.Request) {
	user := h.mcpUser(w, r)
	if user == nil {
		return
	}
	if err := h.opts.Tokens.Delete(r.Context(), user.ID, access.MCPToken); err != nil && !errors.Is(err, storage.ErrNotFound) {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// now is the clock of the handlers.
func (h *Handlers) now() time.Time {
	if h.opts.Now != nil {
		return h.opts.Now()
	}
	return time.Now()
}
