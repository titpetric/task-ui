package spotify

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DefaultAPI is the Web API's base address.
const DefaultAPI = "https://api.spotify.com/v1"

// The conditions a caller tells a user about plainly.
var (
	// ErrUnauthorized reports a token Spotify refused: it expired or was revoked.
	ErrUnauthorized = errors.New("spotify refused the access token")
	// ErrPremium reports a playback command on an account without Premium.
	ErrPremium = errors.New("playback control needs Spotify Premium")
	// ErrNoDevice reports a playback command with no active device.
	ErrNoDevice = errors.New("no active Spotify device: open Spotify on a phone, computer or speaker and start something there")
	// ErrScope reports a call the granted scopes do not cover.
	ErrScope = errors.New("the link does not allow this: an admin has to enable write access, then link the account again")
)

// Error is an answer of the Web API that is no success.
type Error struct {
	Status  int
	Reason  string
	Message string
}

func (e *Error) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("spotify answered %d", e.Status)
	}
	return fmt.Sprintf("spotify answered %d: %s", e.Status, e.Message)
}

// Client calls the Web API with an access token it is handed per call, so
// one client serves every user. Every request has a timeout, and a 429 is
// retried after the Retry-After it names, up to Retries times.
type Client struct {
	// API is the base address; empty is DefaultAPI.
	API string
	// HTTP sends the requests; nil is a client with a 10 second timeout.
	HTTP *http.Client
	// Retries is how many times a 429 is retried; 0 is 2.
	Retries int
	// MaxWait caps a single Retry-After wait; 0 is 10 seconds. A longer
	// wait answers the caller instead.
	MaxWait time.Duration
	// Sleep waits between retries; nil waits on the clock.
	Sleep func(ctx context.Context, d time.Duration) error
}

// ErrRateLimited reports a 429 that outlasted the retries.
var ErrRateLimited = errors.New("spotify is rate limiting this app; try again in a moment")

func (c *Client) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 10 * time.Second}
}

func (c *Client) base() string {
	if c.API != "" {
		return strings.TrimRight(c.API, "/")
	}
	return DefaultAPI
}

func wait(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// do sends a request with the bearer token and decodes the answer into out
// when there is one. A 429 is retried after the Retry-After it names.
func (c *Client) do(ctx context.Context, token, method, path string, query url.Values, body, out any) error {
	var payload []byte
	if body != nil {
		var err error
		if payload, err = json.Marshal(body); err != nil {
			return err
		}
	}
	target := c.base() + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	retries, maxWait, sleep := cmp.Or(c.Retries, 2), cmp.Or(c.MaxWait, 10*time.Second), c.Sleep
	if sleep == nil {
		sleep = wait
	}
	for attempt := 0; ; attempt++ {
		status, header, data, err := c.send(ctx, token, method, target, payload)
		if err != nil {
			return err
		}
		switch {
		case status == http.StatusTooManyRequests:
			after := retryAfter(header)
			if attempt >= retries || after > maxWait {
				return ErrRateLimited
			}
			if err := sleep(ctx, after); err != nil {
				return err
			}
		case status >= http.StatusBadRequest:
			return readError(status, data)
		case out == nil || len(bytes.TrimSpace(data)) == 0 || status == http.StatusNoContent:
			return nil
		default:
			return json.Unmarshal(data, out)
		}
	}
}

// send makes one request and reads the answer whole.
func (c *Client) send(ctx context.Context, token, method, target string, payload []byte) (int, http.Header, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, target, bytes.NewReader(payload))
	if err != nil {
		return 0, nil, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http().Do(req)
	if err != nil {
		// A transport error can name the URL but never the header.
		return 0, nil, nil, fmt.Errorf("spotify unreachable: %w", unwrapURL(err))
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	return resp.StatusCode, resp.Header, data, err
}

// retryAfter reads the wait a 429 asks for; a second when it names none.
func retryAfter(h http.Header) time.Duration {
	secs, err := strconv.Atoi(h.Get("Retry-After"))
	if err != nil {
		return time.Second
	}
	return time.Duration(secs) * time.Second
}

// unwrapURL drops the URL a transport error carries, which may hold a query.
func unwrapURL(err error) error {
	var uerr *url.Error
	if errors.As(err, &uerr) {
		return uerr.Err
	}
	return err
}

// readError turns an error answer into the typed conditions, or an Error.
func readError(status int, data []byte) error {
	var body struct {
		Error struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
			Reason  string `json:"reason"`
		} `json:"error"`
	}
	_ = json.Unmarshal(data, &body)
	e := &Error{Status: status, Reason: body.Error.Reason, Message: body.Error.Message}
	switch {
	case status == http.StatusUnauthorized:
		return ErrUnauthorized
	case e.Reason == "PREMIUM_REQUIRED":
		return ErrPremium
	case e.Reason == "NO_ACTIVE_DEVICE":
		return ErrNoDevice
	case status == http.StatusForbidden && strings.Contains(strings.ToLower(e.Message), "scope"):
		return ErrScope
	}
	return e
}

// User is a Spotify account.
type User struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Product     string `json:"product"`
}

// Me reads the account the token belongs to.
func (c *Client) Me(ctx context.Context, token string) (*User, error) {
	var u User
	return &u, c.do(ctx, token, http.MethodGet, "/me", nil, nil, &u)
}

// Artist, Album, Track are the parts of a track the answers use.
type Artist struct {
	Name string `json:"name"`
}

type Album struct {
	Name string `json:"name"`
}

type Track struct {
	ID      string   `json:"id"`
	URI     string   `json:"uri"`
	Name    string   `json:"name"`
	Artists []Artist `json:"artists"`
	Album   Album    `json:"album"`
}

// Label is "Name - Artist, Artist".
func (t Track) Label() string {
	if len(t.Artists) == 0 {
		return t.Name
	}
	return t.Name + " - " + t.ArtistNames()
}

// ArtistNames joins the artists with commas.
func (t Track) ArtistNames() string {
	names := make([]string, 0, len(t.Artists))
	for _, a := range t.Artists {
		names = append(names, a.Name)
	}
	return strings.Join(names, ", ")
}

// Device is a place Spotify plays.
type Device struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	IsActive      bool   `json:"is_active"`
	VolumePercent int    `json:"volume_percent"`
}

// Playback is the player's state; a nil Item is nothing loaded or an ad.
type Playback struct {
	IsPlaying bool   `json:"is_playing"`
	Device    Device `json:"device"`
	Item      *Track `json:"item"`
	Progress  int    `json:"progress_ms"`
}

// Playback reads the player's state; nil is no active device.
func (c *Client) Playback(ctx context.Context, token string) (*Playback, error) {
	var p Playback
	data := json.RawMessage{}
	if err := c.do(ctx, token, http.MethodGet, "/me/player", nil, nil, &data); err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return &p, json.Unmarshal(data, &p)
}

// PlayOptions say what to play; none resumes.
type PlayOptions struct {
	URIs       []string `json:"uris,omitempty"`
	ContextURI string   `json:"context_uri,omitempty"`
}

// Play starts or resumes playback.
func (c *Client) Play(ctx context.Context, token string, o PlayOptions) error {
	var body any
	if len(o.URIs) > 0 || o.ContextURI != "" {
		body = o
	}
	return c.do(ctx, token, http.MethodPut, "/me/player/play", nil, body, nil)
}

// Pause pauses playback.
func (c *Client) Pause(ctx context.Context, token string) error {
	return c.do(ctx, token, http.MethodPut, "/me/player/pause", nil, nil, nil)
}

// Next skips to the next track.
func (c *Client) Next(ctx context.Context, token string) error {
	return c.do(ctx, token, http.MethodPost, "/me/player/next", nil, nil, nil)
}

// Previous goes back a track.
func (c *Client) Previous(ctx context.Context, token string) error {
	return c.do(ctx, token, http.MethodPost, "/me/player/previous", nil, nil, nil)
}

// Volume sets the volume, 0 to 100.
func (c *Client) Volume(ctx context.Context, token string, percent int) error {
	if percent < 0 || percent > 100 {
		return errors.New("volume is 0 to 100")
	}
	return c.do(ctx, token, http.MethodPut, "/me/player/volume", url.Values{"volume_percent": {strconv.Itoa(percent)}}, nil, nil)
}

// Search finds tracks by a query.
func (c *Client) Search(ctx context.Context, token, query string, limit int) ([]Track, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	var out struct {
		Tracks struct {
			Items []Track `json:"items"`
		} `json:"tracks"`
	}
	q := url.Values{"q": {query}, "type": {"track"}, "limit": {strconv.Itoa(limit)}}
	if err := c.do(ctx, token, http.MethodGet, "/search", q, nil, &out); err != nil {
		return nil, err
	}
	return out.Tracks.Items, nil
}

// Playlist is a playlist of the user's library.
type Playlist struct {
	ID    string `json:"id"`
	URI   string `json:"uri"`
	Name  string `json:"name"`
	Owner struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
	} `json:"owner"`
	Collaborative bool `json:"collaborative"`
	Tracks        struct {
		Total int `json:"total"`
	} `json:"tracks"`
}

// Playlists lists every playlist of the user, following the pages.
func (c *Client) Playlists(ctx context.Context, token string) ([]Playlist, error) {
	var all []Playlist
	for offset := 0; ; offset += 50 {
		var page struct {
			Items []Playlist `json:"items"`
			Next  string     `json:"next"`
		}
		q := url.Values{"limit": {"50"}, "offset": {strconv.Itoa(offset)}}
		if err := c.do(ctx, token, http.MethodGet, "/me/playlists", q, nil, &page); err != nil {
			return nil, err
		}
		all = append(all, page.Items...)
		if page.Next == "" || len(page.Items) == 0 || len(all) >= 1000 {
			return all, nil
		}
	}
}

// NewPlaylist is a playlist to make: its name, what it says it is, and
// whether anyone may see it.
type NewPlaylist struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Public      bool   `json:"public"`
}

// CreatePlaylist makes a playlist under the user and returns it.
func (c *Client) CreatePlaylist(ctx context.Context, token, userID string, p NewPlaylist) (*Playlist, error) {
	var out Playlist
	return &out, c.do(ctx, token, http.MethodPost, "/users/"+url.PathEscape(userID)+"/playlists", nil, p, &out)
}

// PlaylistItems lists the tracks of a playlist, following the pages.
func (c *Client) PlaylistItems(ctx context.Context, token, id string) ([]Track, error) {
	var all []Track
	for offset := 0; ; offset += 100 {
		var page struct {
			Items []struct {
				Item  *Track `json:"item"`
				Track *Track `json:"track"`
			} `json:"items"`
			Next string `json:"next"`
		}
		q := url.Values{"limit": {"100"}, "offset": {strconv.Itoa(offset)}}
		if err := c.do(ctx, token, http.MethodGet, "/playlists/"+url.PathEscape(id)+"/items", q, nil, &page); err != nil {
			return nil, err
		}
		for _, entry := range page.Items {
			switch {
			case entry.Item != nil:
				all = append(all, *entry.Item)
			case entry.Track != nil:
				all = append(all, *entry.Track)
			}
		}
		if page.Next == "" || len(page.Items) == 0 || len(all) >= 5000 {
			return all, nil
		}
	}
}

// AddItems appends tracks to a playlist.
func (c *Client) AddItems(ctx context.Context, token, id string, uris []string) error {
	return c.do(ctx, token, http.MethodPost, "/playlists/"+url.PathEscape(id)+"/items", nil, map[string]any{"uris": uris}, nil)
}

// RemoveItems removes every occurrence of tracks from a playlist.
func (c *Client) RemoveItems(ctx context.Context, token, id string, uris []string) error {
	items := make([]map[string]string, 0, len(uris))
	for _, uri := range uris {
		items = append(items, map[string]string{"uri": uri})
	}
	return c.do(ctx, token, http.MethodDelete, "/playlists/"+url.PathEscape(id)+"/items", nil, map[string]any{"items": items}, nil)
}

// Saved reports which tracks are in the Liked Songs.
func (c *Client) Saved(ctx context.Context, token string, ids []string) ([]bool, error) {
	var out []bool
	err := c.do(ctx, token, http.MethodGet, "/me/tracks/contains", url.Values{"ids": {strings.Join(ids, ",")}}, nil, &out)
	return out, err
}

// Save adds tracks to the Liked Songs.
func (c *Client) Save(ctx context.Context, token string, ids []string) error {
	return c.do(ctx, token, http.MethodPut, "/me/tracks", nil, map[string]any{"ids": ids}, nil)
}

// Unsave removes tracks from the Liked Songs.
func (c *Client) Unsave(ctx context.Context, token string, ids []string) error {
	return c.do(ctx, token, http.MethodDelete, "/me/tracks", nil, map[string]any{"ids": ids}, nil)
}
