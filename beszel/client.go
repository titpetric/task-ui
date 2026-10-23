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
	"net/url"
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

	containers       map[string][]Container
	containersAt     time.Time
	history          map[string][]Sample
	historyAt        time.Time
	containerHistory map[string]containerHistory
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
	c.stats(ctx, list.Items)
	c.systems, c.at = list.Items, time.Now()
	return c.systems, nil
}

// statRecord is a row of the system_stats collection.
type statRecord struct {
	System string `json:"system"`
	Stats  Stats  `json:"stats"`
}

// statList is a page of the system_stats collection.
type statList struct {
	Items []statRecord `json:"items"`
}

// stats fills in each system's latest minute of measures. The hub writes
// one a minute per host, so the newest records of the last ten minutes
// hold every host that reports; a host without one keeps no stats.
func (c *Client) stats(ctx context.Context, systems []System) {
	since := time.Now().UTC().Add(-10 * time.Minute).Format("2006-01-02 15:04:05")
	filter := url.QueryEscape("type='1m' && created>='" + since + "'")
	var list statList
	if _, err := c.do(ctx, "/api/collections/system_stats/records?perPage=500&sort=-created&filter="+filter, &list); err != nil {
		return
	}
	latest := map[string]Stats{}
	for _, record := range list.Items {
		if _, ok := latest[record.System]; !ok {
			latest[record.System] = record.Stats
		}
	}
	for i := range systems {
		if stats, ok := latest[systems[i].ID]; ok {
			stats := stats
			systems[i].Stats = &stats
		}
	}
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
