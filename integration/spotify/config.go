package spotify

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// Scopes. Read is what a link asks for by default; Write only when an
// admin enables it on the integration.
var (
	ReadScopes = []string{
		"user-read-playback-state",
		"user-read-currently-playing",
		"playlist-read-private",
		"user-library-read",
	}
	WriteScopes = []string{
		"user-modify-playback-state",
		"playlist-modify-public",
		"playlist-modify-private",
		"user-library-modify",
	}
)

// CallbackPath is where Spotify sends the browser back, under the
// instance's public URL.
const CallbackPath = "/api/integrations/spotify/callback"

// FinishPath is where the callback page sends the browser on, inside the
// session (see Service.Complete).
const FinishPath = "/api/integrations/spotify/finish"

// The reasons a link or a call is not available.
var (
	ErrOff      = errors.New("the Spotify integration is not set up")
	ErrWho      = errors.New("the Spotify integration is not open to you")
	ErrNoLink   = errors.New("your Spotify is not linked: link it from Settings, Integrations")
	ErrRedirect = errors.New("the Spotify redirect URI has to be a public https address, or http on a loopback host; set the URL of the instance (TASK_UI_URL) or the redirect URI")
)

// Config is the integration's settings, read.
type Config struct {
	ClientID     string
	ClientSecret string
	Redirect     string
	// Write adds the write scopes to a link.
	Write bool
	Who   string
}

// Scopes returns the scopes a link asks for.
func (c Config) Scopes() []string {
	scopes := append([]string{}, ReadScopes...)
	if c.Write {
		scopes = append(scopes, WriteScopes...)
	}
	return scopes
}

// Actor is who asks: the user and the role.
type Actor struct {
	Name string
	Role string
}

// Config reads the integration's settings for an actor: ErrOff when it is
// not set up, ErrWho when its Who names neither the user nor the role.
func (s *Service) Config(ctx context.Context, a Actor) (Config, error) {
	values, err := s.Settings(ctx)
	if err != nil {
		return Config{}, err
	}
	c := Config{
		ClientID:     strings.TrimSpace(values["client_id"]),
		ClientSecret: strings.TrimSpace(values["client_secret"]),
		Redirect:     strings.TrimSpace(values["redirect_uri"]),
		Write:        values["scopes"] == "write",
		Who:          strings.TrimSpace(values["who"]),
	}
	if c.ClientID == "" || c.ClientSecret == "" {
		return Config{}, ErrOff
	}
	if c.Who != "" && !listed(c.Who, a.Name) && !listed(c.Who, a.Role) {
		return Config{}, ErrWho
	}
	// The instance owns the address: Spotify sends the browser to this
	// instance and nowhere else, so a stored value is only ever a paste
	// into the wrong box.
	if s.BaseURL != "" {
		c.Redirect = strings.TrimRight(s.BaseURL, "/") + CallbackPath
	}
	if !validRedirect(c.Redirect) {
		return Config{}, ErrRedirect
	}
	return c, nil
}

// validRedirect accepts https, or http on a loopback host, which Spotify
// allows for development.
func validRedirect(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.Fragment != "" || u.User != nil {
		return false
	}
	switch u.Scheme {
	case "https":
		return true
	case "http":
		host := u.Hostname()
		return host == "127.0.0.1" || host == "[::1]" || host == "::1" || host == "localhost"
	}
	return false
}

func listed(list, item string) bool {
	for _, name := range strings.Split(list, ",") {
		name = strings.TrimSpace(name)
		if name == "*" || (name != "" && name == item) {
			return true
		}
	}
	return false
}
