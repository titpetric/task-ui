package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/storage"
)

// UserMCPToken is one of a user's MCP tokens as the API shows it: its ID
// to delete it by, its label and its dates. Never the token, which shows
// once, in the POST that made it.
type UserMCPToken struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

// UserMCP is a user's MCP access: the tokens their settings hold, newest
// first, and Enabled for a reader that wants the one question answered.
type UserMCP struct {
	Enabled bool           `json:"enabled"`
	Tokens  []UserMCPToken `json:"tokens"`
}

// newUserMCP returns the access of a user's MCP token rows.
func newUserMCP(rows []model.UserAuthToken) UserMCP {
	out := UserMCP{Enabled: len(rows) > 0, Tokens: make([]UserMCPToken, 0, len(rows))}
	for _, row := range rows {
		out.Tokens = append(out.Tokens, UserMCPToken{
			ID: row.ID, Name: row.Name, CreatedAt: row.CreatedAt, ExpiresAt: row.ExpiresAt, UsedAt: row.UsedAt,
		})
	}
	return out
}

// mcpUser returns the user whose MCP tokens a request manages, when the
// caller may: user/edit manages anyone's, and a user may issue and delete
// their own once they hold one. It answers the request itself when not.
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
		if held, err := h.opts.Tokens.ListKind(r.Context(), user.ID, model.TokenMCP); err == nil && len(held) > 0 {
			return user
		}
	}
	if !h.require(w, r, "user/edit") {
		return nil
	}
	return user
}

// issueMCP makes one more MCP token for a user, under the name the
// request asked for or the first free one. Its 201 holds the token, the
// one place it shows; storage keeps its hash.
//
// It replaces nothing: a user holds as many tokens as they have clients,
// and an old one goes by its own delete or at its expiry.
func (h *Handlers) issueMCP(w http.ResponseWriter, r *http.Request) {
	user := h.mcpUser(w, r)
	if user == nil {
		return
	}
	if !user.Mcp {
		fail(w, http.StatusConflict, user.Name+" is no MCP agent user: an admin turns it on first")
		return
	}
	// A body is optional: the one-click action of a row's menu posts none
	// and takes the first free name, and the form posts the label a reader
	// typed.
	var req struct {
		Name string `json:"name"`
	}
	if r.ContentLength > 0 {
		if err := readJSON(r, &req); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	held, err := h.opts.Tokens.ListKind(r.Context(), user.ID, model.TokenMCP)
	if err != nil {
		failErr(w, err)
		return
	}
	taken := make([]string, 0, len(held))
	for _, row := range held {
		taken = append(taken, row.Name)
	}
	token, hash, err := access.NewToken()
	if err != nil {
		failErr(w, err)
		return
	}
	expires := h.now().Add(access.MCPTokenTTL)
	row, err := h.opts.Tokens.Add(r.Context(), user.ID, model.TokenMCP, access.TokenName(req.Name, taken), hash, &expires)
	if err != nil {
		failErr(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusCreated, map[string]any{"id": row.ID, "name": row.Name, "token": token, "expires_at": row.ExpiresAt})
}

// deleteMCPToken removes one of a user's MCP tokens, by the ID their page
// lists it under. It stops signing in right away.
func (h *Handlers) deleteMCPToken(w http.ResponseWriter, r *http.Request) {
	user := h.mcpUser(w, r)
	if user == nil {
		return
	}
	if err := h.opts.Tokens.DeleteID(r.Context(), user.ID, param(r, "token")); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// disableMCP removes every MCP token of a user; the tokens of `task-ui
// token` stay, because nothing on a page issued them.
func (h *Handlers) disableMCP(w http.ResponseWriter, r *http.Request) {
	user := h.mcpUser(w, r)
	if user == nil {
		return
	}
	if err := h.opts.Tokens.DeleteKind(r.Context(), user.ID, model.TokenMCP); err != nil && !errors.Is(err, storage.ErrNotFound) {
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
