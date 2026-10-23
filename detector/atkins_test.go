package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/skills"
)

// withPipeline returns a folder holding an atkins pipeline file. What
// is in it does not matter: atkins reads it, and the stub answers.
func withPipeline(t *testing.T, name string) string {
	t.Helper()

	dir := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("jobs:\n  build: echo\n"), 0o644))
	return dir
}

func TestAtkinsDetector_Kind(t *testing.T) {
	assert.Equal(t, "atkins", atkinsDetector{}.Kind())
	assert.Equal(t, AtkinsKind, atkinsDetector{}.Kind())
}

func TestPipelineFile(t *testing.T) {
	assert.Equal(t, "", pipelineFile(t.TempDir()))

	dir := withPipeline(t, "atkins.yaml")
	assert.Equal(t, "atkins.yaml", pipelineFile(dir))

	// A folder holding several answers with the one atkins reads first.
	assert.NoError(t, os.WriteFile(filepath.Join(dir, ".atkins.yml"), []byte("jobs:\n"), 0o644))
	assert.Equal(t, ".atkins.yml", pipelineFile(dir))

	// A directory by that name is not a pipeline.
	bare := t.TempDir()
	assert.NoError(t, os.Mkdir(filepath.Join(bare, "atkins.yml"), 0o755))
	assert.Equal(t, "", pipelineFile(bare))
}

func TestPipelineSection(t *testing.T) {
	assert.Nil(t, pipelineSection(nil))
	assert.Nil(t, pipelineSection([]skills.Section{{Kind: skills.KindSkill, ID: "git"}}))

	sections := []skills.Section{
		{Kind: skills.KindSkill, ID: "git"},
		{Kind: skills.KindPipeline, Desc: "Infra"},
	}
	section := pipelineSection(sections)
	assert.NotNil(t, section)
	assert.Equal(t, "Infra", section.Desc)
}

func TestAtkinsTarget(t *testing.T) {
	// A job the file hides is not offered, whatever it says about
	// itself.
	assert.Nil(t, atkinsTarget(skills.Command{Name: "x", Desc: "Hidden", Hidden: true}))
	// A job with neither a desc nor a step says nothing about itself.
	assert.Nil(t, atkinsTarget(skills.Command{Name: "x"}))

	target := atkinsTarget(skills.Command{
		Name: "build", Desc: " Build ", Interactive: true,
		Steps: []string{"go build ./..."},
	})
	assert.NotNil(t, target)
	assert.Equal(t, "build", target.Name)
	assert.Equal(t, "Build", target.Description)
	assert.True(t, target.Interactive)
	assert.Equal(t, []string{"go build ./..."}, target.Steps)
}

func TestUseProber(t *testing.T) {
	first := &skills.Prober{Command: "first"}
	second := &skills.Prober{Command: "second"}

	restore := UseProber(first)
	assert.Equal(t, "first", prober.Bin())

	// Each call puts back the one it replaced, so nested wiring unwinds
	// in the order it was made.
	inner := UseProber(second)
	assert.Equal(t, "second", prober.Bin())
	inner()
	assert.Equal(t, "first", prober.Bin())
	restore()
}
