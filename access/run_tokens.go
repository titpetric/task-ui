package access

import (
	"context"
	"sync"
)

// RunTokens are the tokens task-ui mints for a Claude run it starts, so
// Claude reaches task-ui's MCP server as the run's user: minted when
// the run starts, revoked when it ends, held in memory alone, never in
// the user's tokens and never shown. A run token signs in as its user,
// with the user's own permissions, whatever the user's MCP access.
type RunTokens struct {
	mu    sync.Mutex
	users map[string]string // token hash -> user ID
}

// NewRunTokens returns an empty set of run tokens.
func NewRunTokens() *RunTokens {
	return &RunTokens{users: map[string]string{}}
}

// Issue mints a token for the user and returns it with the revoke that
// ends it.
func (t *RunTokens) Issue(userID string) (string, func(), error) {
	token, hash, err := NewToken()
	if err != nil {
		return "", nil, err
	}
	t.mu.Lock()
	t.users[hash] = userID
	t.mu.Unlock()
	revoke := func() {
		t.mu.Lock()
		delete(t.users, hash)
		t.mu.Unlock()
	}
	return token, revoke, nil
}

// User returns the ID of the user a live run token signs in as.
func (t *RunTokens) User(token string) (string, bool) {
	if t == nil {
		return "", false
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	id, ok := t.users[HashToken(token)]
	return id, ok
}

type runKey struct{}

// FromRun reports whether the request's user signed in with a run
// token: Claude in a run task-ui started, which needs no MCP access of
// the user's.
func FromRun(ctx context.Context) bool {
	run, _ := ctx.Value(runKey{}).(bool)
	return run
}

// withRun marks a context as a run token's.
func withRun(ctx context.Context) context.Context {
	return context.WithValue(ctx, runKey{}, true)
}
