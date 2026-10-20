// Package beszel reads the systems of a Beszel hub, which watches the hosts, for
// task-ui's host inventory. The hub is PocketBase, and the client lists its
// systems collection with a token.
package beszel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ErrToken reports a hub that refused the token: it expired, or the
// user's password changed, which revokes every token of the user.
var ErrToken = errors.New("beszel refused the token")

// Client reads a hub's systems with a token, and keeps the last answer
// for TTL so a page that shows them does not call the hub on every view.
//
// The token is an impersonation token of a read-only user, which a
// superuser mints with POST /api/collections/users/impersonate/{id} and a
// duration: nothing refreshes it and it needs no password, so task-ui keeps
// no login to the hub.
type Client struct {
	URL   string
	Token string
	TTL   time.Duration
	HTTP  *http.Client

	mu      sync.Mutex
	systems []System
	at      time.Time
}

// systemList is a page of the systems collection.
type systemList struct {
	Items []System `json:"items"`
}

// Systems returns the hub's systems by name, at most TTL old.
func (c *Client) Systems(ctx context.Context) ([]System, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.systems != nil && time.Since(c.at) < c.TTL {
		return c.systems, nil
	}
	var list systemList
	status, err := c.do(ctx, "/api/collections/systems/records?perPage=500&sort=name", &list)
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return nil, ErrToken
	}
	if err != nil {
		return nil, err
	}
	c.systems, c.at = list.Items, time.Now()
	return c.systems, nil
}

// do reads a path from the hub and decodes a success into out. It returns
// the status, and an error for anything but a success.
func (c *Client) do(ctx context.Context, path string, out any) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(c.URL, "/")+path, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", c.Token)
	client := c.HTTP
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if res.StatusCode >= 300 {
		return res.StatusCode, fmt.Errorf("beszel GET %s: %s", path, res.Status)
	}
	return res.StatusCode, json.Unmarshal(data, out)
}
