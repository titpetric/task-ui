package argv_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/argv"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestSplit(t *testing.T) {
	t.Parallel()
	cases := map[string][]string{
		"":                        nil,
		"   ":                     nil,
		"ps -a":                   {"ps", "-a"},
		"  logs\t-f  web ":        {"logs", "-f", "web"},
		`exec web sh -c 'ls; id'`: {"exec", "web", "sh", "-c", "ls; id"},
		`"; id"`:                  {"; id"},
		`a\ b c`:                  {"a b", "c"},
		`"say \"hi\" \n"`:         {`say "hi" \n`},
		`'$HOME' * $(id) |`:       {"$HOME", "*", "$(id)", "|"},
		`''`:                      {""},
		`a"b"'c'`:                 {"abc"},
	}
	for line, want := range cases {
		got, err := argv.Split(line)
		assert.NoError(t, err, line)
		assert.Equal(t, want, got, line)
	}
	for _, line := range []string{`'open`, `"open`, `trailing\`} {
		_, err := argv.Split(line)
		assert.ErrorIs(t, err, argv.ErrUnterminated, line)
	}
	_, err := argv.Split(strings.Repeat("a ", argv.MaxArgs+1))
	assert.Error(t, err)
	_, err = argv.Split("'a\nb'")
	assert.Error(t, err)
}

func TestCheck(t *testing.T) {
	t.Parallel()
	assert.NoError(t, argv.Check(nil))
	assert.NoError(t, argv.Check([]string{"ps", "-a"}))
	assert.Error(t, argv.Check(make([]string, argv.MaxArgs+1)))
	assert.Error(t, argv.Check([]string{strings.Repeat("x", argv.MaxBytes+1)}))
	assert.NoError(t, argv.Check([]string{strings.Repeat("x", argv.MaxBytes)}))
	assert.Error(t, argv.Check([]string{"a\x00b"}))
	assert.Error(t, argv.Check([]string{"a\rb"}))
}
