package forgejo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
)

// Sync brings the workspaces and projects under the root in line with
// the owners of repositories the token sees, the organisations of its
// user and the users of the instance alike: a workspace per owner, the
// repository named like it as the home, every other repository a
// project, and a push hook on each the token administers. A user with
// no readable repository gets no workspace. It adds and updates, and
// never removes: it reports a project whose repository Forgejo lacks
// and leaves it alone, as it does a home or a checkout that is not what
// Forgejo says. With DryRun it reports what it would do and changes nothing;
// the report's shape is in the package documentation, and Color paints it.
type Sync struct {
	Client     *Client
	Git        Git
	Workspaces Workspaces
	Projects   Projects
	// Root is the directory that holds the workspaces.
	Root string
	// HookURL is where Forgejo posts a push, <URL>/api/webhooks/forgejo;
	// empty installs no hook, and the report says why.
	HookURL string
	// Secret signs the hooks' bodies.
	Secret string
	// HTTPS clones over https.
	HTTPS  bool
	DryRun bool
	// Color paints the report, see Paint; the caller decides, Colored says
	// when a writer wants it.
	Color bool

	w     io.Writer
	paint Paint
	count map[string]int
}

// Run syncs every organisation, then every user, and writes the report to w, a block per owner and
// a footer with the counts; it returns an error only when the listing of the organisations fails.
func (s *Sync) Run(ctx context.Context, w io.Writer) error {
	s.w, s.paint, s.count = w, Paint(s.Color), map[string]int{}
	orgs, err := s.Client.Orgs(ctx)
	if err != nil {
		return fmt.Errorf("listing the organisations: %w", err)
	}
	if len(orgs) == 0 {
		s.line("%s", s.paint.Dim("the token's user is in no organisation"))
	}
	synced := map[string]map[string]bool{}
	for _, org := range orgs {
		s.sync(ctx, org.Owner(), synced)
	}
	users := s.users(ctx)
	for _, user := range users {
		s.sync(ctx, user.Owner(), synced)
	}
	s.orphans(ctx, synced)
	if s.HookURL == "" {
		s.line("%s", s.paint.Dim("hooks skipped: the instance has no public URL (TASK_UI_URL)"))
	}
	s.line("")
	s.line("%s", s.footer(len(orgs), len(users)))
	return nil
}

// footer sums the report up the way terraform does: Plan for a dry run, Applied for a sync, the
// adds by kind, then the updates, unchanged, skipped, left alone and failed rows, the last three
// kinds when there are any, over the owners.
func (s *Sync) footer(orgs, users int) string {
	head, to := "Applied:", ""
	if s.DryRun {
		head, to = "Plan:", "to "
	}
	parts := []string{
		fmt.Sprintf("%d workspaces %s", s.count["workspace"], s.verb(to, "create")),
		fmt.Sprintf("%d homes %s", s.count["home"], s.verb(to, "clone")),
		fmt.Sprintf("%d projects %s", s.count["project"], s.verb(to, "clone")),
		fmt.Sprintf("%d hooks %s", s.count["hook"], s.verb(to, "install")),
	}
	if n := s.count["update"]; n > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", n, s.verb(to, "update")))
	}
	parts = append(parts, fmt.Sprintf("%d unchanged", s.count["unchanged"]), fmt.Sprintf("%d skipped", s.count["skipped"]))
	if n := s.count["alone"]; n > 0 {
		parts = append(parts, fmt.Sprintf("%d left alone", n))
	}
	if n := s.count["failed"]; n > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", n))
	}
	return fmt.Sprintf("%s %s, over %d organisations and %d users.", s.paint.Bold(head), strings.Join(parts, ", "), orgs, users)
}

// verb returns "to create" for a plan and "created" for a sync.
func (s *Sync) verb(to, verb string) string {
	if to != "" {
		return to + verb
	}
	return s.did(verb)
}

// users lists the users of the instance; a token without read:user
// lists none, and the report says so, as it does for any other refusal.
func (s *Sync) users(ctx context.Context) []User {
	users, err := s.Client.Users(ctx)
	switch {
	case Status(err) == http.StatusForbidden:
		s.line("%s", s.paint.Dim("users skipped: the token lacks read:user"))
	case err != nil:
		s.line("%s", s.paint.Dim(fmt.Sprintf("users skipped: listing them: %v", err)))
	}
	return users
}

// sync syncs one owner, and skips a name that is no folder name and a
// user without a readable repository; an organisation keeps its
// workspace even empty.
func (s *Sync) sync(ctx context.Context, owner Owner, synced map[string]map[string]bool) {
	ws := strings.ToLower(owner.Name)
	s.header(owner)
	if !folder.ValidID(ws) || strings.Contains(ws, "/") || layout.Reserved(ws) {
		s.skip(owner.Name, "%q is no folder name", ws)
		return
	}
	repos, err := s.repos(ctx, owner)
	if err != nil {
		s.skip(owner.Name, "listing its repositories: %v", err)
		return
	}
	if owner.Kind == OwnerUser && len(repos) == 0 {
		s.skip(owner.Name, "no repositories")
		return
	}
	s.owner(ctx, owner, repos, synced)
}

// header opens an owner's block, "# org Malta" or "# user black".
func (s *Sync) header(owner Owner) {
	kind, name := "org", s.paint.Cyan(owner.Name)
	if owner.Kind == OwnerUser {
		kind, name = "user", s.paint.Magenta(owner.Name)
	}
	s.line("%s %s", s.paint.Bold("# "+kind), name)
}

// repos lists the repositories of an owner, by its kind.
func (s *Sync) repos(ctx context.Context, owner Owner) ([]Repo, error) {
	if owner.Kind == OwnerUser {
		return s.Client.UserRepos(ctx, owner.Name)
	}
	return s.Client.Repos(ctx, owner.Name)
}

// owner makes the workspace of an owner from its repositories, and
// records their names in synced, by workspace, archived ones too.
func (s *Sync) owner(ctx context.Context, owner Owner, repos []Repo, synced map[string]map[string]bool) {
	ws := strings.ToLower(owner.Name)
	row, ok := s.workspace(ctx, owner, ws)
	if !ok {
		return
	}
	names := map[string]bool{}
	synced[ws] = names
	var home *Repo
	for i := range repos {
		names[strings.ToLower(repos[i].Name)] = true
		if strings.EqualFold(repos[i].Name, owner.Name) {
			home = &repos[i]
		}
	}
	var hooked []Repo
	if home == nil {
		s.same("home", ws, "no %s/%s repository: %s", owner.Name, owner.Name, s.homeIs(row))
	} else {
		s.home(ctx, ws, row, *home)
		hooked = append(hooked, *home)
	}
	for _, repo := range repos {
		if home != nil && repo.Name == home.Name {
			continue
		}
		if s.project(ctx, ws, repo) {
			hooked = append(hooked, repo)
		}
	}
	if s.HookURL == "" {
		return
	}
	for _, repo := range hooked {
		s.hook(ctx, repo)
	}
}

// homeIs says what the home of a workspace without a home repository
// on Forgejo is: empty, or the checkout its row names.
func (s *Sync) homeIs(row *model.Workspace) string {
	if row == nil || row.HomeURL == "" {
		return "home is empty"
	}
	return "the home stays " + row.HomeURL
}

// workspace returns the row of an owner's workspace, created with its
// home and src when there is none, and whether to go on with the owner:
// a dry run goes on without a row; a name a project at the root holds,
// or a row the storage refuses, stops it.
func (s *Sync) workspace(ctx context.Context, owner Owner, ws string) (*model.Workspace, bool) {
	row, err := s.Workspaces.GetByFolder(ctx, ws)
	if err == nil {
		s.same("workspace", ws, "ok")
		return row, true
	}
	if _, err := s.Projects.GetByFolder(ctx, ws); err == nil {
		s.skip(owner.Name, "a project named %s is at the root", ws)
		return nil, false
	}
	s.count["workspace"]++
	s.add("workspace", ws, "%s", s.did("create"))
	if s.DryRun {
		return nil, true
	}
	if err := layout.Ensure(s.Root, ws); err != nil {
		s.fail("workspace", ws, err)
		return nil, false
	}
	row = &model.Workspace{Folder: ws, Title: owner.Title(), Description: owner.Description}
	if err := s.Workspaces.Create(ctx, row); err != nil {
		s.fail("workspace", ws, err)
		return nil, false
	}
	return row, true
}

// home makes the repository named like the owner the workspace's home: cloned into an empty home/, pulled from a home/
// that is its checkout, left alone and reported for any other home/,
// and the row set to its clone URL and default branch.
func (s *Sync) home(ctx context.Context, ws string, row *model.Workspace, repo Repo) {
	url, dir := repo.URL(s.HTTPS), layout.HomeDir(s.Root, ws)
	switch {
	case row != nil && row.HomeURL != "" && Key(row.HomeURL) != Key(url):
		s.alone("home", repo.FullName, "the home is a checkout of %s, left alone: replace it on the workspace's settings page", row.HomeURL)
		return
	case s.Git.IsRepo(dir):
		s.count["update"]++
		s.change("home", repo.FullName, "%s", s.did("pull"))
		if !s.DryRun {
			if err := s.Git.Pull(ctx, dir, io.Discard); err != nil {
				s.fail("home", repo.FullName, fmt.Errorf("pull failed: %w", err))
			}
		}
	case !empty(dir):
		s.alone("home", repo.FullName, "%s/home is not empty and no checkout: left alone", ws)
		return
	default:
		s.count["home"]++
		s.add("home", repo.FullName, "%s %s", s.did("clone"), repo.DefaultBranch)
		if !s.DryRun {
			_ = os.Remove(dir)
			if _, err := s.Git.Clone(ctx, url, dir, repo.DefaultBranch); err != nil {
				_ = os.MkdirAll(dir, 0o755)
				s.fail("home", repo.FullName, err)
				return
			}
		}
	}
	if s.DryRun || row == nil || (row.HomeURL == url && row.HomeBranch == repo.DefaultBranch) {
		return
	}
	row.HomeURL, row.HomeBranch = url, repo.DefaultBranch
	if err := s.Workspaces.Update(ctx, row); err != nil {
		s.fail("home", repo.FullName, err)
	}
}

// project makes a repository a project of the workspace: cloned into
// src/<name> with its row when there is neither, the row added when the
// checkout is there, the row's URL updated when it changed, and left
// alone otherwise; it skips an archived repository. It reports whether
// the repository is a project to hook.
func (s *Sync) project(ctx context.Context, ws string, repo Repo) bool {
	if repo.Archived {
		s.skip(repo.FullName, "archived")
		return false
	}
	if !folder.ValidID(repo.Name) || layout.Reserved(repo.Name) {
		s.skip(repo.FullName, "%q is no folder name", repo.Name)
		return false
	}
	id, url, dir := ws+"/"+repo.Name, repo.URL(s.HTTPS), layout.ProjectDir(s.Root, ws, repo.Name)
	row := &model.Project{Folder: id, SourceURL: url, Branch: repo.DefaultBranch}
	switch current, err := s.Projects.GetByFolder(ctx, id); {
	case err == nil && current.SourceURL == url:
		s.same("project", id, "exists")
		return true
	case err == nil:
		s.count["update"]++
		s.change("project", id, "%s url", s.did("update"))
		current.SourceURL = url
		row = current
	case s.Git.IsRepo(dir):
		s.add("project", id, "checkout exists, %s row", s.did("add"))
	case !empty(dir):
		s.alone("project", id, "src/%s exists and is no checkout: left alone", repo.Name)
		return false
	default:
		s.count["project"]++
		s.add("project", id, "%s %s", s.did("clone"), repo.DefaultBranch)
		if !s.DryRun {
			if _, err := s.Git.Clone(ctx, url, dir, repo.DefaultBranch); err != nil {
				s.fail("project", id, err)
				return false
			}
		}
	}
	if !s.DryRun {
		if err := s.Projects.Upsert(ctx, row); err != nil {
			s.fail("project", id, err)
		}
	}
	return true
}

// hook installs the push hook on a repository when it has none, and
// updates one that posts to the URL but differs; it skips a repository
// the token does not administer, as its hooks are out of reach.
func (s *Sync) hook(ctx context.Context, repo Repo) {
	if !repo.Permissions.Admin {
		s.skip("hook "+repo.FullName, "the token has no admin access")
		return
	}
	hooks, err := s.Client.Hooks(ctx, repo.Owner(), repo.Name)
	if err != nil {
		s.fail("hook", repo.FullName, fmt.Errorf("listing the hooks: %w", err))
		return
	}
	want := PushHook(s.HookURL, s.Secret)
	for _, hook := range hooks {
		if hook.Config["url"] != s.HookURL {
			continue
		}
		if hook.Current(s.HookURL, s.Secret) {
			s.same("hook", repo.FullName, "ok")
			return
		}
		s.count["update"]++
		s.change("hook", repo.FullName, "%s", s.did("update"))
		if !s.DryRun {
			if _, err := s.Client.UpdateHook(ctx, repo.Owner(), repo.Name, hook.ID, want); err != nil {
				s.fail("hook", repo.FullName, err)
			}
		}
		return
	}
	s.count["hook"]++
	s.add("hook", repo.FullName, "%s", s.did("install"))
	if !s.DryRun {
		if _, err := s.Client.CreateHook(ctx, repo.Owner(), repo.Name, want); err != nil {
			s.fail("hook", repo.FullName, err)
		}
	}
}

// orphans reports the projects of the synced workspaces whose
// repository is not on Forgejo any more.
func (s *Sync) orphans(ctx context.Context, synced map[string]map[string]bool) {
	projects, err := s.Projects.List(ctx)
	if err != nil {
		s.count["failed"]++
		s.line("%s listing the projects: %v", s.paint.Red("!"), err)
		return
	}
	for _, project := range projects {
		ws, name, _ := strings.Cut(project.Folder, "/")
		names, ok := synced[ws]
		if !ok || name == "" || strings.Contains(name, "/") || names[strings.ToLower(name)] {
			continue
		}
		s.count["alone"]++
		s.line("%s orphan %s: not on Forgejo, left alone", s.paint.Red("!"), s.paint.Blue(project.Folder))
	}
}

// did returns the past tense of an action, or what a dry run would do.
func (s *Sync) did(verb string) string {
	if s.DryRun {
		return "would " + verb
	}
	if strings.HasSuffix(verb, "e") {
		return verb + "d"
	}
	return verb + "ed"
}

// add writes a row sync creates, "+" in green.
func (s *Sync) add(kind, name, format string, args ...any) {
	s.row(s.paint.Green("+"), kind, s.name(kind, name), format, args...)
}

// change writes a row sync brings current, "~" in yellow.
func (s *Sync) change(kind, name, format string, args ...any) {
	s.row(s.paint.Yellow("~"), kind, s.name(kind, name), format, args...)
}

// same writes a row sync finds in place, dim, and counts it.
func (s *Sync) same(kind, name, format string, args ...any) {
	s.count["unchanged"]++
	s.row("", kind, name, format, args...)
}

// skip writes a row sync leaves out, dim with the reason, and counts it.
func (s *Sync) skip(name, format string, args ...any) {
	s.count["skipped"]++
	s.row("", "skipped", name, format, args...)
}

// alone writes a row sync leaves alone where Forgejo says otherwise, "!" in red, and counts it.
func (s *Sync) alone(kind, name, format string, args ...any) {
	s.count["alone"]++
	s.row(s.paint.Red("!"), kind, s.name(kind, name), format, args...)
}

// fail writes a row sync could not do, "!" in red, and counts it.
func (s *Sync) fail(kind, name string, err error) {
	s.count["failed"]++
	s.row(s.paint.Red("!"), kind, s.name(kind, name), "%v", err)
}

// row writes one row of an owner's block, two spaces in: the symbol or a
// space in its place, the kind, the name and the text; a row with no
// symbol is dim as a whole.
func (s *Sync) row(symbol, kind, name, format string, args ...any) {
	text := fmt.Sprintf("%s %s: %s", kind, name, fmt.Sprintf(format, args...))
	if symbol == "" {
		s.line("    %s", s.paint.Dim(text))
		return
	}
	s.line("  %s %s", symbol, text)
}

// name colours a name by its kind: a workspace and its home cyan, a project blue, a hook plain.
func (s *Sync) name(kind, name string) string {
	switch kind {
	case "workspace", "home":
		return s.paint.Cyan(name)
	case "project":
		return s.paint.Blue(name)
	}
	return name
}

// line writes one line of the report.
func (s *Sync) line(format string, args ...any) {
	fmt.Fprintf(s.w, format+"\n", args...)
}

// empty reports a directory with nothing in it, or no directory.
func empty(dir string) bool {
	entries, err := os.ReadDir(dir)
	return err != nil || len(entries) == 0
}
