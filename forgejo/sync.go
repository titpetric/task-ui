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
// Forgejo says. With DryRun it reports what it would do and changes nothing.
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
	// HTTPS clones over https in place of ssh.
	HTTPS  bool
	DryRun bool

	w     io.Writer
	count map[string]int
}

// Run syncs every organisation, then every user, and writes the report to w, one line per action
// and a summary; it returns an error only when the listing of the organisations fails.
func (s *Sync) Run(ctx context.Context, w io.Writer) error {
	s.w, s.count = w, map[string]int{}
	orgs, err := s.Client.Orgs(ctx)
	if err != nil {
		return fmt.Errorf("listing the organisations: %w", err)
	}
	if len(orgs) == 0 {
		s.line("the token's user is in no organisation")
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
		s.line("hooks skipped: the instance has no public URL (TASK_UI_URL)")
	}
	s.line("summary: %d organisations, %d users, %d workspaces created, %d homes cloned, %d projects cloned, %d hooks installed, %d skipped, %d not on Forgejo",
		len(orgs), len(users), s.count["workspace"], s.count["home"], s.count["project"], s.count["hook"], s.count["skipped"], s.count["orphan"])
	if s.DryRun {
		s.line("a dry run: nothing changed")
	}
	return nil
}

// users lists the users of the instance; a token without read:user
// lists none, and the report says so, as it does for any other refusal.
func (s *Sync) users(ctx context.Context) []User {
	users, err := s.Client.Users(ctx)
	switch {
	case Status(err) == http.StatusForbidden:
		s.line("users skipped: the token lacks read:user")
	case err != nil:
		s.line("users skipped: listing them: %v", err)
	}
	return users
}

// sync syncs one owner, and skips a name that is no folder name and a
// user without a readable repository; an organisation keeps its
// workspace even empty.
func (s *Sync) sync(ctx context.Context, owner Owner, synced map[string]map[string]bool) {
	ws := strings.ToLower(owner.Name)
	if !folder.ValidID(ws) || strings.Contains(ws, "/") || layout.Reserved(ws) {
		s.skip("%s: %q is no folder name", owner.Name, ws)
		return
	}
	repos, err := s.repos(ctx, owner)
	if err != nil {
		s.skip("%s: listing its repositories: %v", owner.Name, err)
		return
	}
	if owner.Kind == OwnerUser && len(repos) == 0 {
		s.skip("%s: no repositories", owner.Name)
		return
	}
	s.owner(ctx, owner, repos, synced)
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
		s.line("home %s: no %s/%s repository: %s", ws, owner.Name, owner.Name, s.homeIs(row))
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
		s.line("workspace %s: ok", ws)
		return row, true
	}
	if _, err := s.Projects.GetByFolder(ctx, ws); err == nil {
		s.skip("%s: a project named %s is at the root", owner.Name, ws)
		return nil, false
	}
	s.count["workspace"]++
	s.line("workspace %s: %s", ws, s.did("create"))
	if s.DryRun {
		return nil, true
	}
	if err := layout.Ensure(s.Root, ws); err != nil {
		s.line("workspace %s: %v", ws, err)
		return nil, false
	}
	row = &model.Workspace{Folder: ws, Title: owner.Title(), Description: owner.Description}
	if err := s.Workspaces.Create(ctx, row); err != nil {
		s.line("workspace %s: %v", ws, err)
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
		s.line("home %s: the home is a checkout of %s: replace it on the workspace's settings page", repo.FullName, row.HomeURL)
		return
	case s.Git.IsRepo(dir):
		s.line("home %s: %s", repo.FullName, s.did("pull"))
		if !s.DryRun {
			if err := s.Git.Pull(ctx, dir, io.Discard); err != nil {
				s.line("home %s: pull failed: %v", repo.FullName, err)
			}
		}
	case !empty(dir):
		s.line("home %s: %s/home is not empty and no checkout: left alone", repo.FullName, ws)
		return
	default:
		s.count["home"]++
		s.line("home %s: %s %s", repo.FullName, s.did("clone"), repo.DefaultBranch)
		if !s.DryRun {
			_ = os.Remove(dir)
			if _, err := s.Git.Clone(ctx, url, dir, repo.DefaultBranch); err != nil {
				_ = os.MkdirAll(dir, 0o755)
				s.line("home %s: %v", repo.FullName, err)
				return
			}
		}
	}
	if s.DryRun || row == nil || (row.HomeURL == url && row.HomeBranch == repo.DefaultBranch) {
		return
	}
	row.HomeURL, row.HomeBranch = url, repo.DefaultBranch
	if err := s.Workspaces.Update(ctx, row); err != nil {
		s.line("home %s: %v", repo.FullName, err)
	}
}

// project makes a repository a project of the workspace: cloned into
// src/<name> with its row when there is neither, the row added when the
// checkout is there, the row's URL updated when it changed, and left
// alone otherwise; it skips an archived repository. It reports whether
// the repository is a project to hook.
func (s *Sync) project(ctx context.Context, ws string, repo Repo) bool {
	if repo.Archived {
		s.skip("%s: archived", repo.FullName)
		return false
	}
	if !folder.ValidID(repo.Name) || layout.Reserved(repo.Name) {
		s.skip("%s: %q is no folder name", repo.FullName, repo.Name)
		return false
	}
	id, url, dir := ws+"/"+repo.Name, repo.URL(s.HTTPS), layout.ProjectDir(s.Root, ws, repo.Name)
	row := &model.Project{Folder: id, SourceURL: url, Branch: repo.DefaultBranch}
	switch current, err := s.Projects.GetByFolder(ctx, id); {
	case err == nil && current.SourceURL == url:
		s.line("project %s: exists", id)
		return true
	case err == nil:
		s.line("project %s: %s url", id, s.did("update"))
		current.SourceURL = url
		row = current
	case s.Git.IsRepo(dir):
		s.line("project %s: checkout exists, %s row", id, s.did("add"))
	case !empty(dir):
		s.line("project %s: src/%s exists and is no checkout: left alone", id, repo.Name)
		return false
	default:
		s.count["project"]++
		s.line("project %s: %s %s", id, s.did("clone"), repo.DefaultBranch)
		if !s.DryRun {
			if _, err := s.Git.Clone(ctx, url, dir, repo.DefaultBranch); err != nil {
				s.line("project %s: %v", id, err)
				return false
			}
		}
	}
	if !s.DryRun {
		if err := s.Projects.Upsert(ctx, row); err != nil {
			s.line("project %s: %v", id, err)
		}
	}
	return true
}

// hook installs the push hook on a repository when it has none, and
// updates one that posts to the URL but differs; it skips a repository
// the token does not administer, as its hooks are out of reach.
func (s *Sync) hook(ctx context.Context, repo Repo) {
	if !repo.Permissions.Admin {
		s.count["skipped"]++
		s.line("hook %s: skipped, the token has no admin access", repo.FullName)
		return
	}
	hooks, err := s.Client.Hooks(ctx, repo.Owner(), repo.Name)
	if err != nil {
		s.line("hook %s: listing the hooks: %v", repo.FullName, err)
		return
	}
	want := PushHook(s.HookURL, s.Secret)
	for _, hook := range hooks {
		if hook.Config["url"] != s.HookURL {
			continue
		}
		if hook.Current(s.HookURL, s.Secret) {
			s.line("hook %s: ok", repo.FullName)
			return
		}
		s.line("hook %s: %s", repo.FullName, s.did("update"))
		if !s.DryRun {
			if _, err := s.Client.UpdateHook(ctx, repo.Owner(), repo.Name, hook.ID, want); err != nil {
				s.line("hook %s: %v", repo.FullName, err)
			}
		}
		return
	}
	s.count["hook"]++
	s.line("hook %s: %s", repo.FullName, s.did("install"))
	if !s.DryRun {
		if _, err := s.Client.CreateHook(ctx, repo.Owner(), repo.Name, want); err != nil {
			s.line("hook %s: %v", repo.FullName, err)
		}
	}
}

// orphans reports the projects of the synced workspaces whose
// repository is not on Forgejo any more.
func (s *Sync) orphans(ctx context.Context, synced map[string]map[string]bool) {
	projects, err := s.Projects.List(ctx)
	if err != nil {
		s.line("listing the projects: %v", err)
		return
	}
	for _, project := range projects {
		ws, name, _ := strings.Cut(project.Folder, "/")
		names, ok := synced[ws]
		if !ok || name == "" || strings.Contains(name, "/") || names[strings.ToLower(name)] {
			continue
		}
		s.count["orphan"]++
		s.line("not on Forgejo: %s", project.Folder)
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

// skip reports something sync leaves out and counts it.
func (s *Sync) skip(format string, args ...any) {
	s.count["skipped"]++
	s.line("skipped "+format, args...)
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
