package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/titpetric/task-ui/spotify"
)

// spotifyTools are the tools of the user's own Spotify account: they act
// as the user on the account the user linked, through the API's routes,
// so a refusal names the permission it needs (spotify_link for the reads,
// spotify_control for the writes). They are a set of their own because no
// other tool reaches an account that is not task-ui's; docs/guide/mcp.md.
func spotifyTools() []Tool {
	str := func(text string) Property { return Property{Type: "string", Description: text} }
	track := str("A track: its Spotify ID, spotify:track: URI or open.spotify.com link; the track that is playing when empty.")
	return []Tool{
		{
			Name:        "spotify_now_playing",
			Description: "Show what is playing on the Spotify account you linked: the track, artist and device. Needs spotify_link.",
			InputSchema: object(map[string]Property{}),
			call:        spotifyNowPlaying,
		},
		{
			Name:        "spotify_search",
			Description: "Search Spotify for tracks. Needs spotify_link. To play one, call spotify_playback with its URI.",
			InputSchema: object(map[string]Property{
				"query": str("What to search for: a title, an artist, or both."),
				"limit": {Type: "integer", Description: "How many tracks, 10 unless given, 50 at most."},
			}, "query"),
			call: spotifySearch,
		},
		{
			Name:        "spotify_playlists",
			Description: "List the playlists of the Spotify account you linked, or with playlist what is in one. Needs spotify_link.",
			InputSchema: object(map[string]Property{"playlist": str("A playlist's name or ID, to list its tracks.")}),
			call:        spotifyPlaylists,
		},
		{
			Name: "spotify_playback",
			Description: "Control playback on the Spotify account you linked. Needs spotify_control, Spotify Premium and an active device; " +
				"the answer says plainly when one is missing.",
			InputSchema: object(map[string]Property{
				"action": {Type: "string", Description: "What to do.", Enum: []string{"play", "pause", "next", "previous", "volume"}},
				"volume": {Type: "integer", Description: "The level for volume, 0 to 100."},
				"query":  str("For play: search for this and play the first track."),
				"uri":    str("For play: a spotify: URI of a track, album, artist or playlist to play. Play with neither resumes."),
			}, "action"),
			call: spotifyPlayback,
		},
		{
			Name: "spotify_like",
			Description: "Save a track to the Liked Songs of the Spotify account you linked, or with unlike remove it. " +
				"Idempotent: a track already in the state asked for changes nothing and the answer says so. Needs spotify_control.",
			InputSchema: object(map[string]Property{
				"track":  track,
				"unlike": {Type: "boolean", Description: "Remove the track from the Liked Songs instead."},
			}),
			call: spotifyLike,
		},
		{
			Name: "spotify_playlist_add",
			Description: "Add a track to one of your Spotify playlists, picked by name or ID. It never creates a playlist and never adds a track twice; " +
				"a name that fits several playlists is an error that lists them, so ask the user which. Needs spotify_control.",
			InputSchema: object(map[string]Property{
				"playlist": str("The playlist's name or ID, from spotify_playlists."),
				"track":    track,
			}, "playlist"),
			call: spotifyPlaylistAdd,
		},
		{
			Name: "spotify_playlist_create",
			Description: "Make a Spotify playlist of your own and fill it with the tracks you collected: each one an ID, a spotify:track: URI, an " +
				"open.spotify.com link, or words to search for, which take the first hit. The answer lists what went in and what nothing was found " +
				"for. Needs spotify_control.",
			InputSchema: object(map[string]Property{
				"name":        str("What the playlist is called."),
				"description": str("What the playlist is for, shown under its name."),
				"public":      {Type: "boolean", Description: "Make it public. Private unless set."},
				"tracks": {Type: "array", Description: "The tracks to put in it, in the order they go in, at most 500.", Items: &Schema{Type: "string"}},
			}, "name"),
			call: spotifyPlaylistCreate,
		},
		{
			Name:        "spotify_playlist_remove",
			Description: "Remove a track from one of your Spotify playlists, picked by name or ID; a track that is not in it changes nothing. Needs spotify_control.",
			InputSchema: object(map[string]Property{
				"playlist": str("The playlist's name or ID, from spotify_playlists."),
				"track":    track,
			}, "playlist"),
			call: spotifyPlaylistRemove,
		},
	}
}

func spotifyNowPlaying(ctx context.Context, c *client, _ json.RawMessage) (string, error) {
	var view spotify.PlayingView
	if err := c.get(ctx, "/spotify/now-playing", &view); err != nil {
		return "", err
	}
	switch {
	case !view.Active:
		return "Nothing is playing: no Spotify device is active.", nil
	case view.Track == "":
		return fmt.Sprintf("Nothing is loaded on %s.", view.Device), nil
	}
	state := "Paused"
	if view.Playing {
		state = "Playing"
	}
	return fmt.Sprintf("%s: %s - %s (%s) on %s (%s), volume %d%%.\nTrack ID %s, URI %s", state, view.Track, view.Artists, view.Album, view.Device, view.Kind, view.Volume, view.TrackID, view.URI), nil
}

type spotifySearchArgs struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

func spotifySearch(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in spotifySearchArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	path := "/spotify/search?q=" + url.QueryEscape(in.Query)
	if in.Limit > 0 {
		path += "&limit=" + strconv.Itoa(in.Limit)
	}
	var out struct {
		Tracks []spotify.Track `json:"tracks"`
	}
	if err := c.get(ctx, path, &out); err != nil {
		return "", err
	}
	if len(out.Tracks) == 0 {
		return "Nothing found.", nil
	}
	var b strings.Builder
	for _, t := range out.Tracks {
		fmt.Fprintf(&b, "%s | %s | %s\n", t.Label(), t.Album.Name, t.URI)
	}
	return b.String(), nil
}

type spotifyPlaylistsArgs struct {
	Playlist string `json:"playlist"`
}

func spotifyPlaylists(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in spotifyPlaylistsArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	var b strings.Builder
	if in.Playlist != "" {
		var out struct {
			Playlist spotify.Playlist `json:"playlist"`
			Tracks   []spotify.Track  `json:"tracks"`
		}
		if err := c.get(ctx, "/spotify/playlists?playlist="+url.QueryEscape(in.Playlist), &out); err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "%s (%d tracks)\n", out.Playlist.Name, len(out.Tracks))
		for _, t := range out.Tracks {
			fmt.Fprintf(&b, "  %s | %s\n", t.Label(), t.URI)
		}
		return b.String(), nil
	}
	var out struct {
		Playlists []spotify.Playlist `json:"playlists"`
	}
	if err := c.get(ctx, "/spotify/playlists", &out); err != nil {
		return "", err
	}
	if len(out.Playlists) == 0 {
		return "No playlists.", nil
	}
	for _, p := range out.Playlists {
		fmt.Fprintf(&b, "%s | by %s | id %s\n", p.Name, p.Owner.DisplayName, p.ID)
	}
	return b.String(), nil
}

func spotifyPlayback(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in spotify.ControlRequest
	if err := decode(args, &in); err != nil {
		return "", err
	}
	return spotifyMessage(ctx, c, "/spotify/playback", in)
}

func spotifyLike(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in spotify.LikeRequest
	if err := decode(args, &in); err != nil {
		return "", err
	}
	return spotifyMessage(ctx, c, "/spotify/like", in)
}

func spotifyPlaylistAdd(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in spotify.PlaylistRequest
	if err := decode(args, &in); err != nil {
		return "", err
	}
	in.Remove = false
	return spotifyMessage(ctx, c, "/spotify/playlist", in)
}

func spotifyPlaylistCreate(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in spotify.NewPlaylistRequest
	if err := decode(args, &in); err != nil {
		return "", err
	}
	data, err := c.call(ctx, http.MethodPost, "/spotify/playlists", in)
	if err != nil {
		return "", err
	}
	var out spotify.PlaylistView
	if err := json.Unmarshal(data, &out); err != nil {
		return "", err
	}
	lines := []string{out.Message, "ID " + out.ID + ", URI " + out.URI}
	for _, track := range out.Added {
		lines = append(lines, "+ "+track)
	}
	for _, miss := range out.Missed {
		lines = append(lines, "? nothing found for "+miss)
	}
	return strings.Join(lines, "\n"), nil
}

func spotifyPlaylistRemove(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in spotify.PlaylistRequest
	if err := decode(args, &in); err != nil {
		return "", err
	}
	in.Remove = true
	return spotifyMessage(ctx, c, "/spotify/playlist", in)
}

// spotifyMessage posts a change and returns the message of its answer.
func spotifyMessage(ctx context.Context, c *client, path string, body any) (string, error) {
	data, err := c.call(ctx, http.MethodPost, path, body)
	if err != nil {
		return "", err
	}
	var out struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", err
	}
	return out.Message, nil
}
