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

	// A folder holding two pipeline files answers with the one atkins reads first.
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
	// in reverse order.
	inner := UseProber(second)
	assert.Equal(t, "second", prober.Bin())
	inner()
	assert.Equal(t, "first", prober.Bin())
	restore()
}

// A listing atkins' linter refuses becomes a source with the error and
// the findings, each file relative to the folder, and no targets.
func TestFromListing_Lint(t *testing.T) {
	dir := t.TempDir()
	listing := skills.Listing{
		Error: "pipeline 'Broken' has errors",
		Lint: []skills.Finding{
			{Pipeline: "Broken", Path: filepath.Join(dir, "atkins.yml"), Job: "build", Detail: "step references task 'go:fmt'"},
			{Pipeline: "Broken", Detail: "no file"},
			{Pipeline: "go", Path: filepath.Join(filepath.Dir(dir), ".atkins", "skills", "go.yml"), Job: "fmt", Detail: "skill"},
		},
	}

	source := fromListing(dir, "atkins.yml", listing)
	assert.Equal(t, "atkins.yml: pipeline 'Broken' has errors", source.Error)
	assert.Len(t, source.Targets, 0)
	assert.Len(t, source.Errors, 3)
	assert.Equal(t, "atkins.yml", source.Errors[0].Path)
	assert.Equal(t, "build", source.Errors[0].Job)
	assert.Equal(t, "atkins.yml", source.Errors[1].Path)
	assert.Equal(t, "../.atkins/skills/go.yml", source.Errors[2].Path)
}

// An error without findings is one finding; a broken listing that lists
// its sections too lists their jobs beside the findings.
func TestFromListing_ErrorAndJobs(t *testing.T) {
	listing := skills.Listing{
		Error:    "pipeline 'Half' has errors",
		Sections: []skills.Section{{Kind: skills.KindPipeline, Desc: "Half", Cmds: []skills.Command{{Name: "test", Desc: "Test"}}}},
	}
	source := fromListing(t.TempDir(), "atkins.yml", listing)
	assert.Len(t, source.Errors, 1)
	assert.Equal(t, "pipeline 'Half' has errors", source.Errors[0].Detail)
	assert.Equal(t, "atkins.yml", source.Errors[0].Path)
	assert.Len(t, source.Targets, 1)
	assert.Equal(t, "Half", source.Name)

	// An empty listing reads as unreadable; a clean one has
	// no errors.
	assert.Contains(t, fromListing(t.TempDir(), "atkins.yml", skills.Listing{}).Error, "could not read this pipeline")
	clean := fromListing(t.TempDir(), "atkins.yml", skills.Listing{Sections: listing.Sections})
	assert.Equal(t, "", clean.Error)
	assert.Nil(t, clean.Errors)
}
