package model_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// The CLI kind is the empty string, which is what a row written before the
// column existed holds: a token nothing issued through a page reads as one
// of `task-ui token`, and the two kinds never collide.
func TestTokenKind(t *testing.T) {
	t.Parallel()
	var kind model.TokenKind = model.TokenCLI
	assert.Equal(t, "", kind)
	assert.Equal(t, "mcp", model.TokenMCP)
	assert.NotEqual(t, model.TokenCLI, model.TokenMCP)
}
