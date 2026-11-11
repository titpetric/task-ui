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
var ErrToken = errors.New("beszel refused the token: it expired, or the hub user's password changed; enter a new one under Settings, Integrations")

// authPath answers 401 for a token the hub refuses.
const authPath = "/api/collections/users/auth-refresh"

// Client reads a hub's systems with a token, and keeps what the hub
// answered for TTL so a page that shows them calls it once.
//
// The token is a PocketBase auth token of a Beszel user: one a password
// login minted renews through Alive, and one minted by impersonation
// stands until its duration runs out.
type Client struct {
	URL   string
	Token string
	TTL   time.Duration
	HTTP  *http.Client

	mu      sync.Mutex
	systems []System
	at      time.Time
	// err is what the last read answered, kept as long as an answer.
	err error

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
// A refusal is kept for TTL as an answer is, because every page of the inventory reads this and a hub that refused the token refuses it again.
func (c *Client) Systems(ctx context.Context) ([]System, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.at.IsZero() && time.Since(c.at) < c.TTL {
		return c.systems, c.err
	}
	c.systems, c.err = c.read(ctx)
	c.at = time.Now()
	return c.systems, c.err
}

// read asks the hub for its systems and fills in the measures of each.
func (c *Client) read(ctx context.Context) ([]System, error) {
	var list systemList
	status, err := c.do(ctx, "/api/collections/systems/records?perPage=500&sort=name", &list)
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return nil, ErrToken
	}
	if err != nil {
		return nil, err
	}
	// An empty page is a hub with no host and a hub that refused the token
	// alike, and Alive is what tells them apart.
	if len(list.Items) == 0 {
		if _, err := c.Alive(ctx); err != nil {
			return nil, err
		}
	}
	c.stats(ctx, list.Items)
	return list.Items, nil
}

// Alive returns the token the hub accepts, a new one where a password login
// minted the one it holds, or ErrToken for a token the hub refuses. This is
// the one call that reports the token at all: PocketBase reads a list rule
// as a filter, so a refused token lists an empty page under 200.
func (c *Client) Alive(ctx context.Context) (string, error) {
	var out struct {
		Token string `json:"token"`
	}
	status, err := c.send(ctx, http.MethodPost, authPath, &out)
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return "", ErrToken
	}
	if err != nil {
		return "", err
	}
	return out.Token, nil
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
	return c.send(ctx, http.MethodGet, path, out)
}

// send calls a path of the hub with the token and decodes a success into
// out. It returns the status, and an error for anything but a success.
func (c *Client) send(ctx context.Context, method, path string, out any) (int, error) {
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.URL, "/")+path, nil)
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
