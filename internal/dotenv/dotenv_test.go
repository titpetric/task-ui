package dotenv_test

import (
	"testing"

	"github.com/titpetric/task-ui/dotenv"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestFormat(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "A=1\nB=\"has space\"\n", dotenv.Format(map[string]string{"B": "has space", "A": "1"}))
}

func TestQuote(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{
		"plain":    "plain",
		"":         `""`,
		"a b":      `"a b"`,
		`say "hi"`: `'say "hi"'`,
		`it's "x"`: `it's "x"`,
		"a#b":      `"a#b"`,
	} {
		assert.Equal(t, want, dotenv.Quote(in), in)
	}
}

func TestKeys(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []string{"a", "b"}, dotenv.Keys(map[string]string{"b": "", "a": ""}))
}
