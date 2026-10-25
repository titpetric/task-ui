package forgejo

import (
	"slices"
)

// Hook is a webhook of a repository, as the API lists and takes it. The
// listing leaves the secret out of Config; a create or an update sends
// it.
type Hook struct {
	ID     int64             `json:"id,omitempty"`
	Type   string            `json:"type,omitempty"`
	Config map[string]string `json:"config"`
	Events []string          `json:"events"`
	Active bool              `json:"active"`
}

// PushHook returns the hook sync installs: a push hook posting JSON to url, signed with secret.
func PushHook(url, secret string) Hook {
	return Hook{
		Type:   "forgejo",
		Config: map[string]string{"url": url, "content_type": "json", "secret": secret},
		Events: []string{"push"},
		Active: true,
	}
}

// Current reports whether the hook is the push hook sync installs: it posts to url, is active and hears pushes,
// and its secret, when the listing shows one, is secret.
func (h Hook) Current(url, secret string) bool {
	if h.Config["url"] != url || !h.Active || !slices.Contains(h.Events, "push") {
		return false
	}
	shown, ok := h.Config["secret"]
	return !ok || shown == secret
}
