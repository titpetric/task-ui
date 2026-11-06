package spotify

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"
)

// Service is Spotify for the users of task-ui: it links accounts, keeps
// their tokens fresh and acts on an account as its user.
type Service struct {
	// Settings reads the integration's settings, decrypted.
	Settings func(ctx context.Context) (map[string]string, error)
	Tokens   Tokens
	Client   *Client
	OAuth    *OAuth
	States   *States
	// Limit caps the write calls of a user; nil is 30 a minute.
	Limit *Limiter
	// BaseURL is the instance's public URL, TASK_UI_URL, which the default
	// redirect URI derives from.
	BaseURL string
	Now     func() time.Time

	locks sync.Map
	limit sync.Once
}

func (s *Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s *Service) client() *Client {
	if s.Client == nil {
		s.Client = &Client{}
	}
	return s.Client
}

func (s *Service) oauth() *OAuth {
	if s.OAuth == nil {
		s.OAuth = &OAuth{}
	}
	return s.OAuth
}

func (s *Service) userLock(user string) *sync.Mutex {
	m, _ := s.locks.LoadOrStore(user, &sync.Mutex{})
	return m.(*sync.Mutex)
}

// Status is a user's link as a page shows it. It holds no token.
type Status struct {
	// Available is whether the integration is set up and open to the user.
	Available bool `json:"available"`
	// Reason says why not, when it is not.
	Reason string `json:"reason,omitempty"`
	Linked bool   `json:"linked"`
	Broken bool   `json:"broken"`
	// Write is whether the link was granted the write scopes.
	Write   bool   `json:"write"`
	Account string `json:"account,omitempty"`
}

// Status reads the link of the actor.
func (s *Service) Status(ctx context.Context, a Actor) (Status, error) {
	var st Status
	if _, err := s.Config(ctx, a); err != nil {
		if errors.Is(err, ErrOff) || errors.Is(err, ErrWho) || errors.Is(err, ErrRedirect) {
			st.Reason = err.Error()
			return st, nil
		}
		return st, err
	}
	st.Available = true
	link, err := s.Tokens.Load(ctx, a.Name)
	if err != nil || link == nil {
		return st, err
	}
	st.Linked, st.Broken, st.Account = true, link.Broken, link.Name
	st.Write = strings.Contains(link.Scope, "user-modify-playback-state")
	return st, nil
}

// Begin starts a link for the actor in a session and returns the address
// of Spotify's consent page.
func (s *Service) Begin(ctx context.Context, a Actor, session string) (string, error) {
	cfg, err := s.Config(ctx, a)
	if err != nil {
		return "", err
	}
	if s.States == nil {
		s.States = &States{Now: s.Now}
	}
	state, verifier := s.States.New(a.Name, session)
	return s.oauth().AuthorizeURL(cfg.ClientID, cfg.Redirect, cfg.Scopes(), state, verifier), nil
}

// Complete finishes a link: it checks the state against the actor and the
// session, trades the code for tokens and keeps them.
func (s *Service) Complete(ctx context.Context, a Actor, session, state, code string) error {
	cfg, err := s.Config(ctx, a)
	if err != nil {
		return err
	}
	if s.States == nil {
		return ErrState
	}
	verifier, err := s.States.Take(state, a.Name, session)
	if err != nil {
		return err
	}
	if code == "" {
		return errors.New("Spotify sent no code")
	}
	token, err := s.oauth().Exchange(ctx, cfg.ClientID, cfg.ClientSecret, cfg.Redirect, code, verifier)
	if err != nil {
		return s.redact(err, token, cfg)
	}
	link := Link{Token: token}
	if me, err := s.client().Me(ctx, token.Access); err == nil {
		link.Name = me.DisplayName
		if link.Name == "" {
			link.Name = me.ID
		}
	}
	if err := s.Tokens.Save(ctx, a.Name, link); err != nil {
		return err
	}
	slog.Info("spotify linked", "user", a.Name)
	return nil
}

// Unlink deletes a user's stored tokens. actor is who asks; only the user
// or an admin may.
func (s *Service) Unlink(ctx context.Context, actor, target string, admin bool) error {
	if target == "" {
		target = actor
	}
	if target != actor && !admin {
		return errors.New("you can only disconnect your own Spotify")
	}
	if err := s.Tokens.Delete(ctx, target); err != nil {
		return err
	}
	slog.Info("spotify unlinked", "actor", actor, "target", target)
	return nil
}

// redact keeps a credential out of an error: the tokens and the app's
// secret are replaced wherever the message holds them.
func (s *Service) redact(err error, token Token, cfg Config) error {
	msg := err.Error()
	for _, secret := range []string{token.Access, token.Refresh, cfg.ClientSecret} {
		if len(secret) >= 8 {
			msg = strings.ReplaceAll(msg, secret, "[redacted]")
		}
	}
	if msg == err.Error() {
		return err
	}
	return redacted{msg: msg, err: err}
}

type redacted struct {
	msg string
	err error
}

func (r redacted) Error() string { return r.msg }
func (r redacted) Unwrap() error { return r.err }

// fresh returns a link whose access token is good for another minute,
// refreshing it when it is not or when force asks.
func (s *Service) fresh(ctx context.Context, a Actor, cfg Config, force bool) (*Link, error) {
	lock := s.userLock(a.Name)
	lock.Lock()
	defer lock.Unlock()
	link, err := s.Tokens.Load(ctx, a.Name)
	if err != nil {
		return nil, err
	}
	switch {
	case link == nil:
		return nil, ErrNoLink
	case link.Broken:
		return nil, ErrBroken
	}
	if !force && link.Access != "" && s.now().Add(time.Minute).Before(link.Expires) {
		return link, nil
	}
	token, err := s.oauth().Refresh(ctx, cfg.ClientID, cfg.ClientSecret, link.Refresh)
	if errors.Is(err, ErrBroken) {
		link.Broken = true
		if serr := s.Tokens.Save(ctx, a.Name, *link); serr != nil {
			return nil, serr
		}
		slog.Warn("spotify link broken", "user", a.Name)
		return nil, ErrBroken
	}
	if err != nil {
		return nil, s.redact(err, Token{Access: link.Access, Refresh: link.Refresh}, cfg)
	}
	if token.Scope == "" {
		token.Scope = link.Scope
	}
	link.Token = token
	if err := s.Tokens.Save(ctx, a.Name, *link); err != nil {
		return nil, err
	}
	return link, nil
}

// call runs fn with the actor's access token. A 401 refreshes once and
// tries again. A write call needs the write scopes enabled and the rate.
func (s *Service) call(ctx context.Context, a Actor, write bool, fn func(token string) error) error {
	cfg, err := s.Config(ctx, a)
	if err != nil {
		return err
	}
	if write {
		if !cfg.Write {
			return ErrScope
		}
		s.limit.Do(func() {
			if s.Limit == nil {
				s.Limit = &Limiter{Now: s.Now}
			}
		})
		if !s.Limit.Allow(a.Name) {
			return ErrLimited
		}
	}
	link, err := s.fresh(ctx, a, cfg, false)
	if err != nil {
		return err
	}
	err = fn(link.Access)
	if errors.Is(err, ErrUnauthorized) {
		if link, err = s.fresh(ctx, a, cfg, true); err != nil {
			return err
		}
		err = fn(link.Access)
	}
	if err != nil {
		return s.redact(err, link.Token, cfg)
	}
	return nil
}

// PlayingView is what is playing, as the API answers.
type PlayingView struct {
	Playing  bool   `json:"playing"`
	Track    string `json:"track,omitempty"`
	TrackID  string `json:"track_id,omitempty"`
	URI      string `json:"uri,omitempty"`
	Artists  string `json:"artists,omitempty"`
	Album    string `json:"album,omitempty"`
	Device   string `json:"device,omitempty"`
	Kind     string `json:"device_type,omitempty"`
	Volume   int    `json:"volume,omitempty"`
	Progress int    `json:"progress_ms,omitempty"`
	// Active is false when no device is playing or paused.
	Active bool `json:"active"`
}

func playingView(p *Playback) PlayingView {
	if p == nil {
		return PlayingView{}
	}
	v := PlayingView{Active: true, Playing: p.IsPlaying, Device: p.Device.Name, Kind: p.Device.Type, Volume: p.Device.VolumePercent, Progress: p.Progress}
	if p.Item != nil {
		v.Track, v.TrackID, v.URI, v.Album = p.Item.Name, p.Item.ID, p.Item.URI, p.Item.Album.Name
		v.Artists = p.Item.ArtistNames()
	}
	return v
}

// NowPlaying reads the current track, artist and device.
func (s *Service) NowPlaying(ctx context.Context, a Actor) (PlayingView, error) {
	var view PlayingView
	err := s.call(ctx, a, false, func(token string) error {
		p, err := s.client().Playback(ctx, token)
		view = playingView(p)
		return err
	})
	return view, err
}

// Search finds tracks.
func (s *Service) Search(ctx context.Context, a Actor, query string, limit int) ([]Track, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("search needs a query")
	}
	var tracks []Track
	err := s.call(ctx, a, false, func(token string) (err error) {
		tracks, err = s.client().Search(ctx, token, query, limit)
		return err
	})
	return tracks, err
}

// Playlists lists the user's playlists.
func (s *Service) Playlists(ctx context.Context, a Actor) ([]Playlist, error) {
	var lists []Playlist
	err := s.call(ctx, a, false, func(token string) (err error) {
		lists, err = s.client().Playlists(ctx, token)
		return err
	})
	return lists, err
}

// PlaylistContents reads a playlist picked by name or ID and its tracks.
func (s *Service) PlaylistContents(ctx context.Context, a Actor, name string) (Playlist, []Track, error) {
	var (
		list   Playlist
		tracks []Track
	)
	err := s.call(ctx, a, false, func(token string) error {
		lists, err := s.client().Playlists(ctx, token)
		if err != nil {
			return err
		}
		if list, err = pick(lists, name); err != nil {
			return err
		}
		tracks, err = s.client().PlaylistItems(ctx, token, list.ID)
		return err
	})
	return list, tracks, err
}

// AmbiguousError reports a playlist name that fits several playlists; the
// caller asks the user which.
type AmbiguousError struct {
	Name       string
	Candidates []string
}

func (e *AmbiguousError) Error() string {
	return fmt.Sprintf("%q fits several playlists, ask which: %s", e.Name, strings.Join(e.Candidates, "; "))
}

// pick finds a playlist by ID, or by name: an exact name first, ignoring
// case, else the only playlist whose name contains it.
func pick(lists []Playlist, name string) (Playlist, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Playlist{}, errors.New("name the playlist")
	}
	for _, l := range lists {
		if l.ID == name || l.URI == name {
			return l, nil
		}
	}
	var exact, partial []Playlist
	for _, l := range lists {
		switch lower := strings.ToLower(l.Name); {
		case lower == strings.ToLower(name):
			exact = append(exact, l)
		case strings.Contains(lower, strings.ToLower(name)):
			partial = append(partial, l)
		}
	}
	matches := exact
	if len(matches) == 0 {
		matches = partial
	}
	switch len(matches) {
	case 0:
		return Playlist{}, fmt.Errorf("no playlist named %q", name)
	case 1:
		return matches[0], nil
	}
	e := &AmbiguousError{Name: name}
	for _, l := range matches {
		e.Candidates = append(e.Candidates, fmt.Sprintf("%s (by %s, id %s)", l.Name, l.Owner.DisplayName, l.ID))
	}
	return Playlist{}, e
}

var (
	errTrack = errors.New("a track is a Spotify ID, spotify:track: URI or open.spotify.com link")
	// trackRef matches a bare ID, a spotify:track: URI or a track link,
	// with an optional query or fragment.
	trackRef = regexp.MustCompile(`^(?:spotify:track:|https?://open\.spotify\.com/(?:[^/]+/)?track/)?([0-9A-Za-z]{22})/?(?:[?#].*)?$`)
)

// trackID reads a track ID from an ID, a spotify:track: URI or an
// open.spotify.com link.
func trackID(ref string) (string, error) {
	m := trackRef.FindStringSubmatch(strings.TrimSpace(ref))
	if m == nil {
		return "", errTrack
	}
	return m[1], nil
}

// target returns the track a change is about: the one given, else the one playing.
func (s *Service) target(ctx context.Context, token, ref string) (id, label string, err error) {
	if strings.TrimSpace(ref) != "" {
		id, err = trackID(ref)
		return id, id, err
	}
	p, err := s.client().Playback(ctx, token)
	if err != nil {
		return "", "", err
	}
	if p == nil || p.Item == nil {
		return "", "", errors.New("nothing is playing: name a track")
	}
	return p.Item.ID, p.Item.Label(), nil
}

// ControlRequest is a playback command.
type ControlRequest struct {
	// Action is play, pause, next, previous or volume.
	Action string `json:"action"`
	// Volume is 0 to 100, for volume.
	Volume *int `json:"volume,omitempty"`
	// Query finds a track to play.
	Query string `json:"query,omitempty"`
	// URI is a track, album, artist or playlist to play.
	URI string `json:"uri,omitempty"`
}

// Control runs a playback command and says what happened. It needs the
// write scopes, Premium and an active device, and says so when one is missing.
func (s *Service) Control(ctx context.Context, a Actor, req ControlRequest) (string, error) {
	var msg string
	err := s.call(ctx, a, true, func(token string) (err error) {
		msg, err = s.control(ctx, token, req)
		return err
	})
	return msg, err
}

func (s *Service) control(ctx context.Context, token string, req ControlRequest) (string, error) {
	c := s.client()
	switch req.Action {
	case "pause":
		return "Paused.", c.Pause(ctx, token)
	case "next":
		return "Skipped to the next track.", c.Next(ctx, token)
	case "previous":
		return "Went back a track.", c.Previous(ctx, token)
	case "volume":
		if req.Volume == nil {
			return "", errors.New("volume needs a level, 0 to 100")
		}
		return fmt.Sprintf("Volume set to %d.", *req.Volume), c.Volume(ctx, token, *req.Volume)
	case "play":
		return s.play(ctx, token, req)
	}
	return "", fmt.Errorf("unknown action %q: play, pause, next, previous or volume", req.Action)
}

// play starts the track a query finds, the URI given, or resumes.
func (s *Service) play(ctx context.Context, token string, req ControlRequest) (string, error) {
	c := s.client()
	switch {
	case req.Query != "":
		tracks, err := c.Search(ctx, token, req.Query, 1)
		if err != nil {
			return "", err
		}
		if len(tracks) == 0 {
			return "", fmt.Errorf("nothing found for %q", req.Query)
		}
		return "Playing " + tracks[0].Label() + ".", c.Play(ctx, token, PlayOptions{URIs: []string{tracks[0].URI}})
	case strings.HasPrefix(req.URI, "spotify:track:"):
		return "Playing " + req.URI + ".", c.Play(ctx, token, PlayOptions{URIs: []string{req.URI}})
	case req.URI != "":
		return "Playing " + req.URI + ".", c.Play(ctx, token, PlayOptions{ContextURI: req.URI})
	}
	return "Resumed.", c.Play(ctx, token, PlayOptions{})
}

// LikeRequest likes or unlikes a track; none given is the one playing.
type LikeRequest struct {
	Track  string `json:"track,omitempty"`
	Unlike bool   `json:"unlike,omitempty"`
}

// ChangeView says what a write did.
type ChangeView struct {
	// Changed is false when the state already was the one asked for.
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// Like saves the track to the Liked Songs, or removes it. It is
// idempotent: a track already in the state asked for changes nothing and
// the answer says so.
func (s *Service) Like(ctx context.Context, a Actor, req LikeRequest) (ChangeView, error) {
	var out ChangeView
	err := s.call(ctx, a, true, func(token string) error {
		c := s.client()
		id, label, err := s.target(ctx, token, req.Track)
		if err != nil {
			return err
		}
		saved, err := c.Saved(ctx, token, []string{id})
		if err != nil {
			return err
		}
		out, err = change(len(saved) == 1 && saved[0], !req.Unlike, label, "your Liked Songs",
			func() error { return c.Save(ctx, token, []string{id}) },
			func() error { return c.Unsave(ctx, token, []string{id}) })
		return err
	})
	return out, err
}

// change moves a track into or out of a collection, or reports that it
// already is where the caller wants it. have is the current state, add the
// wanted one.
func change(have, add bool, label, where string, put, drop func() error) (ChangeView, error) {
	switch {
	case add && have:
		return ChangeView{Message: label + " is already in " + where + "; nothing changed."}, nil
	case !add && !have:
		return ChangeView{Message: label + " is not in " + where + "; nothing changed."}, nil
	case add:
		return ChangeView{Changed: true, Message: "Added " + label + " to " + where + "."}, put()
	}
	return ChangeView{Changed: true, Message: "Removed " + label + " from " + where + "."}, drop()
}

// PlaylistRequest adds or removes a track of a playlist picked by name or ID.
type PlaylistRequest struct {
	Playlist string `json:"playlist"`
	Track    string `json:"track,omitempty"`
	Remove   bool   `json:"remove,omitempty"`
}

// ChangePlaylist adds the track to a playlist, or removes it. The
// playlist is picked among the user's own, never created: an unknown name
// is an error and an ambiguous one a list to ask about. A track already
// in the playlist is not added twice.
func (s *Service) ChangePlaylist(ctx context.Context, a Actor, req PlaylistRequest) (ChangeView, error) {
	var out ChangeView
	err := s.call(ctx, a, true, func(token string) error {
		c := s.client()
		lists, err := c.Playlists(ctx, token)
		if err != nil {
			return err
		}
		list, err := pick(lists, req.Playlist)
		if err != nil {
			return err
		}
		me, err := c.Me(ctx, token)
		if err != nil {
			return err
		}
		if list.Owner.ID != me.ID && !list.Collaborative {
			return fmt.Errorf("%s belongs to %s and is not collaborative: you can't change it", list.Name, list.Owner.DisplayName)
		}
		id, label, err := s.target(ctx, token, req.Track)
		if err != nil {
			return err
		}
		items, err := c.PlaylistItems(ctx, token, list.ID)
		if err != nil {
			return err
		}
		in := slices.ContainsFunc(items, func(t Track) bool { return t.ID == id })
		uri := []string{"spotify:track:" + id}
		out, err = change(in, !req.Remove, label, list.Name,
			func() error { return c.AddItems(ctx, token, list.ID, uri) },
			func() error { return c.RemoveItems(ctx, token, list.ID, uri) })
		return err
	})
	return out, err
}
