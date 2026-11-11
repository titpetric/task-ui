//go:build integration

package claude

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/integration/git"
)

// fakeClaude writes a script that answers every prompt with one line.
func fakeClaude(t *testing.T, answer string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "claude")
	script := "#!/bin/sh\ncat >/dev/null\nprintf '%s\\n' " + answer + "\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestTriager_Triage(t *testing.T) {
	signedIn := func(context.Context) (map[string]string, error) {
		return map[string]string{"CLAUDE_CODE_OAUTH_TOKEN": "token"}, nil
	}
	t.Run("the case the model names", func(t *testing.T) {
		triager := &Triager{Command: fakeClaude(t, git.CaseDiverged), Timeout: 10 * time.Second, Env: signedIn}
		got, err := triager.Triage(context.Background(), "/root/project", "Not possible to fast-forward", "# branch.head main")
		if err != nil {
			t.Fatal(err)
		}
		if got != git.CaseDiverged {
			t.Errorf("the triage names %q, want %q", got, git.CaseDiverged)
		}
	})
	t.Run("an answer in no case", func(t *testing.T) {
		triager := &Triager{Command: fakeClaude(t, "'I had a look at the reflog'"), Timeout: 10 * time.Second, Env: signedIn}
		got, err := triager.Triage(context.Background(), "/root/project", "error: RPC failed", "")
		if err != nil {
			t.Fatal(err)
		}
		if got != "" {
			t.Errorf("the triage names %q, want no case", got)
		}
	})
	t.Run("no model", func(t *testing.T) {
		triager := &Triager{Command: fakeClaude(t, git.CaseDirty), Timeout: 10 * time.Second}
		if _, err := triager.Triage(context.Background(), "/root/project", "", ""); !errors.Is(err, ErrNoToken) {
			t.Errorf("a triage without the integrations answered %v, want %v", err, ErrNoToken)
		}
	})
	t.Run("no sign in", func(t *testing.T) {
		triager := &Triager{Command: fakeClaude(t, git.CaseDirty), Timeout: 10 * time.Second, Env: func(context.Context) (map[string]string, error) {
			return map[string]string{}, nil
		}}
		if _, err := triager.Triage(context.Background(), "/root/project", "", ""); !errors.Is(err, ErrNoToken) {
			t.Errorf("a triage with no token answered %v, want %v", err, ErrNoToken)
		}
	})
}
