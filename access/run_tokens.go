package access

import (
	"context"
	"sync"
)

// RunTokens are the tokens task-ui mints for a Claude run it starts, so
// Claude reaches task-ui's MCP server as the run's user: minted when
// the run starts, revoked when it ends, held in memory alone, never in
// the user's tokens and never shown. A run token signs in as its user,
// with the user's own permissions, whatever the user's MCP access, and
// names the thread the run answers, so what it starts reports there.
type RunTokens struct {
	mu     sync.Mutex
	tokens map[string]runToken // token hash -> its run
}

// runToken is the run a token signs in: its user and its thread.
type runToken struct {
	user, thread string
}

// NewRunTokens returns an empty set of run tokens.
func NewRunTokens() *RunTokens {
	return &RunTokens{tokens: map[string]runToken{}}
}

// Issue mints a token for the user of a run that answers the thread, a
// note's ID, and returns it with the revoke that ends it.
func (t *RunTokens) Issue(userID, thread string) (string, func(), error) {
	token, hash, err := NewToken()
	if err != nil {
		return "", nil, err
	}
	t.mu.Lock()
	t.tokens[hash] = runToken{user: userID, thread: thread}
	t.mu.Unlock()
	revoke := func() {
		t.mu.Lock()
		delete(t.tokens, hash)
		t.mu.Unlock()
	}
	return token, revoke, nil
}

// User returns the ID of the user a live run token signs in as.
func (t *RunTokens) User(token string) (string, bool) {
	run, ok := t.run(token)
	return run.user, ok
}

// Thread returns the ID of the note a live run token's run answers.
func (t *RunTokens) Thread(token string) string {
	run, _ := t.run(token)
	return run.thread
}

func (t *RunTokens) run(token string) (runToken, bool) {
	if t == nil {
		return runToken{}, false
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	run, ok := t.tokens[HashToken(token)]
	return run, ok
}

type runKey struct{}

// FromRun reports whether the request's user signed in with a run
// token: Claude in a run task-ui started, which needs no MCP access of
// the user's.
func FromRun(ctx context.Context) bool {
	_, ok := ctx.Value(runKey{}).(runToken)
	return ok
}

// RunThread returns the ID of the note the run whose token signed the
// request in answers, empty for a request of no run.
func RunThread(ctx context.Context) string {
	run, _ := ctx.Value(runKey{}).(runToken)
	return run.thread
}

// withRun marks a context as a run token's.
func withRun(ctx context.Context, run runToken) context.Context {
	return context.WithValue(ctx, runKey{}, run)
}
