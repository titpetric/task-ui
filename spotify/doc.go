// Package spotify links a user's Spotify account to task-ui and acts on it
// as that user: what is playing, playback, search, playlists and the
// Liked Songs.
//
// The user links the account once with the authorization code flow and
// PKCE (Begin, Complete). The tokens live encrypted in the user's own
// secret scope, user:<name>, which no run reads and no page lists, and
// are refreshed on demand. Service is what the API serves; Client is the
// thin typed layer over Spotify's Web API.
package spotify
