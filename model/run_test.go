package model_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestRun_Key(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "infra/make:build", (&model.Run{Folder: "infra", Job: "make:build"}).Key())
	assert.Equal(t, "malta/dockerfiles/build", (&model.Run{Folder: "malta", Dir: "dockerfiles", Job: "build"}).Key())
}

func TestRunKey(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "infra/plan", model.RunKey("infra", "", "plan"))
	assert.Equal(t, "infra/images/plan", model.RunKey("infra", "images", "plan"))
}

func TestRun_ArgList(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []string{"ps", "; id"}, (&model.Run{Args: `["ps","; id"]`}).ArgList())
	assert.Empty(t, (&model.Run{}).ArgList())
}

func TestRun_Finished(t *testing.T) {
	t.Parallel()
	assert.False(t, (*model.Run)(nil).Finished())
	assert.False(t, (&model.Run{State: model.StateRunning}).Finished())
	assert.True(t, (&model.Run{State: model.StateOK}).Finished())
}

func TestRun_Elapsed(t *testing.T) {
	t.Parallel()
	start := time.Now().Add(-time.Minute)
	end := start.Add(10 * time.Second)
	assert.Equal(t, time.Duration(0), (*model.Run)(nil).Elapsed())
	assert.Equal(t, time.Duration(0), (&model.Run{}).Elapsed())
	assert.Equal(t, 10*time.Second, (&model.Run{StartedAt: &start, FinishedAt: &end}).Elapsed())
	assert.GreaterOrEqual(t, (&model.Run{StartedAt: &start}).Elapsed(), time.Minute)
}

// Where names the folder, and the directory of a job of a directory.
func TestRun_Where(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "malta", model.Run{Folder: "malta", Job: "build"}.Where())
	assert.Equal(t, "malta/dockerfiles", model.Run{Folder: "malta", Dir: "dockerfiles", Job: "build"}.Where())
}

func TestRun_IsPrompt(t *testing.T) {
	t.Parallel()
	assert.True(t, (&model.Run{Folder: "malta", Dir: "dockerfiles", Job: model.PromptJob}).IsPrompt())
	assert.True(t, (&model.Run{Folder: "internal/malta", Job: "claude-prompt"}).IsPrompt(), "an older prompt's run")
	assert.False(t, (&model.Run{Folder: "malta", Job: "claude-prompt"}).IsPrompt())
	assert.False(t, (&model.Run{Folder: "malta", Job: "@shell"}).IsPrompt())
	assert.False(t, (*model.Run)(nil).IsPrompt())
}

func TestRun_IsNote(t *testing.T) {
	t.Parallel()
	assert.True(t, (&model.Run{Folder: model.Default, Job: model.NoteJob}).IsNote())
	assert.True(t, (&model.Run{Folder: model.Internal, Job: "claude-note"}).IsNote(), "an older note's run")
	assert.False(t, (&model.Run{Folder: "malta", Job: "claude-note"}).IsNote())
	assert.False(t, (&model.Run{Folder: "malta", Job: model.PromptJob}).IsNote())
	assert.False(t, (*model.Run)(nil).IsNote())
}

func TestRun_IsClaude(t *testing.T) {
	t.Parallel()
	for _, run := range []*model.Run{
		{Folder: "malta", Job: model.PromptJob},
		{Folder: model.Default, Job: model.NoteJob},
		{Folder: "malta", Dir: "dockerfiles", Job: model.ClaudeJob},
		{Folder: model.Internal, Job: "claude"},
	} {
		assert.True(t, run.IsClaude(), run.Key())
	}
	assert.False(t, (&model.Run{Folder: "malta", Job: "@shell"}).IsClaude())
	assert.False(t, (*model.Run)(nil).IsClaude())
}
