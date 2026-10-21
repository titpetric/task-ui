package folder_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/skills/skillstest"
)

// Skills lists the skills atkins applies in a folder's directory, each
// job @skill:job of the folder, run as atkins skill:job in the directory
// of the file the skill matched: the checkout of git, the directory of
// compose. Skills lists neither a job whose own condition fails nor an
// inactive skill.
func TestFolder_Skills(t *testing.T) {
	t.Parallel()
	f := checkout(t)
	probe, _ := skillstest.Prober(t)
	f.Probe = probe
	f.Allow = []string{"ops"}
	app := filepath.Join(f.Dir, "images", "app")
	assert.NoError(t, os.MkdirAll(app, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(app, "compose.yml"), []byte("name: Shop\nservices:\n  a:\n    build: .\n"), 0o644))
	plain := filepath.Join(f.Dir, "images", "plain")
	assert.NoError(t, os.MkdirAll(plain, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(plain, "compose.yml"), []byte("services:\n  a:\n    image: x\n"), 0o644))

	list := f.Skills()
	assert.Len(t, list, 1)
	assert.Equal(t, "git", list[0].ID)
	assert.Equal(t, "", list[0].Dir)
	assert.Equal(t, ".git", list[0].When.Matched)

	sub := f.Sub("images/app")
	list = sub.Skills()
	assert.Len(t, list, 2)
	git, compose := list[0], list[1]
	assert.Equal(t, "", git.Dir, "git runs in the checkout, the folder's top")
	assert.Equal(t, "images/app", compose.Dir)
	assert.Equal(t, "compose.yml", compose.When.Matched)
	assert.Equal(t, "shop", compose.Project)
	var ids []string
	for _, job := range compose.Jobs {
		ids = append(ids, job.ID)
	}
	assert.Equal(t, []string{"@compose:build", "@compose:custom", "@compose:down", "@compose:up"}, ids)

	job := sub.Job("@git:status")
	assert.NotNil(t, job)
	assert.Equal(t, "status", job.Name)
	assert.Equal(t, "skill", job.Kind)
	assert.Equal(t, "", job.Dir)
	assert.Equal(t, "acme/api/@git:status", job.Key())
	assert.Equal(t, []string{probe.Command, "git:status"}, job.Command)
	assert.Equal(t, "Show the branch", job.Description)

	up := sub.Job("@compose:up")
	assert.Equal(t, "images/app", up.Dir)
	assert.Equal(t, "acme/api/images/app/@compose:up", up.Key())
	assert.True(t, sub.Job("@compose:custom").TakesArgs)
	assert.Equal(t, []string{"docker", "compose"}, sub.Job("@compose:custom").Exec)
	assert.False(t, up.TakesArgs)

	plainSub := f.Sub("images/plain")
	assert.Nil(t, plainSub.Job("@compose:build"), "the file builds nothing")
	assert.Equal(t, "plain", plainSub.Skills()[1].Project)
	assert.NotNil(t, plainSub.Job("@compose:up"))
	assert.Nil(t, sub.Job("@go:test"), "an inactive skill")
	assert.Nil(t, sub.Job("@nope:x"))
	assert.Nil(t, f.Job("@compose:up"), "no compose file at the top")

	// Without a probe there are no skills.
	bare := checkout(t)
	assert.Nil(t, bare.Skills())
	assert.Nil(t, bare.Job("@git:status"))
}

// A skill matched above the folder runs in the folder's own directory:
// atkins finds it again from there.
func TestFolder_SkillsAbove(t *testing.T) {
	t.Parallel()
	probe, _ := skillstest.Prober(t)
	root, err := filepath.EvalSymlinks(t.TempDir())
	assert.NoError(t, err)
	assert.NoError(t, os.MkdirAll(filepath.Join(root, ".git"), 0o755))
	dir := filepath.Join(root, "apps", "shop")
	assert.NoError(t, os.MkdirAll(filepath.Join(dir, "web"), 0o755))
	f := &folder.Folder{ID: "shop", Dir: dir, Probe: probe}
	assert.Equal(t, "", f.Job("@git:status").Dir)
	assert.Equal(t, "web", f.Sub("web").Job("@git:status").Dir)
}

// SkillJob finds a skill's job by its atkins ID, without the @.
func TestFolder_SkillJob(t *testing.T) {
	t.Parallel()
	f := checkout(t)
	probe, _ := skillstest.Prober(t)
	f.Probe = probe
	job := f.SkillJob("git:pull")
	assert.NotNil(t, job)
	assert.Equal(t, "@git:pull", job.ID)
	assert.Equal(t, job, f.Job("@git:pull"))
	assert.Nil(t, f.SkillJob("@git:pull"))
	assert.Nil(t, f.SkillJob("git:nope"))
}
