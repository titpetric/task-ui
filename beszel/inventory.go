package beszel

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrOff reports a hub that is not set up: no URL or no token.
var ErrOff = errors.New("the Beszel integration is not set up")

// Inventory is the host inventory: the systems of the hub the Beszel
// integration names. It reads the settings on every call and keeps one
// client for as long as they stay the same.
type Inventory struct {
	// Settings returns the hub's URL and token.
	Settings func(ctx context.Context) (url, token string, err error)
	// TTL is how long an answer of the hub stays good.
	TTL time.Duration

	mu     sync.Mutex
	client *Client
}

// Systems returns the hub's systems, or ErrOff without a hub.
func (i *Inventory) Systems(ctx context.Context) ([]System, error) {
	url, token, err := i.Settings(ctx)
	if err != nil {
		return nil, err
	}
	if url == "" || token == "" {
		return nil, ErrOff
	}
	i.mu.Lock()
	if i.client == nil || i.client.URL != url || i.client.Token != token {
		i.client = &Client{URL: url, Token: token, TTL: i.TTL}
	}
	client := i.client
	i.mu.Unlock()
	return client.Systems(ctx)
}
