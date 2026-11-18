package runner

import (
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/folder"
)

// envValue returns the last value of name in an environment.
func envValue(env []string, name string) string {
	value := ""
	for _, kv := range env {
		key, val, ok := cut(kv)
		if ok && key == name {
			value = val
		}
	}
	return value
}

// cut splits an environment entry at the first equals sign.
func cut(entry string) (string, string, bool) {
	for i := 0; i < len(entry); i++ {
		if entry[i] == '=' {
			return entry[:i], entry[i+1:], true
		}
	}
	return "", "", false
}

// A command of a prompt run names its Claude session and the directory that holds it.
func TestManager_commandSession(t *testing.T) {
	t.Parallel()
	m := New(Options{ClaudeDir: "/data/claude"})
	id := "01arz3ndektsv4rrffq69g5fav"
	session := &Session{
		Run:    &model.Run{ID: "r1", UserName: "ana", ClaudeSession: id, TermCols: 120, TermRows: 30},
		Folder: &folder.Folder{ID: "internal", Dir: t.TempDir()},
		Job:    &folder.Job{ID: model.PromptJob, Command: []string{"true"}},
	}
	env := m.claudeEnv(session)
	assert.Equal(t, id, envValue(env, "TASK_UI_CLAUDE_SESSION"))
	assert.Equal(t, "/data/claude/ana_"+id, envValue(env, "TASK_UI_CLAUDE_CONFIG"))
	assert.Equal(t, "", envValue(env, "TASK_UI_CLAUDE_RESUME"))
}

// A resumed command asks Claude Code to continue the session.
func TestManager_commandResume(t *testing.T) {
	t.Parallel()
	m := New(Options{ClaudeDir: "/data/claude"})
	session := &Session{
		Run:    &model.Run{ID: "r1", UserName: "ana", ClaudeSession: "01arz3ndektsv4rrffq69g5fav", TermCols: 120, TermRows: 30},
		Folder: &folder.Folder{ID: "internal", Dir: t.TempDir()},
		Job:    &folder.Job{ID: model.PromptJob, Command: []string{"true"}},
		resume: true,
	}
	assert.Equal(t, "1", envValue(m.claudeEnv(session), "TASK_UI_CLAUDE_RESUME"))
}

// A run without a session gets neither variable.
func TestManager_commandNoSession(t *testing.T) {
	t.Parallel()
	m := New(Options{ClaudeDir: "/data/claude"})
	session := &Session{
		Run:    &model.Run{ID: "r1", UserName: "ana", TermCols: 120, TermRows: 30},
		Folder: &folder.Folder{ID: "internal", Dir: t.TempDir()},
		Job:    &folder.Job{ID: "lint", Command: []string{"true"}},
	}
	env := m.claudeEnv(session)
	assert.Equal(t, "", envValue(env, "TASK_UI_CLAUDE_SESSION"))
	assert.Equal(t, "", envValue(env, "TASK_UI_CLAUDE_CONFIG"))
}
