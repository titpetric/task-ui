package spotify_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/integration/spotify"
	"github.com/titpetric/task-ui/internal/assert"
)

// The consent address carries PKCE, the state and the least scopes; the
// code exchange sends the verifier whose hash the challenge was.
func TestService_Link(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newFake(t)
	svc, tokens := f.service(map[string]string{"scopes": "read"})

	authorize, err := svc.Begin(ctx, tit, "s1")
	assert.NoError(t, err)
	u, _ := url.Parse(authorize)
	q := u.Query()
	assert.Equal(t, f.URL+"/authorize", u.Scheme+"://"+u.Host+u.Path)
	assert.Equal(t, "S256", q.Get("code_challenge_method"))
	assert.Equal(t, "https://task.example.com"+spotify.CallbackPath, q.Get("redirect_uri"))
	assert.Equal(t, strings.Join(spotify.ReadScopes, " "), q.Get("scope"), "write scopes stay off until an admin enables them")
	assert.NotContains(t, authorize, "client-secret-xyz")

	assert.NoError(t, svc.Complete(ctx, tit, "s1", q.Get("state"), "code"))
	sum := sha256.Sum256([]byte(f.verifier))
	assert.Equal(t, base64.RawURLEncoding.EncodeToString(sum[:]), q.Get("code_challenge"))
	link, _ := tokens.Load(ctx, "tit")
	assert.Equal(t, "Tit", link.Name)

	st, err := svc.Status(ctx, tit)
	assert.NoError(t, err)
	assert.True(t, st.Linked)
	assert.True(t, st.Available)
}

// The state is single use and bound to the user and the session.
func TestService_State(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newFake(t)
	svc, tokens := f.service(nil)

	state := func() string {
		authorize, _ := svc.Begin(ctx, tit, "s1")
		u, _ := url.Parse(authorize)
		return u.Query().Get("state")
	}
	// Another session, another user, a missing and a made up state.
	assert.ErrorIs(t, svc.Complete(ctx, tit, "other", state(), "c"), spotify.ErrState)
	assert.ErrorIs(t, svc.Complete(ctx, spotify.Actor{Name: "eve", Role: "runner"}, "s1", state(), "c"), spotify.ErrState)
	assert.ErrorIs(t, svc.Complete(ctx, tit, "s1", "", "c"), spotify.ErrState)
	assert.ErrorIs(t, svc.Complete(ctx, tit, "s1", "made-up", "c"), spotify.ErrState)
	// Single use: the second try with a good state fails.
	good := state()
	assert.NoError(t, svc.Complete(ctx, tit, "s1", good, "c"))
	assert.ErrorIs(t, svc.Complete(ctx, tit, "s1", good, "c"), spotify.ErrState)
	// A refused attempt used its state up as well.
	bad := state()
	assert.ErrorIs(t, svc.Complete(ctx, tit, "other", bad, "c"), spotify.ErrState)
	assert.ErrorIs(t, svc.Complete(ctx, tit, "s1", bad, "c"), spotify.ErrState)
	link, _ := tokens.Load(ctx, "tit")
	assert.NotNil(t, link)

	// Expiry after ten minutes.
	now := time.Now()
	states := &spotify.States{Now: func() time.Time { return now }}
	st, _ := states.New("tit", "s1")
	now = now.Add(spotify.StateTTL + time.Second)
	_, err := states.Take(st, "tit", "s1")
	assert.ErrorIs(t, err, spotify.ErrState)
}

func TestService_Config(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newFake(t)

	svc, _ := f.service(map[string]string{"client_id": ""})
	_, err := svc.Begin(ctx, tit, "s")
	assert.ErrorIs(t, err, spotify.ErrOff)

	svc, _ = f.service(map[string]string{"who": "admin"})
	_, err = svc.Begin(ctx, tit, "s")
	assert.ErrorIs(t, err, spotify.ErrWho)
	st, _ := svc.Status(ctx, tit)
	assert.False(t, st.Available)

	svc, _ = f.service(map[string]string{"redirect_uri": "http://task.example.com/cb"})
	_, err = svc.Begin(ctx, tit, "s")
	assert.ErrorIs(t, err, spotify.ErrRedirect, "plain http on a public host is refused")

	svc, _ = f.service(map[string]string{"redirect_uri": "http://127.0.0.1:3000/cb"})
	_, err = svc.Begin(ctx, tit, "s")
	assert.NoError(t, err)

	svc, _ = f.service(map[string]string{"redirect_uri": ""})
	svc.BaseURL = "https://task.example.com/"
	authorize, err := svc.Begin(ctx, tit, "s")
	assert.NoError(t, err)
	assert.Contains(t, authorize, url.QueryEscape("https://task.example.com"+spotify.CallbackPath))
}

// An expired access token is refreshed before the call; a 401 refreshes
// once and tries again; invalid_grant marks the link broken.
func TestService_Refresh(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newFake(t)
	svc, tokens := f.service(nil)
	link(t, svc)

	now := time.Now()
	svc.Now = func() time.Time { return now }
	view, err := svc.NowPlaying(ctx, tit)
	assert.NoError(t, err)
	assert.Equal(t, "Kitchen", view.Device)
	assert.Equal(t, "Artist", view.Artists)
	issued := f.issued

	// Expiry: the next call refreshes first.
	now = now.Add(2 * time.Hour)
	_, err = svc.NowPlaying(ctx, tit)
	assert.NoError(t, err)
	assert.Equal(t, issued+1, f.issued)

	// Revoked access token: 401, one refresh, one retry.
	l, _ := tokens.Load(ctx, "tit")
	l.Access = "stale-stale-stale"
	l.Expires = now.Add(time.Hour)
	assert.NoError(t, tokens.Save(ctx, "tit", *l))
	_, err = svc.NowPlaying(ctx, tit)
	assert.NoError(t, err)
	assert.Equal(t, issued+2, f.issued)

	// Revoked refresh token.
	f.grant = true
	now = now.Add(5 * time.Hour)
	_, err = svc.NowPlaying(ctx, tit)
	assert.ErrorIs(t, err, spotify.ErrBroken)
	assert.NotContains(t, err.Error(), f.refresh)
	st, _ := svc.Status(ctx, tit)
	assert.True(t, st.Broken)
	// A broken link makes no more requests to the account service.
	_, err = svc.NowPlaying(ctx, tit)
	assert.ErrorIs(t, err, spotify.ErrBroken)

	// Linking again heals it.
	f.grant = false
	link(t, svc)
	st, _ = svc.Status(ctx, tit)
	assert.False(t, st.Broken)
}

func TestService_NotLinked(t *testing.T) {
	t.Parallel()
	f := newFake(t)
	svc, _ := f.service(nil)
	_, err := svc.NowPlaying(context.Background(), tit)
	assert.ErrorIs(t, err, spotify.ErrNoLink)
}

// Unlinking deletes the tokens; only the user or an admin may.
func TestService_Unlink(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newFake(t)
	svc, tokens := f.service(nil)
	link(t, svc)

	assert.Error(t, svc.Unlink(ctx, "eve", "tit", false))
	l, _ := tokens.Load(ctx, "tit")
	assert.NotNil(t, l, "someone else's unlink changed nothing")
	assert.NoError(t, svc.Unlink(ctx, "root", "tit", true), "an admin may")
	l, _ = tokens.Load(ctx, "tit")
	assert.Nil(t, l)
	link(t, svc)
	assert.NoError(t, svc.Unlink(ctx, "tit", "", false))
	l, _ = tokens.Load(ctx, "tit")
	assert.Nil(t, l)
	_, err := svc.NowPlaying(ctx, tit)
	assert.ErrorIs(t, err, spotify.ErrNoLink)
}

// Liking is idempotent, and so is unliking.
func TestService_Like(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newFake(t)
	svc, _ := f.service(nil)
	link(t, svc)

	out, err := svc.Like(ctx, tit, spotify.LikeRequest{})
	assert.NoError(t, err)
	assert.True(t, out.Changed)
	assert.True(t, f.liked[trackA])
	assert.Contains(t, out.Message, "Song - Artist")

	puts := count(f.log(), "PUT /v1/me/tracks")
	out, err = svc.Like(ctx, tit, spotify.LikeRequest{})
	assert.NoError(t, err)
	assert.False(t, out.Changed)
	assert.Contains(t, out.Message, "already")
	assert.Equal(t, puts, count(f.log(), "PUT /v1/me/tracks"), "no second save")

	out, err = svc.Like(ctx, tit, spotify.LikeRequest{Unlike: true})
	assert.NoError(t, err)
	assert.True(t, out.Changed)
	assert.False(t, f.liked[trackA])
	out, err = svc.Like(ctx, tit, spotify.LikeRequest{Unlike: true})
	assert.NoError(t, err)
	assert.False(t, out.Changed)

	// A given track, as a link.
	out, err = svc.Like(ctx, tit, spotify.LikeRequest{Track: "https://open.spotify.com/track/" + trackB + "?si=x"})
	assert.NoError(t, err)
	assert.True(t, out.Changed)
	assert.True(t, f.liked[trackB])
	_, err = svc.Like(ctx, tit, spotify.LikeRequest{Track: "https://open.spotify.com/album/" + trackB})
	assert.Error(t, err)

	// Nothing playing and nothing named.
	f.playing = ""
	_, err = svc.Like(ctx, tit, spotify.LikeRequest{})
	assert.Error(t, err)
}

func count(log []string, call string) int {
	n := 0
	for _, c := range log {
		if c == call {
			n++
		}
	}
	return n
}

// A playlist is found by name, an ambiguous name is asked about, a track
// is never added twice, and a playlist is never created.
func TestService_Playlists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newFake(t)
	svc, _ := f.service(nil)
	link(t, svc)

	lists, err := svc.Playlists(ctx, tit)
	assert.NoError(t, err)
	assert.Len(t, lists, 4)

	_, err = svc.ChangePlaylist(ctx, tit, spotify.PlaylistRequest{Playlist: "road trip"})
	var ambiguous *spotify.AmbiguousError
	assert.True(t, errors.As(err, &ambiguous))
	assert.Len(t, ambiguous.Candidates, 2)
	assert.Equal(t, 0, count(f.log(), "POST /v1/me/playlists"))
	assert.Equal(t, 0, count(f.log(), "POST /v1/users/u1/playlists"))

	_, err = svc.ChangePlaylist(ctx, tit, spotify.PlaylistRequest{Playlist: "nope"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no playlist")

	_, err = svc.ChangePlaylist(ctx, tit, spotify.PlaylistRequest{Playlist: "Theirs"})
	assert.Error(t, err, "someone else's playlist")

	// The ID picks among equals.
	out, err := svc.ChangePlaylist(ctx, tit, spotify.PlaylistRequest{Playlist: "pl3"})
	assert.NoError(t, err)
	assert.True(t, out.Changed)
	assert.Equal(t, []string{trackA}, f.items["pl3"])
	out, err = svc.ChangePlaylist(ctx, tit, spotify.PlaylistRequest{Playlist: "focus"})
	assert.NoError(t, err)
	assert.False(t, out.Changed, "already there")
	assert.Equal(t, []string{trackA}, f.items["pl3"])

	out, err = svc.ChangePlaylist(ctx, tit, spotify.PlaylistRequest{Playlist: "Focus", Remove: true})
	assert.NoError(t, err)
	assert.True(t, out.Changed)
	out, err = svc.ChangePlaylist(ctx, tit, spotify.PlaylistRequest{Playlist: "Focus", Remove: true})
	assert.NoError(t, err)
	assert.False(t, out.Changed)

	list, tracks, err := svc.PlaylistContents(ctx, tit, "pl1")
	assert.NoError(t, err)
	assert.Equal(t, "Road trip", list.Name)
	assert.Len(t, tracks, 1)
}

// Playback tells a user plainly what is missing.
func TestService_Control(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newFake(t)
	svc, _ := f.service(nil)
	link(t, svc)

	msg, err := svc.Control(ctx, tit, spotify.ControlRequest{Action: "play", Query: "found"})
	assert.NoError(t, err)
	assert.Equal(t, "Playing Found - Artist.", msg)
	assert.Equal(t, 1, count(f.log(), "PUT /v1/me/player/play"))
	vol := 30
	msg, err = svc.Control(ctx, tit, spotify.ControlRequest{Action: "volume", Volume: &vol})
	assert.NoError(t, err)
	assert.Equal(t, "Volume set to 30.", msg)
	over := 130
	_, err = svc.Control(ctx, tit, spotify.ControlRequest{Action: "volume", Volume: &over})
	assert.Error(t, err)
	_, err = svc.Control(ctx, tit, spotify.ControlRequest{Action: "dance"})
	assert.Error(t, err)

	f.noDev = true
	_, err = svc.Control(ctx, tit, spotify.ControlRequest{Action: "next"})
	assert.ErrorIs(t, err, spotify.ErrNoDevice)
	f.noDev, f.premium = false, true
	_, err = svc.Control(ctx, tit, spotify.ControlRequest{Action: "pause"})
	assert.ErrorIs(t, err, spotify.ErrPremium)
}

// Write calls need the write scopes enabled by an admin, and are rate limited per user.
func TestService_Writes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newFake(t)
	svc, _ := f.service(map[string]string{"scopes": "read"})
	link(t, svc)
	_, err := svc.Control(ctx, tit, spotify.ControlRequest{Action: "pause"})
	assert.ErrorIs(t, err, spotify.ErrScope)
	_, err = svc.Like(ctx, tit, spotify.LikeRequest{})
	assert.ErrorIs(t, err, spotify.ErrScope)
	_, err = svc.NowPlaying(ctx, tit)
	assert.NoError(t, err, "reads are fine")

	svc, _ = f.service(nil)
	link(t, svc)
	svc.Limit = &spotify.Limiter{Max: 2, Window: time.Minute}
	for i := 0; i < 2; i++ {
		_, err = svc.Control(ctx, tit, spotify.ControlRequest{Action: "pause"})
		assert.NoError(t, err)
	}
	_, err = svc.Control(ctx, tit, spotify.ControlRequest{Action: "pause"})
	assert.ErrorIs(t, err, spotify.ErrLimited)
	// Another user has a window of their own.
	assert.True(t, svc.Limit.Allow("someone-else"))
}

// A credential never reaches an error: a message that holds a token or the
// app's secret is scrubbed before it leaves the service.
func TestService_Redaction(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newFake(t)
	svc, tokens := f.service(nil)
	link(t, svc)

	f.grant = true
	l, _ := tokens.Load(ctx, "tit")
	l.Expires = time.Now().Add(-time.Hour)
	assert.NoError(t, tokens.Save(ctx, "tit", *l))
	// invalid_grant echoes the refresh token in its description; the
	// service answers ErrBroken, which holds none of it.
	_, err := svc.NowPlaying(ctx, tit)
	assert.ErrorIs(t, err, spotify.ErrBroken)
	for _, secret := range []string{f.refresh, f.access, "client-secret-xyz"} {
		assert.NotContains(t, err.Error(), secret)
	}

	// A refusal that echoes the secret is scrubbed.
	f.grant = false
	l, _ = tokens.Load(ctx, "tit")
	l.Broken = false
	assert.NoError(t, tokens.Save(ctx, "tit", *l))
	bad := &spotify.Service{
		Settings: func(context.Context) (map[string]string, error) {
			return map[string]string{"client_id": "cid", "client_secret": "client-secret-xyz", "redirect_uri": "https://task.example.com/cb", "scopes": "read"}, nil
		},
		Tokens: tokens,
		OAuth:  &spotify.OAuth{Accounts: f.URL + "/nowhere-that-answers-with-the-secret-client-secret-xyz"},
		Client: svc.Client, States: &spotify.States{},
	}
	_, err = bad.NowPlaying(ctx, tit)
	if err != nil {
		assert.NotContains(t, err.Error(), "client-secret-xyz")
	}

	// A status view serialises without any credential.
	st, _ := svc.Status(ctx, tit)
	raw := fmtJSON(st)
	for _, secret := range []string{f.refresh, f.access, "client-secret-xyz"} {
		assert.NotContains(t, raw, secret)
	}
}
