package spotify

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// DefaultAccounts is Spotify's account service.
const DefaultAccounts = "https://accounts.spotify.com"

// ErrBroken reports a link Spotify no longer honours: the refresh token
// was revoked or expired (invalid_grant). The user links again.
var ErrBroken = errors.New("the Spotify link is broken: link your Spotify again")

// Token is what a link holds of the account.
type Token struct {
	Access  string
	Refresh string
	Expires time.Time
	Scope   string
}

// OAuth speaks to the account service: the code exchange and the refresh.
type OAuth struct {
	// Accounts is the service's base address; empty is DefaultAccounts.
	Accounts string
	// HTTP sends the requests; nil is a client with a 10 second timeout.
	HTTP *http.Client
	Now  func() time.Time
}

func (o *OAuth) accounts() string {
	if o.Accounts != "" {
		return strings.TrimRight(o.Accounts, "/")
	}
	return DefaultAccounts
}

func (o *OAuth) now() time.Time {
	if o.Now != nil {
		return o.Now()
	}
	return time.Now()
}

// AuthorizeURL is where the user's browser goes to grant access, with the
// PKCE challenge of the verifier.
func (o *OAuth) AuthorizeURL(clientID, redirect string, scopes []string, state, verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	q := url.Values{
		"client_id":             {clientID},
		"response_type":         {"code"},
		"redirect_uri":          {redirect},
		"scope":                 {strings.Join(scopes, " ")},
		"state":                 {state},
		"code_challenge_method": {"S256"},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(sum[:])},
	}
	return o.accounts() + "/authorize?" + q.Encode()
}

// Exchange trades the code of the redirect for tokens.
func (o *OAuth) Exchange(ctx context.Context, clientID, clientSecret, redirect, code, verifier string) (Token, error) {
	return o.token(ctx, clientID, clientSecret, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirect},
		"code_verifier": {verifier},
	}, "")
}

// Refresh trades the refresh token for a new access token. Spotify may
// rotate the refresh token; the old one stays when it does not.
func (o *OAuth) Refresh(ctx context.Context, clientID, clientSecret, refresh string) (Token, error) {
	return o.token(ctx, clientID, clientSecret, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refresh},
	}, refresh)
}

func (o *OAuth) token(ctx context.Context, clientID, clientSecret string, form url.Values, keep string) (Token, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.accounts()+"/api/token", strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)
	client := o.HTTP
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return Token{}, fmt.Errorf("spotify accounts unreachable: %w", unwrapURL(err))
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Token{}, err
	}
	var body struct {
		Access  string `json:"access_token"`
		Refresh string `json:"refresh_token"`
		Scope   string `json:"scope"`
		Expires int    `json:"expires_in"`
		Error   string `json:"error"`
		Desc    string `json:"error_description"`
	}
	_ = json.Unmarshal(data, &body)
	switch {
	case body.Error == "invalid_grant":
		return Token{}, ErrBroken
	case resp.StatusCode != http.StatusOK || body.Access == "":
		// The description is Spotify's own and holds no credential.
		return Token{}, fmt.Errorf("spotify accounts refused the request (%d): %s %s", resp.StatusCode, body.Error, body.Desc)
	}
	token := Token{Access: body.Access, Refresh: body.Refresh, Scope: body.Scope, Expires: o.now().Add(time.Duration(body.Expires) * time.Second)}
	if token.Refresh == "" {
		token.Refresh = keep
	}
	return token, nil
}

// pending is a link started and not finished.
type pending struct {
	user, session, verifier string
	expires                 time.Time
}

// StateTTL is how long a started link may take to finish.
const StateTTL = 10 * time.Minute

// States holds the links in progress. A state is random, bound to the
// user and their session, single use, and expires; the PKCE verifier
// never leaves the server.
type States struct {
	Now func() time.Time

	mu sync.Mutex
	m  map[string]pending
}

// ErrState reports a state that is missing, unknown, expired, used or not
// the user's.
var ErrState = errors.New("the link request is unknown, expired or not yours: start again from the integration page")

func (s *States) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func random(n int) string {
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

// New starts a link for a user in a session and returns its state and PKCE verifier.
func (s *States) New(user, session string) (state, verifier string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.m == nil {
		s.m = map[string]pending{}
	}
	now := s.now()
	for k, p := range s.m {
		if !now.Before(p.expires) {
			delete(s.m, k)
		}
	}
	state, verifier = random(32), random(64)
	s.m[state] = pending{user: user, session: session, verifier: verifier, expires: now.Add(StateTTL)}
	return state, verifier
}

// Take consumes a state and returns its verifier. The state is gone after
// the first try, a wrong user or session included.
func (s *States) Take(state, user, session string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.m[state]
	if state == "" || !ok {
		return "", ErrState
	}
	delete(s.m, state)
	if !s.now().Before(p.expires) || p.user != user || p.session != session {
		return "", ErrState
	}
	return p.verifier, nil
}
