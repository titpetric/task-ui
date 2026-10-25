package forgejo_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/forgejo"
	"github.com/titpetric/task-ui/internal/assert"
)

// Paint on wraps its text in an escape and the reset, off returns it
// plain, and neither touches an empty string.
func TestPaint(t *testing.T) {
	t.Parallel()
	on, off := forgejo.Paint(true), forgejo.Paint(false)
	for code, paint := range map[string]func(string) string{
		"1": on.Bold, "2": on.Dim, "31": on.Red, "32": on.Green, "33": on.Yellow, "34": on.Blue, "35": on.Magenta, "36": on.Cyan,
	} {
		assert.Equal(t, "\x1b["+code+"mx\x1b[0m", paint("x"))
		assert.Equal(t, "", paint(""))
	}
	for _, paint := range []func(string) string{off.Bold, off.Dim, off.Red, off.Green, off.Yellow, off.Blue, off.Magenta, off.Cyan} {
		assert.Equal(t, "x", paint("x"))
	}
}

// A buffer and a file want no colour, a run of task-ui does, and
// NO_COLOR turns it off everywhere.
func TestColored(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("TASK_UI_RUN_ID", "")
	assert.False(t, forgejo.Colored(&bytes.Buffer{}))
	f, err := os.Create(filepath.Join(t.TempDir(), "report"))
	assert.NoError(t, err)
	defer f.Close()
	assert.False(t, forgejo.Colored(f))

	t.Setenv("TASK_UI_RUN_ID", "r1")
	assert.True(t, forgejo.Colored(&bytes.Buffer{}))
	t.Setenv("NO_COLOR", "1")
	assert.False(t, forgejo.Colored(&bytes.Buffer{}))
}

// With Color the report carries ANSI escapes: the owner headers bold
// with the name in the colour of its kind, a green "+" before a
// workspace sync creates and cyan on its name, blue on a project, a red
// "!" on an orphan, and a skipped row dim as a whole; the footer's head
// is bold.
func TestSync_Color(t *testing.T) {
	t.Parallel()
	fake, workspaces, projects := instance(t)
	s, _ := newSync(t, fake, workspaces, projects, t.TempDir())
	s.DryRun, s.Color = true, true
	var out bytes.Buffer
	assert.NoError(t, s.Run(context.Background(), &out))
	report := out.String()
	for _, line := range []string{
		"\x1b[1m# org\x1b[0m \x1b[36mMalta\x1b[0m",
		"\x1b[1m# user\x1b[0m \x1b[35mblack\x1b[0m",
		"  \x1b[32m+\x1b[0m workspace \x1b[36mmalta\x1b[0m: would create",
		"  \x1b[32m+\x1b[0m home \x1b[36mMalta/malta\x1b[0m: would clone main",
		"  \x1b[32m+\x1b[0m project \x1b[34mmalta/apps\x1b[0m: would clone main",
		"  \x1b[32m+\x1b[0m hook Malta/apps: would install",
		"    \x1b[2mskipped Malta/old-thing: archived\x1b[0m",
		"    \x1b[2mworkspace acme: ok\x1b[0m",
		"\x1b[31m!\x1b[0m orphan \x1b[34macme/vhosts\x1b[0m: not on Forgejo, left alone",
	} {
		assert.Contains(t, report, line+"\n")
	}
	assert.Contains(t, report, "\x1b[1mPlan:\x1b[0m 3 workspaces to create, ")

	// A pull and a hook update are yellow "~" rows.
	fake, workspaces, projects = instance(t)
	root := t.TempDir()
	s, _ = newSync(t, fake, workspaces, projects, root)
	s.Color = true
	fake.hooks["Malta/deploy"] = []forgejo.Hook{{ID: 7, Config: map[string]string{"url": hookURL, "content_type": "json"}, Events: []string{"push"}, Active: false}}
	assert.NoError(t, os.MkdirAll(filepath.Join(root, "malta", "home", ".git"), 0o755))
	out.Reset()
	assert.NoError(t, s.Run(context.Background(), &out))
	assert.Contains(t, out.String(), "  \x1b[33m~\x1b[0m home \x1b[36mMalta/malta\x1b[0m: pulled\n")
	assert.Contains(t, out.String(), "  \x1b[33m~\x1b[0m hook Malta/deploy: updated\n")
	assert.Contains(t, out.String(), "\x1b[1mApplied:\x1b[0m 3 workspaces created, 1 homes cloned, 5 projects cloned, 5 hooks installed, 2 updated, ")
}
