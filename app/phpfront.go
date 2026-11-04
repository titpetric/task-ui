package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/hostinfo"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/phpfront"
	"github.com/titpetric/task-ui/phpscript"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/retention"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/skills"
	"github.com/titpetric/task-ui/sshprobe"
	"github.com/titpetric/task-ui/tooling"
	"github.com/titpetric/task-ui/workflow"
)

// FrontBase is where the ported front end answers, which is the root: it is
// the front end the instance serves. Every link it writes carries this, so a
// page writes {base}/settings and the empty string makes that /settings.
const FrontBase = ""

// GoBase is where the Go pages still answer, the front end being replaced.
// They are kept to be read against the ported ones on a live instance, which
// is what scripts/compare.sh fetches, and they go in the cutover commit
// together with this line.
//
// A page under it draws its links without a prefix, because the Go view layer
// has no notion of one: a link followed from a page here leaves for the
// ported page at the root. That is what keeping them is for — reading one
// page and diffing it — and not browsing the old front end.
const GoBase = "/legacy"

// FrontHost answers phpfront.Host off the instance: the catalogue, the live
// runner, the checkouts, the permission file and the event bus. Every row the
// front end reads it reads itself.
//
// Rules hands over the titles in file order and the path lists the file
// decides, so the decision in code/Task/Access.php is the only part of
// permission/ that is ported.
type FrontHost struct {
	Catalogue    folder.Catalogue
	Runner       *runner.Manager
	Git          *git.Git
	Events       *event.Bus
	Recordings   *recording.Store
	Integrations *integration.Store

	// Secrets is the store over the instance vault, asked what a variable
	// holds: the rows are PHP's and the key is not, so the value of one is
	// read here and nowhere else.
	Secrets *secret.Store

	// Workflows is the workflow runner, asked which workflows are running:
	// the rows say running of a run whose process is gone, and the two front
	// ends read the one memory so they mark the same workflows.
	Workflows *workflow.Runner

	// Artifacts is the tree the runs wrote, which the front end lists a
	// directory at a time, and Disk the meter both front ends measure it
	// with, so neither walks the tree a second time.
	Artifacts *artifactstore.Store
	Disk      *diskusage.Meter

	// Hosts is the Beszel hub, SSH who task-ui signs in to each host as,
	// and Facts what a host last told over that login.
	Hosts *beszel.Inventory
	SSH   *sshprobe.Prober
	Facts *hostinfo.Reader

	// Tools is what atkins has recorded for the tree, the prober both front
	// ends ask, so running atkins is not a cost the second one pays again.
	Tools *tooling.Prober

	// Keystore is the key folder, which answers the public half of a key no
	// row holds, and Config what the server runs with, the one the Instance
	// page of either front end draws.
	Keystore *keystore.Local
	Config   *view.Instance

	// Tracing is the tracer's options and Plugins the modules registered at
	// startup: the Telemetry and the Sidebar settings.
	Tracing *Telemetry
	Plugins *phpscript.Manager

	// Cleaner measures what the retention rules apply to, and Storage what
	// the instance keeps on disk, both the meters the Go pages read, so a
	// page drawn twice walks the disk once.
	Cleaner *retention.Cleaner
	Storage *diskusage.Inventory

	// Root is the workspace root, which a checkout's status is relative to.
	Root string

	// Now is the instance's clock, the one TASK_UI_NOW pins. An hours rule
	// is read at that moment, because view.Page.InHours reads it at the
	// page's own, and a binding on the process clock would answer open where
	// the Go page beside it answers closed. Nil is time.Now.
	Now func() time.Time

	// Dispatcher starts the runs a page asks for and Runs answers which run
	// one turned into, the pair the plugins already use: a dispatched run is
	// an ordinary prompt either way, so there is one of each on the
	// instance.
	//
	// FrontRoot is where phpfront.Extract put the front end, which the
	// action file a dispatch names is resolved inside. It is the same path
	// as phpfront.Options.Root; it is here because the dispatcher takes the
	// root it reads in, and the host is what knows which tree this is.
	Dispatcher *Dispatcher
	Runs       *ModuleRuns
	FrontRoot  string

	// Compose is the compose project task-ui itself runs in: compose up and
	// down of that project ask before they run, as they stop the instance
	// the page was drawn by.
	Compose string

	// subs keeps what Folder.Sub read of the directories a Jobs table
	// probes, so one page view reads each directory once.
	subs frontSubs
}

// frontFrameRows is the least rows a frame on a page shows, which is what
// frontend/page.go trims to.
const frontFrameRows = 8

// Whoami is the name the Guard resolved for this request.
func (h *FrontHost) Whoami(ctx context.Context) string {
	user := access.UserFrom(ctx)
	if user == nil {
		return ""
	}
	return user.Name
}

// Folders is the catalogue: each folder and the jobs it declares.
func (h *FrontHost) Folders(context.Context) []phpfront.Folder {
	if h.Catalogue == nil {
		return nil
	}
	found := h.Catalogue.List()
	out := make([]phpfront.Folder, 0, len(found))
	for _, f := range found {
		jobs := make([]phpfront.Job, 0, len(f.Jobs))
		for _, job := range f.Jobs {
			jobs = append(jobs, frontJob(job))
		}
		out = append(out, phpfront.Folder{
			ID:      f.ID,
			Title:   f.Title,
			Dir:     f.Dir,
			Allow:   f.Allow,
			Sources: frontSources(f.Sources),
			Error:   f.Error,
			Errors:  frontFindings(f.Errors),
			Jobs:    jobs,
		})
	}
	return out
}

// frontSources is the files a folder's conventions were detected in, which
// its row in the sidebar names itself by.
func frontSources(sources []*detector.Source) []string {
	out := make([]string, 0, len(sources))
	for _, source := range sources {
		if source != nil {
			out = append(out, source.File)
		}
	}
	return out
}

// frontFindings is what atkins' linter said about a folder's pipelines.
func frontFindings(found []skills.Finding) []phpfront.Finding {
	out := make([]phpfront.Finding, 0, len(found))
	for _, finding := range found {
		out = append(out, phpfront.Finding{
			Path: finding.Path, Job: finding.Job, Issue: finding.Issue, Detail: finding.Detail,
		})
	}
	return out
}

// Projects is the folder IDs of the directories right inside a named
// workspace's src: a checkout with no pipeline file is in no catalogue and
// the sidebar lists it anyway.
func (h *FrontHost) Projects(_ context.Context, workspace string) []string {
	if h.Root == "" || workspace == "" {
		return nil
	}
	entries, err := os.ReadDir(layout.SrcDir(h.Root, workspace))
	if err != nil {
		return nil
	}
	var ids []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			ids = append(ids, workspace+"/"+entry.Name())
		}
	}
	return ids
}

// FolderDir is the directory a folder ID names, the way frontend/side_dirty.go
// answers it: the one the catalogue holds, else where the layout puts the
// project among the named workspaces.
func (h *FrontHost) FolderDir(_ context.Context, id string, workspaces []string) string {
	if id == "" {
		return ""
	}
	if h.Catalogue != nil {
		if f := h.Catalogue.Find(id); f != nil && f.Dir != "" {
			return f.Dir
		}
	}
	if h.Root == "" {
		return ""
	}
	return layout.Dir(h.Root, workspaces, id)
}

// IsDir reports whether a directory inside a folder of the catalogue is
// there, which is what a pin of a directory is resolved against.
func (h *FrontHost) IsDir(_ context.Context, id, rel string) bool {
	if h.Catalogue == nil {
		return false
	}
	f := h.Catalogue.Find(id)
	if f == nil {
		return false
	}
	dir, ok := f.Path(rel)
	if !ok {
		return false
	}
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// Commit is the time of the last commit in a directory, through the cache
// both front ends share, and "" for a directory in no checkout or one whose
// checkout has no commit.
//
// It answers in the shape a timestamp column holds, which Task\Clock calls
// STORED, in UTC: the sidebar orders a workspace's projects by this and
// compares two of them as text, and a commit carries the committer's own zone,
// which no two projects need share.
func (h *FrontHost) Commit(_ context.Context, dir string) string {
	if h.Git == nil || dir == "" {
		return ""
	}
	at := h.Git.Commit(dir)
	if at.IsZero() {
		return ""
	}
	return at.UTC().Format(frontStamp)
}

// frontStamp is Task\Clock::STORED, the shape a timestamp column holds.
const frontStamp = "2006-01-02 15:04:05.999 -0700 MST"

// Job is the job a run names, resolved the way frontend/run.go resolves it:
// the folder of the catalogue, the directory inside it the run ran in, and
// the job that directory holds. Each step answers nothing where the thing is
// gone, and a job that is not there is one nobody may run.
func (h *FrontHost) Job(_ context.Context, id, dir, job string) phpfront.Job {
	if h.Catalogue == nil {
		return phpfront.Job{}
	}
	f := h.Catalogue.Find(id)
	if f == nil {
		return phpfront.Job{}
	}
	sub := f.Sub(dir)
	if sub == nil {
		return phpfront.Job{}
	}
	found := sub.Job(job)
	if found == nil {
		return phpfront.Job{}
	}
	out := frontFullJob(found)
	out.Found = true
	out.FolderAllow = sub.Allow
	// The button that shows the source is drawn for a target of a pipeline
	// file of the place itself, so a script, a skill and a job of a
	// directory inside are lines no page reads.
	if found.Kind != "script" && found.Kind != "skill" && found.Dir == "" {
		out.SourceLines = sub.Lines(found.Source)
	}
	for _, text := range found.DependsOn {
		dep := phpfront.Depends{Text: text}
		if first := sub.Job(text); first != nil {
			dep.Found, dep.ID, dep.Name, dep.Dir = true, first.ID, first.Name, first.Dir
		}
		out.Depends = append(out.Depends, dep)
	}
	return out
}

// frontFullJob is one job of the catalogue in the depth its own page and the
// Jobs table draw: what the button says, what a run starts, and what gates it.
// Folders answers the narrow frontJob instead, as the sidebar draws every job
// of every folder and names none of them.
func frontFullJob(job *folder.Job) phpfront.Job {
	if job == nil {
		return phpfront.Job{}
	}
	out := frontJob(job)
	out.Kind = job.Kind
	out.Source = job.Source
	out.Description = job.Description
	out.Summary = job.Summary
	out.Command = job.Command
	out.Steps = job.Steps
	out.Aliases = job.Aliases
	out.Interactive = job.Interactive
	out.TakesArgs = job.TakesArgs
	out.ArgsCommand = job.ArgsCommand()
	out.Key = job.Key()
	return out
}

// frontJob is one job of the catalogue as the front end's gate reads it. The
// two kinds come from the folder package, so the rule that names a terminal
// or a host job is stated once, where the jobs are made.
func frontJob(job *folder.Job) phpfront.Job {
	if job == nil {
		return phpfront.Job{}
	}
	return phpfront.Job{
		ID:    job.ID,
		Name:  job.Name,
		Dir:   job.Dir,
		Shell: job.IsShell(),
		Host:  job.IsHost(),
		Allow: job.Allow,
	}
}

// Running is the runs the runner holds, by the key their job is listed under.
func (h *FrontHost) Running(context.Context) []phpfront.Running {
	if h.Runner == nil {
		return nil
	}
	var out []phpfront.Running
	for _, session := range h.Runner.List() {
		run := session.Snapshot()
		if run.Finished() {
			continue
		}
		out = append(out, phpfront.Running{Key: run.Key(), ID: run.ID, State: string(run.State)})
	}
	return out
}

// Flows is the workflow run in progress of every workflow that has one, as
// the workflow runner holds them.
func (h *FrontHost) Flows(context.Context) []phpfront.Flow {
	if h.Workflows == nil {
		return nil
	}
	held := h.Workflows.Running()
	out := make([]phpfront.Flow, 0, len(held))
	for id, run := range held {
		out = append(out, phpfront.Flow{WorkflowID: id, RunID: run})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].WorkflowID < out[j].WorkflowID })
	return out
}

// Dirty counts the changed and untracked files under a checkout directory.
func (h *FrontHost) Dirty(_ context.Context, dir string) int {
	if h.Git == nil || h.Root == "" || dir == "" {
		return 0
	}
	status := h.Git.Status(dir, h.Root)
	if status == nil {
		return 0
	}
	rel, err := filepath.Rel(filepath.Join(h.Root, filepath.FromSlash(status.Checkout)), dir)
	if err != nil {
		return 0
	}
	return status.Under(strings.TrimPrefix(filepath.ToSlash(rel), "."))
}

// Rules is config/permissions.yml, titles and path lists both.
func (h *FrontHost) Rules(context.Context) phpfront.Rules {
	file := config.Default()
	return phpfront.Rules{
		Rules:       frontRules(file.Rules),
		Sections:    frontRules(file.Sections),
		Permissions: frontRules(file.Permissions),
		Paths:       file.Paths(),
		Checks:      file.Checks(),
		Extras:      frontExtras(file),
	}
}

// Publish sends an event on the stream every live label reads.
func (h *FrontHost) Publish(_ context.Context, topic string, payload map[string]any) error {
	if h.Events == nil {
		return nil
	}
	name, _ := payload["event"].(string)
	h.Events.Publish(topic, name, payload)
	return nil
}

// Dispatch starts a run over a comment, through the dispatcher the plugins
// use, with the action file resolved inside the front end's own tree.
//
// An instance with no board starts nothing and says so: a page whose form
// posts into a dispatch draws the error rather than reporting a run that is
// not there.
func (h *FrontHost) Dispatch(ctx context.Context, comment, action, folder string) (string, error) {
	if h.Dispatcher == nil {
		return "", errors.New("this instance starts no runs")
	}
	return h.Dispatcher.Dispatch(ctx, h.FrontRoot, comment, action, folder)
}

// RunOf is the run a note holds, "" where the note has none yet or the
// instance keeps no notes.
func (h *FrontHost) RunOf(ctx context.Context, note string) string {
	if h.Runs == nil {
		return ""
	}
	return h.Runs.RunOf(ctx, note)
}

// Session is a run as the runner holds it. A run that waits says so and says
// nothing else, which is why the position replaces the message.
func (h *FrontHost) Session(_ context.Context, id string) phpfront.Session {
	found, ok := h.held(id)
	if !ok {
		return phpfront.Session{}
	}
	run := found.Snapshot()
	out := phpfront.Session{
		Live:     true,
		Attached: found.Attached(),
		Message:  found.Status().Message,
		State:    run.State,
		ExitCode: run.ExitCode,
		Title:    run.Title,
	}
	if run.FinishedAt != nil {
		out.FinishedAt = run.FinishedAt.String()
	}
	if at := h.Runner.Position(id); at > 0 {
		out.Message = fmt.Sprintf("waits for a free slot, number %d in line", at)
	}
	return out
}

// Frame is the last screen of a live run, as the SVG a page embeds.
func (h *FrontHost) Frame(_ context.Context, id string) string {
	found, ok := h.held(id)
	if !ok {
		return ""
	}
	return found.Frame().Trim(frontFrameRows).SVG(runner.DarkPalette)
}

// Screen is the last frame of a run that has ended, replayed off the recording
// the run wrote, which is what a pipeline draws under every step. The caller
// holds the row, so the path of the recording and the columns it was written
// at come over with the ID.
//
// A run still going answers from the live session, as Frame does, and so does
// a recording that will not replay: a frame is decoration and a page that
// cannot draw one still draws the step.
func (h *FrontHost) Screen(_ context.Context, id, castPath string, cols int) phpfront.Screen {
	found, live := h.held(id)
	fromSession := func() phpfront.Screen {
		if !live {
			return phpfront.Screen{}
		}
		drawn := found.Frame().Trim(frontFrameRows).SVG(runner.DarkPalette)
		return phpfront.Screen{SVG: drawn}
	}
	if live && !found.Status().Finished() || h.Recordings == nil || castPath == "" {
		return fromSession()
	}
	cast, err := h.Recordings.Read(&model.Run{CastPath: castPath})
	if errors.Is(err, recording.ErrExpired) {
		gone := runner.TextFrame(cols, frontFrameRows, recording.Expired).SVG(runner.DarkPalette)
		return phpfront.Screen{SVG: gone, Complete: true}
	}
	if err != nil {
		return fromSession()
	}
	defer cast.Close()
	replay, err := runner.ReplayCast(cast)
	if err != nil {
		return fromSession()
	}
	drawn := replay.Screen.Trim(frontFrameRows).SVG(runner.DarkPalette)
	return phpfront.Screen{SVG: drawn, Lines: replay.Lines(), Complete: replay.Complete()}
}

// Expired reports whether a run names a recording that is not on disk, which
// is what the retention sweep leaves behind and no column records.
func (h *FrontHost) Expired(_ context.Context, castPath, state string) bool {
	if h.Recordings == nil {
		return false
	}
	return h.Recordings.IsExpired(&model.Run{CastPath: castPath, State: state})
}

// Claude reports whether the reader has a token for Claude. The tokens are
// the integration store's secrets, which the front end never reads itself.
func (h *FrontHost) Claude(ctx context.Context) bool {
	user := access.UserFrom(ctx)
	if h.Integrations == nil || user == nil {
		return false
	}
	env, err := h.Integrations.Env(ctx, user.Name, user.EffectiveRole())
	if err != nil {
		return false
	}
	return env["CLAUDE_CODE_OAUTH_TOKEN"] != "" || env["ANTHROPIC_AUTH_TOKEN"] != ""
}

// Tree is one directory of the artifact tree, as the artifacts page lists it.
// A directory's size is the space the files under it take, which the listing
// measures the way the page's Space does and the entry itself does not carry.
func (h *FrontHost) Tree(_ context.Context, dir string) []phpfront.TreeItem {
	if h.Artifacts == nil {
		return nil
	}
	items := h.Artifacts.List(dir)
	out := make([]phpfront.TreeItem, 0, len(items))
	for _, item := range items {
		each := phpfront.TreeItem{
			Name:    item.Name,
			Dir:     item.Dir,
			Size:    item.Size,
			ModTime: item.ModTime.String(),
		}
		if item.Dir {
			if at, err := h.Artifacts.Path(path.Join(dir, item.Name)); err == nil {
				each.Size = diskusage.Measure(at).Files
			}
		}
		out = append(out, each)
	}
	return out
}

// Space is the space the artifact tree takes, from the meter the Go pages
// read, so the tree is walked once however a reader came to it.
func (h *FrontHost) Space(context.Context) phpfront.Space {
	if h.Artifacts == nil || h.Disk == nil {
		return phpfront.Space{}
	}
	usage := h.Disk.Measure(h.Artifacts.Dir)
	return phpfront.Space{
		Measured: true,
		Files:    usage.Files,
		History:  usage.History,
		Ignored:  usage.Ignored,
	}
}

// Tooling is what atkins has recorded for the tree, the scopes of its report
// with the project's first.
//
// It reports commands and not runs, so it says nothing about who ran what and
// needs no permission of its own beyond the page it sits on: a reader of the
// runs page may already see that the tree runs go test, because the runs are
// there.
//
// Empty without a prober, without a root, and on a tree nobody has run a job
// in through atkins, which is the usual answer.
func (h *FrontHost) Tooling(context.Context) []phpfront.ToolingScope {
	if h.Tools == nil || h.Root == "" {
		return nil
	}
	found := h.Tools.Probe(h.Root).Scopes()
	out := make([]phpfront.ToolingScope, 0, len(found))
	for _, scope := range found {
		out = append(out, frontToolingScope(scope))
	}
	return out
}

// frontToolingScope is one database of the report.
func frontToolingScope(scope *tooling.Scope) phpfront.ToolingScope {
	out := phpfront.ToolingScope{
		Name:       scope.Name,
		Repository: scope.Repository,
		Database:   scope.Database,
		Days:       scope.Days,
		From:       scope.From,
		To:         scope.To,
		Runs:       scope.Runs,
		TotalMs:    scope.TotalMs,
	}
	for _, command := range scope.Commands {
		each := phpfront.ToolingCommand{
			Command:   command.Command,
			Tool:      command.Tool,
			Runs:      command.Runs,
			TotalMs:   command.TotalMs,
			AverageMs: command.AverageMs,
			MaxMs:     command.MaxMs,
			Failures:  command.Failures,
		}
		if command.LastRunAt != nil {
			each.LastRunAt = command.LastRunAt.String()
		}
		out.Commands = append(out.Commands, each)
	}
	for _, row := range scope.Usage {
		out.Usage = append(out.Usage, phpfront.ToolingUsage{Date: row.Date, Tool: row.Tool, Runs: row.Runs, TotalMs: row.TotalMs})
	}
	for _, row := range scope.CommandUsage {
		out.CommandUsage = append(out.CommandUsage, phpfront.ToolingUsage{
			Date: row.Date, Command: row.Command, Tool: row.Tool, Runs: row.Runs, TotalMs: row.TotalMs,
		})
	}
	for _, sample := range scope.Samples {
		out.Samples = append(out.Samples, phpfront.ToolingSample{
			At:         sample.At.String(),
			Command:    sample.Command,
			Tool:       sample.Tool,
			Job:        sample.Job,
			DurationMs: sample.DurationMs,
			ExitCode:   sample.ExitCode,
			Failed:     sample.Failed,
		})
	}
	return out
}

// held is the session of a run, where the runner holds one.
func (h *FrontHost) held(id string) (*runner.Session, bool) {
	if h.Runner == nil || id == "" {
		return nil, false
	}
	return h.Runner.Get(id)
}

// frontRules is a titled list of the permission file.
func frontRules(entries []config.Entry) []phpfront.Rule {
	out := make([]phpfront.Rule, 0, len(entries))
	for _, entry := range entries {
		out = append(out, phpfront.Rule{Key: entry.Key, Title: entry.Title})
	}
	return out
}

// frontExtras is the rules a section added of its own, which a section's all
// does not stand for: keyed by the path a check names, titled as the rule
// titles itself. The matrix draws a row per one and composes the title with
// the section's, which it reads off Sections by the path's first segment.
func frontExtras(file *config.Permissions) []phpfront.Rule {
	out := make([]phpfront.Rule, 0, len(file.Extra))
	for _, extra := range file.Extra {
		out = append(out, phpfront.Rule{Key: extra.Path(), Title: extra.Title})
	}
	return out
}
