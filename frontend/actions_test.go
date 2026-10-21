package frontend_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/frontend"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/skills/skillstest"
)

// actionsBar cuts the actions bar out of a page.
func actionsBar(body string) string {
	_, rest, ok := strings.Cut(body, `<div class="actionsbar"`)
	if !ok {
		return ""
	}
	bar, _, _ := strings.Cut(rest, "\n</div>")
	return bar
}

// A folder page and a directory browsed inside a folder carry the
// actions bar: the git line of the checkout, a chip per skill atkins
// applies there, each job a Run of the folder in the directory the job
// runs in, and a chip of the place's own atkins jobs. A viewer gets the
// job pages instead; compose up and down of task-ui's own project ask
// first.
func TestFolder_Actions(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	probe, _ := skillstest.Prober(t)
	f.infra.Probe = probe
	out, err := exec.Command("git", "-c", "init.defaultBranch=main", "init", "-q", f.infra.Dir).CombinedOutput()
	assert.NoError(t, err, string(out))
	app := filepath.Join(f.infra.Dir, "app")
	assert.NoError(t, os.MkdirAll(app, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(app, "compose.yml"), []byte("services:\n  a:\n    build: .\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(app, "atkins.yml"), []byte("jobs:\n  default:\n    steps:\n      - run: echo hi\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(f.infra.Dir, "new.txt"), []byte("x\n"), 0o644))

	top := actionsBar(f.page(t, asOps, "/f/_/infra"))
	assert.Contains(t, top, `<span class="glyph" aria-hidden="true">⎇</span> main`)
	assert.Contains(t, top, `●2`, "new.txt and app/")
	assert.Contains(t, top, `· infra`)
	assert.Contains(t, top, `<span class="chip-name">git</span>`)
	assert.Contains(t, top, `data-folder="infra" data-job="@git:pull" data-title="infra / pull"`)
	assert.NotContains(t, top, "compose")

	bar := actionsBar(f.page(t, asOps, "/f/_/infra/app"))
	assert.Contains(t, bar, `⎇</span> main`)
	assert.Contains(t, bar, `data-folder="infra" data-job="@git:status" data-title="infra / status"`, "git runs in the checkout")
	assert.Contains(t, bar, `data-folder="infra" data-dir="app" data-job="@compose:build" data-title="infra/app / build"`)
	assert.Contains(t, bar, `data-dir="app" data-job="@compose:up"`)
	// compose custom takes args: its Run opens a field naming what they
	// follow first.
	assert.Contains(t, bar, `data-job="@compose:custom" data-title="infra/app / custom" data-args="docker compose" title=`)
	assert.Contains(t, bar, `>custom...</button>`)
	assert.NotContains(t, bar, `data-job="@compose:up" data-title="infra/app / up" data-args`)
	assert.Contains(t, bar, `when compose.yml here, matched compose.yml; runs in app/`)
	assert.Contains(t, bar, `<span class="chip-name">atkins</span>`)
	assert.Contains(t, f.page(t, asOps, "/f/_/infra/app"), "Its jobs run in infra/app/")
	deploy := filepath.Join(f.infra.Dir, "deploy")
	assert.NoError(t, os.MkdirAll(deploy, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(deploy, "compose.yml"), []byte("services:\n  a:\n    image: x\n"), 0o644))
	skillsOnly := f.page(t, asOps, "/f/_/infra/deploy")
	assert.Contains(t, skillsOnly, "the skills below run here")
	assert.NotContains(t, actionsBar(skillsOnly), "@compose:build")
	assert.Contains(t, bar, `data-folder="infra" data-dir="app" data-job="default"`)
	assert.NotContains(t, bar, "data-confirm")

	viewer := actionsBar(f.page(t, asViewer, "/f/_/infra/app"))
	assert.Contains(t, viewer, `<a class="chip-act" href="/f/_/infra/app/@compose:up"`)
	assert.Contains(t, viewer, `<a class="chip-act" href="/f/_/infra/@git:status"`)
	assert.NotContains(t, viewer, "data-run")
	page := f.page(t, asViewer, "/f/_/infra/app/@compose:up")
	assert.Contains(t, page, "infra/app / up")
	assert.Contains(t, page, "atkins compose:up")
	assert.Contains(t, f.page(t, asViewer, "/f/_/infra/@git:status"), "Show the branch")

	// The project task-ui runs in: up and down ask first.
	f.opts.ComposeProject = "app"
	h, err := frontend.New(f.opts)
	assert.NoError(t, err)
	f.server.Config.Handler = h.Routes()
	bar = actionsBar(f.page(t, asOps, "/f/_/infra/app"))
	assert.Contains(t, bar, `data-job="@compose:up" data-title="infra/app / up" data-confirm="Run compose up for app?`)
	assert.Contains(t, bar, `data-job="@compose:down" data-title="infra/app / down" data-confirm="Run compose down for app?`)
	assert.NotContains(t, bar, `data-job="@compose:build" data-title="infra/app / build" data-confirm`)

	// No checkout and no skills: no bar.
	assert.Equal(t, "", actionsBar(f.page(t, asOps, "/f/_/private")))
}

// An admin with a Claude token gets the prompt box on every folder page
// and on a directory browsed inside a folder, posting its place to
// /api/prompts; open on a workspace's page, folded elsewhere.
func TestFolder_Prompt(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	const block = `data-api="/api/prompts" data-attach-result`
	assert.NoError(t, os.MkdirAll(filepath.Join(f.infra.Dir, "images"), 0o755))

	// infra is a workspace; without a Claude token there is no prompt.
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "infra", Title: "Infra"}))
	assert.NotContains(t, f.page(t, asAdmin, "/f/infra"), block)

	// The token reaches everyone; the prompt is the admins'.
	f.withKey(t)
	assert.NoError(t, f.opts.Integrations.Save(ctx, integration.Claude, map[string]string{"token": "sk-1"}))
	admin := f.page(t, asAdmin, "/f/infra")
	assert.Contains(t, admin, `<details class="promptbox" open>`)
	assert.Contains(t, admin, block)
	assert.Contains(t, admin, `<input type="hidden" name="folder" value="infra">`)
	assert.Contains(t, admin, `<textarea name="prompt"`)
	assert.NotContains(t, f.page(t, asOps, "/f/infra"), block)
	assert.NotContains(t, f.page(t, asViewer, "/f/infra"), block)

	// A folder that is not a workspace takes one too, folded, and so
	// does a directory inside a folder.
	private := f.page(t, asAdmin, "/f/_/private")
	assert.Contains(t, private, `<details class="promptbox">`)
	assert.Contains(t, private, `<input type="hidden" name="folder" value="private">`)
	browse := f.page(t, asAdmin, "/f/infra/images")
	assert.Contains(t, browse, block)
	assert.Contains(t, browse, `<input type="hidden" name="dir" value="images">`)
	assert.Contains(t, browse, "Claude Code in infra/images/")
	assert.NotContains(t, f.page(t, asOps, "/f/infra/images"), block)
}

// Enter Shell sits in the actions bar of a folder and of a directory
// inside one for an admin, and for no one else.
func TestFolder_Shell(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	assert.NoError(t, os.MkdirAll(filepath.Join(f.infra.Dir, "images"), 0o755))
	const button = `data-job="@shell"`

	admin := f.page(t, asAdmin, "/f/_/infra")
	assert.Contains(t, admin, button)
	assert.Contains(t, admin, `data-title="infra / shell"`)
	browse := f.page(t, asAdmin, "/f/_/infra/images")
	assert.Contains(t, browse, `data-dir="images" data-job="@shell" data-title="infra/images / shell"`)
	for _, user := range []string{asOps, asViewer, asLimited} {
		assert.NotContains(t, f.page(t, user, "/f/_/infra"), button, user)
		assert.NotContains(t, f.page(t, user, "/f/_/infra/images"), button, user)
	}
}
