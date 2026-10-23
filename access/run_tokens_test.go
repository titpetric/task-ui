package access_test

import (
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
)

// A run token signs in as its user until its revoke, and marks the
// request as a run's with its thread; tokens of other runs stay.
func TestRunTokens(t *testing.T) {
	t.Parallel()
	tokens := access.NewRunTokens()
	token, revoke, err := tokens.Issue("u1", "n1")
	assert.NoError(t, err)
	other, _, err := tokens.Issue("u2", "")
	assert.NoError(t, err)
	id, ok := tokens.User(token)
	assert.True(t, ok)
	assert.Equal(t, "u1", id)
	assert.Equal(t, "n1", tokens.Thread(token))
	assert.Equal(t, "", tokens.Thread(other))
	revoke()
	_, ok = tokens.User(token)
	assert.False(t, ok)
	assert.Equal(t, "", tokens.Thread(token))
	id, _ = tokens.User(other)
	assert.Equal(t, "u2", id)
	_, ok = (*access.RunTokens)(nil).User(other)
	assert.False(t, ok)
}
