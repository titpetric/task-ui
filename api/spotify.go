package api

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/spotify"
)

// spotifyActor is the user of a request as the Spotify service knows them.
func spotifyActor(r *http.Request) spotify.Actor {
	user := access.UserFrom(r.Context())
	return spotify.Actor{Name: user.Name, Role: user.EffectiveRole()}
}

// spotifySession names the session a link starts in without keeping it:
// the hash of the session cookie, or of the bearer token's ID. The state
// of a link is bound to it, so a callback from another session is refused.
func spotifySession(r *http.Request) string {
	value := ""
	if cookie, err := r.Cookie(access.Cookie); err == nil {
		value = "cookie:" + cookie.Value
	} else if id := access.TokenFrom(r.Context()); id != "" {
		value = "token:" + id
	}
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

// spotifyStatusOf maps what the Spotify service says to HTTP statuses.
func spotifyStatusOf(err error) int {
	var ambiguous *spotify.AmbiguousError
	var api *spotify.Error
	switch {
	case errors.Is(err, spotify.ErrOff):
		return http.StatusNotFound
	case errors.Is(err, spotify.ErrWho), errors.Is(err, spotify.ErrScope):
		return http.StatusForbidden
	case errors.Is(err, spotify.ErrNoLink), errors.Is(err, spotify.ErrBroken), errors.Is(err, spotify.ErrNoDevice),
		errors.Is(err, spotify.ErrPremium), errors.As(err, &ambiguous):
		return http.StatusConflict
	case errors.Is(err, spotify.ErrLimited), errors.Is(err, spotify.ErrRateLimited):
		return http.StatusTooManyRequests
	case errors.As(err, &api), errors.Is(err, spotify.ErrUnauthorized):
		return http.StatusBadGateway
	}
	return http.StatusBadRequest
}

// spotifyFail answers with the status of what Spotify's service said.
func spotifyFail(w http.ResponseWriter, err error) {
	fail(w, spotifyStatusOf(err), err.Error())
}

// spotifyOn answers 404 when the server has no Spotify service.
func (h *Handlers) spotifyOn(w http.ResponseWriter) bool {
	if h.opts.Spotify == nil {
		fail(w, http.StatusNotFound, spotify.ErrOff.Error())
		return false
	}
	return true
}

// spotifyRead checks what the reads of an account take: the service and
// spotify_link.
func (h *Handlers) spotifyRead(w http.ResponseWriter, r *http.Request) bool {
	return h.spotifyOn(w) && h.require(w, r, "spotify_link")
}

// spotifyWrite checks what a change of an account takes: the service and
// spotify_control.
func (h *Handlers) spotifyWrite(w http.ResponseWriter, r *http.Request) bool {
	return h.spotifyOn(w) && h.require(w, r, "spotify_control")
}

// spotifyAuthorize sends the browser to Spotify's consent page for the
// user's own account. Takes spotify_link. A client that asks for JSON gets
// the address instead.
func (h *Handlers) spotifyAuthorize(w http.ResponseWriter, r *http.Request) {
	if !h.spotifyRead(w, r) {
		return
	}
	target, err := h.opts.Spotify.Begin(r.Context(), spotifyActor(r), spotifySession(r))
	if err != nil {
		spotifyFail(w, err)
		return
	}
	if r.Header.Get("Accept") == "application/json" {
		writeJSON(w, http.StatusOK, map[string]string{"url": target})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, target, http.StatusSeeOther)
}

var spotifyErrorCode = regexp.MustCompile(`^[a-z_]{1,40}$`)

// SpotifyCallback answers where Spotify sends the browser back. No session
// signs it in: the session cookie is SameSite=Strict, so a navigation that
// comes from Spotify carries none. It holds nothing and trusts nothing, and
// hands the browser on to the finish route from a page of our own, a
// navigation that does carry the cookie.
func SpotifyCallback(w http.ResponseWriter, r *http.Request) {
	q := url.Values{}
	for _, name := range []string{"code", "state", "error"} {
		if v := r.URL.Query().Get(name); v != "" && len(v) <= 2048 {
			q.Set(name, v)
		}
	}
	next := spotify.FinishPath + "?" + q.Encode()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	escaped := html.EscapeString(next)
	_, _ = w.Write([]byte(`<!doctype html><html><head><meta charset="utf-8"><meta name="referrer" content="no-referrer">` +
		`<meta http-equiv="refresh" content="0;url=` + escaped + `"><title>Linking Spotify</title></head>` +
		`<body><p>Linking Spotify… <a href="` + escaped + `">Continue</a></p></body></html>`))
}

// spotifyFinish completes a link inside the user's session: the state has
// to be the one this user's session started, once. Takes spotify_link.
func (h *Handlers) spotifyFinish(w http.ResponseWriter, r *http.Request) {
	if !h.spotifyRead(w, r) {
		return
	}
	q := r.URL.Query()
	if code := q.Get("error"); code != "" {
		// Burn the state, then say what Spotify said, if it looks like a code.
		_ = h.opts.Spotify.Complete(r.Context(), spotifyActor(r), spotifySession(r), q.Get("state"), "")
		if !spotifyErrorCode.MatchString(code) {
			code = "unknown"
		}
		fail(w, http.StatusBadRequest, "Spotify did not link the account: "+code)
		return
	}
	if err := h.opts.Spotify.Complete(r.Context(), spotifyActor(r), spotifySession(r), q.Get("state"), q.Get("code")); err != nil {
		spotifyFail(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, "/settings/integrations?spotify=linked#spotify", http.StatusSeeOther)
}

// spotifyLink reports whether the user has linked an account, and whether
// the integration is open to them. Takes spotify_link. No token is in it.
func (h *Handlers) spotifyLink(w http.ResponseWriter, r *http.Request) {
	if !h.spotifyRead(w, r) {
		return
	}
	status, err := h.opts.Spotify.Status(r.Context(), spotifyActor(r))
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

// spotifyUnlink deletes the stored tokens of the user's account. A user
// disconnects their own; an admin may name another with ?user=. It takes
// no permission beyond being signed in: taking your own data back is
// always allowed.
func (h *Handlers) spotifyUnlink(w http.ResponseWriter, r *http.Request) {
	if !h.spotifyOn(w) {
		return
	}
	user := access.UserFrom(r.Context())
	if err := h.opts.Spotify.Unlink(r.Context(), user.Name, r.URL.Query().Get("user"), user.CanAdmin()); err != nil {
		fail(w, http.StatusForbidden, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// spotifyNowPlaying returns the track, artist and device. Takes spotify_link.
func (h *Handlers) spotifyNowPlaying(w http.ResponseWriter, r *http.Request) {
	if !h.spotifyRead(w, r) {
		return
	}
	view, err := h.opts.Spotify.NowPlaying(r.Context(), spotifyActor(r))
	if err != nil {
		spotifyFail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

// spotifySearch finds tracks by ?q=, ?limit= of them. Takes spotify_link.
func (h *Handlers) spotifySearch(w http.ResponseWriter, r *http.Request) {
	if !h.spotifyRead(w, r) {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	tracks, err := h.opts.Spotify.Search(r.Context(), spotifyActor(r), r.URL.Query().Get("q"), limit)
	if err != nil {
		spotifyFail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tracks": tracks})
}

// spotifyPlaylists lists the playlists, or with ?playlist= (a name or ID)
// one with its tracks. Takes spotify_link.
func (h *Handlers) spotifyPlaylists(w http.ResponseWriter, r *http.Request) {
	if !h.spotifyRead(w, r) {
		return
	}
	actor := spotifyActor(r)
	if name := r.URL.Query().Get("playlist"); name != "" {
		list, tracks, err := h.opts.Spotify.PlaylistContents(r.Context(), actor, name)
		if err != nil {
			spotifyFail(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"playlist": list, "tracks": tracks})
		return
	}
	lists, err := h.opts.Spotify.Playlists(r.Context(), actor)
	if err != nil {
		spotifyFail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"playlists": lists})
}

// spotifyPlayback runs a playback command: play, pause, next, previous or
// volume. Takes spotify_control.
func (h *Handlers) spotifyPlayback(w http.ResponseWriter, r *http.Request) {
	if !h.spotifyWrite(w, r) {
		return
	}
	var req spotify.ControlRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	message, err := h.opts.Spotify.Control(r.Context(), spotifyActor(r), req)
	if err != nil {
		spotifyFail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": message})
}

// spotifyLike saves a track to the Liked Songs, or removes it, and says
// when nothing changed. Takes spotify_control.
func (h *Handlers) spotifyLike(w http.ResponseWriter, r *http.Request) {
	if !h.spotifyWrite(w, r) {
		return
	}
	var req spotify.LikeRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.opts.Spotify.Like(r.Context(), spotifyActor(r), req)
	if err != nil {
		spotifyFail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// spotifyPlaylistChange adds a track to a playlist, or with remove takes it
// out. Takes spotify_control.
func (h *Handlers) spotifyPlaylistChange(w http.ResponseWriter, r *http.Request) {
	if !h.spotifyWrite(w, r) {
		return
	}
	var req spotify.PlaylistRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.opts.Spotify.ChangePlaylist(r.Context(), spotifyActor(r), req)
	if err != nil {
		spotifyFail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
