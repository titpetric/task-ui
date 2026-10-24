//go:build integration

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
	// leaves the child holding the open file until it execs: wait for
	// one clean run.
	for range 100 {
		if err := exec.Command(command, "warm").Run(); !errors.Is(err, syscall.ETXTBSY) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Cleanup(UseProber(&skills.Prober{Command: command, TTL: time.Hour}))
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
// is wrong.
func TestAtkinsDetectorOnABrokenPipeline(t *testing.T) {
	stubAtkins(t, "not json")
	dir := withPipeline(t, ".atkins.yml")

	source := atkinsDetector{}.Detect(dir)
	assert.NotNil(t, source)
	assert.Equal(t, ".atkins.yml", source.File)
	assert.Contains(t, source.Error, "could not read this pipeline")
	assert.Len(t, source.Targets, 0)
}

// A pipeline atkins' linter refuses answers its findings and exits 1;
// the source carries them, so the folder says what is wrong. The
// listing closes the stub's heredoc to exit 1 after it.
func TestAtkinsDetectorOnALintedPipeline(t *testing.T) {
	stubAtkins(t, `{"error":"pipeline 'Broken' has errors","lint":[{"pipeline":"Broken","job":"build","issue":"missing task reference","detail":"no go:fmt"}]}
LISTING
exit 1
cat <<'LISTING'`)
	dir := withPipeline(t, "atkins.yml")

	source := atkinsDetector{}.Detect(dir)
	assert.NotNil(t, source)
	assert.Equal(t, "atkins.yml: pipeline 'Broken' has errors", source.Error)
	assert.Len(t, source.Errors, 1)
	assert.Equal(t, "no go:fmt", source.Errors[0].Detail)
	assert.Equal(t, "atkins.yml", source.Errors[0].Path)
}

// The atkins on the PATH refuses a step that calls a job it lacks, and
// the source names the job and the file.
func TestAtkinsDetectorWithAtkins(t *testing.T) {
	command, err := exec.LookPath("atkins")
	if err != nil {
		t.Skip("no atkins on the PATH")
	}
	t.Cleanup(UseProber(&skills.Prober{Command: command, Timeout: 10 * time.Second}))
	dir := t.TempDir()
	pipeline := "name: Broken\njobs:\n  build:\n    desc: Build\n    steps:\n      - task: go:fmt\n"
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "atkins.yml"), []byte(pipeline), 0o644))

	source := atkinsDetector{}.Detect(dir)
	assert.NotNil(t, source)
	assert.Len(t, source.Targets, 0)
	assert.Len(t, source.Errors, 1)
	assert.Equal(t, "atkins.yml", source.Errors[0].Path)
	assert.Equal(t, "build", source.Errors[0].Job)
	assert.Contains(t, source.Errors[0].Detail, "go:fmt")
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

// FromFile reads the file a configuration names, whatever its name,
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
