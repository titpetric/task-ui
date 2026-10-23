package detector

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"syscall"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/skills"
)

// stubAtkins points the detector at an atkins that prints the listing
// given, and puts the previous prober back when the test ends.
//
// The listing is the contract between the two programs, so a test of
// this package is a test of what it makes of one, not of the atkins
// that produced it.
func stubAtkins(t *testing.T, listing string) {
	t.Helper()

	command := filepath.Join(t.TempDir(), "atkins")
	script := "#!/bin/sh\n[ \"$1\" = warm ] && exit 0\ncat <<'LISTING'\n" + listing + "\nLISTING\n"
	assert.NoError(t, os.WriteFile(command, []byte(script), 0o755))

	// A parallel test that forks while the file is open for writing
	// hands the child the open file until it execs: wait for one clean
	// run.
	for range 100 {
		if err := exec.Command(command, "warm").Run(); !errors.Is(err, syscall.ETXTBSY) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Cleanup(UseProber(&skills.Prober{Command: command, TTL: time.Hour}))
}

// withPipeline returns a folder holding an atkins pipeline file. What
// is in it does not matter: atkins reads it, and the stub answers.
func withPipeline(t *testing.T, name string) string {
	t.Helper()

	dir := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("jobs:\n  build: echo\n"), 0o644))
	return dir
}

// infra is a listing with a job of every shape the folder cards show.
const infra = `[{
  "id": "",
  "kind": "pipeline",
  "path": "/w/infra/atkins.yml",
  "dir": "/w/infra",
  "active": true,
  "desc": "Infra",
  "help": "Deploys things.",
  "include": ["vars.yml"],
  "env": [".env", "secrets.env"],
  "cmds": [
    {"id":"deploy","name":"deploy","desc":"Deploy","cmd":"atkins deploy","active":true,
     "aliases":["ship"],"depends_on":["build"],"env":["deploy.env"],
     "steps":["run: echo deploy","Verify","task: notify","defer: rm -f lock"]},
    {"id":"custom","name":"custom","desc":"Compose","cmd":"atkins custom","active":true,
     "args":true,"exec":["docker","compose"],"steps":["run: echo first","exec: docker compose"]},
    {"id":"nodesc","name":"nodesc","cmd":"atkins nodesc","active":true,"steps":["run: echo"]},
    {"id":"hidden","name":"hidden","desc":"Hidden","cmd":"atkins hidden","active":true,"hidden":true,
     "steps":["run: echo hidden"]},
    {"id":"bare","name":"bare","cmd":"atkins bare","active":true}
  ]
},{
  "id": "git", "kind": "skill", "active": true, "desc": "Git",
  "cmds": [{"id":"git:pull","name":"pull","desc":"Pull","cmd":"atkins git:pull","active":true}]
}]`

func TestAtkinsDetector(t *testing.T) {
	stubAtkins(t, infra)
	dir := withPipeline(t, "atkins.yml")

	source := atkinsDetector{}.Detect(dir)
	assert.NotNil(t, source)
	assert.Equal(t, "", source.Error)

	// What the pipeline says about itself, as atkins reports it.
	assert.Equal(t, AtkinsKind, source.Kind)
	assert.Equal(t, "atkins.yml", source.File)
	assert.Equal(t, []string{"atkins", "-f", "atkins.yml"}, source.Command)
	assert.Equal(t, "Infra", source.Name)
	assert.Equal(t, "Deploys things.", source.Description)
	assert.Equal(t, []string{"vars.yml"}, source.Includes)

	// The pipeline's dotenv files, then each job's, once each.
	assert.Equal(t, []string{".env", "secrets.env", "deploy.env"}, source.EnvFiles)

	// The listing's order is atkins' own, and the skill section beside
	// the pipeline is not this folder's jobs.
	var names []string
	for _, target := range source.Targets {
		names = append(names, target.Name)
	}
	assert.Equal(t, []string{"deploy", "custom", "nodesc"}, names)

	deploy := source.Target("deploy")
	assert.Equal(t, "Deploy", deploy.Description)
	assert.Equal(t, []string{"ship"}, deploy.Aliases)
	assert.Equal(t, []string{"build"}, deploy.DependsOn)
	assert.Equal(t, []string{"deploy.env"}, deploy.EnvFiles)
	assert.Equal(t, []string{"run: echo deploy", "Verify", "task: notify", "defer: rm -f lock"}, deploy.Steps)

	// A job that takes arguments carries the command they follow.
	custom := source.Target("custom")
	assert.True(t, custom.TakesArgs)
	assert.Equal(t, []string{"docker", "compose"}, custom.Exec)
	assert.False(t, deploy.TakesArgs)

	// A job with no desc takes its first step, the way `atkins -l`
	// shows one.
	assert.Equal(t, "run: echo", source.Target("nodesc").Description)

	// A job the file hides, and one with nothing to say about itself,
	// are not offered.
	assert.Nil(t, source.Target("hidden"))
	assert.Nil(t, source.Target("bare"))
}

// A folder with no pipeline file is not an atkins folder, and the
// detector answers without running anything.
func TestAtkinsDetectorWithoutAPipeline(t *testing.T) {
	stubAtkins(t, infra)
	assert.Nil(t, atkinsDetector{}.Detect(t.TempDir()))
}

// A pipeline atkins cannot read still answers, so the folder lists what
// is wrong rather than nothing.
func TestAtkinsDetectorOnABrokenPipeline(t *testing.T) {
	stubAtkins(t, "not json")
	dir := withPipeline(t, ".atkins.yml")

	source := atkinsDetector{}.Detect(dir)
	assert.NotNil(t, source)
	assert.Equal(t, ".atkins.yml", source.File)
	assert.Contains(t, source.Error, "could not read this pipeline")
	assert.Len(t, source.Targets, 0)
}

// Every name atkins discovers is a pipeline, in its order.
func TestAtkinsDetectorFileNames(t *testing.T) {
	stubAtkins(t, infra)

	for _, name := range atkinsNames {
		source := atkinsDetector{}.Detect(withPipeline(t, name))
		assert.NotNil(t, source, name)
		assert.Equal(t, name, source.File)
	}
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

// FromFile reads the file a configuration names, whatever it is called,
// through the same listing.
func TestFromFileReadsAnAtkinsPipeline(t *testing.T) {
	stubAtkins(t, infra)
	dir := withPipeline(t, "ci.yml")

	source := FromFile(filepath.Join(dir, "ci.yml"))
	assert.NotNil(t, source)
	assert.Equal(t, AtkinsKind, source.Kind)
	assert.Equal(t, "ci.yml", source.File)
	assert.Equal(t, "Infra", source.Name)
	assert.True(t, slices.Contains([]string{"deploy"}, source.Targets[0].Name))
}
