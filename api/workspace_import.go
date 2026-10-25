package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/workspace"
)

// importedProject is how the clone of one imported project went.
type importedProject struct {
	Folder string `json:"folder"`
	Cloned bool   `json:"cloned"`
	Output string `json:"output"`
}

// errImport reports a bundle whose content the import refuses, a 400.
var errImport = errors.New("the bundle")

// importWorkspace adds a workspace from a bundle an export wrote: a
// multipart form of the bundle, its passphrase and the folder to import
// it as, the bundle's own when left out; a bundle of the default
// workspace needs one. It checks the folder as a create does, makes its
// directory with its home and src, writes its keys, adds its rows under
// the folder, clones its home repository and then its projects. Up to
// the clones a failure keeps nothing; a home or a project that does not
// clone leaves an empty directory and reports git's output. For admins
// alone, as the export is: the bundle writes private keys into the key
// folder.
func (h *Handlers) importWorkspace(w http.ResponseWriter, r *http.Request) {
	if !access.UserFrom(r.Context()).CanAdmin() {
		fail(w, http.StatusForbidden, "only an admin imports a workspace: the bundle holds private keys and secrets")
		return
	}
	if h.keysOff(w) {
		return
	}
	bundle, target, status, problem := h.readImport(w, r)
	if problem != "" {
		fail(w, status, problem)
		return
	}
	ctx := r.Context()
	if status, problem := h.freeWorkspace(ctx, target); problem != "" {
		fail(w, status, problem)
		return
	}
	undo := &rollback{}
	row, projects, err := h.restore(ctx, bundle, target, undo)
	if err != nil {
		undo.run()
		switch {
		case errors.Is(err, errImport):
			fail(w, http.StatusBadRequest, err.Error())
		default:
			failErr(w, err)
		}
		return
	}
	unknown := h.unknownUsers(ctx, bundle.Manifest.Grants)
	home := importedProject{Folder: layout.HomeID(target)}
	if row.HomeURL != "" {
		output, err := h.makeHome(ctx, target, row.HomeURL, row.HomeBranch)
		home.Cloned, home.Output = err == nil, output
		if err != nil && output == "" {
			home.Output = err.Error()
		}
	}
	cloned := h.cloneImported(ctx, projects)
	h.writeImportedEnv(ctx, target, bundle.Manifest.EnvFiles)
	writeJSON(w, http.StatusCreated, map[string]any{"workspace": row, "home": home, "projects": cloned, "unknown_users": unknown})
}

// readImport reads the multipart form of an import: the bundle,
// decrypted, and the folder it goes to. A problem comes with its
// status.
func (h *Handlers) readImport(w http.ResponseWriter, r *http.Request) (*workspace.Bundle, string, int, string) {
	r.Body = http.MaxBytesReader(w, r.Body, workspace.MaxBundle+1<<20)
	if err := r.ParseMultipartForm(workspace.MaxBundle + 1<<20); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return nil, "", http.StatusRequestEntityTooLarge, fmt.Sprintf("a bundle is at most %d MiB", workspace.MaxBundle>>20)
		}
		return nil, "", http.StatusBadRequest, "send a multipart form: bundle, passphrase and folder"
	}
	defer r.MultipartForm.RemoveAll()
	passphrase := r.FormValue("passphrase")
	if passphrase == "" {
		return nil, "", http.StatusBadRequest, "enter the bundle's passphrase"
	}
	file, _, err := r.FormFile("bundle")
	if err != nil {
		return nil, "", http.StatusBadRequest, "pick the bundle file to import"
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, workspace.MaxBundle+1))
	if err != nil {
		return nil, "", http.StatusBadRequest, err.Error()
	}
	if len(data) > workspace.MaxBundle {
		return nil, "", http.StatusRequestEntityTooLarge, fmt.Sprintf("a bundle is at most %d MiB", workspace.MaxBundle>>20)
	}
	bundle, err := workspace.ReadBundle(bytes.NewReader(data), passphrase)
	if err != nil {
		return nil, "", http.StatusBadRequest, err.Error()
	}
	target := strings.TrimSpace(r.FormValue("folder"))
	source := bundle.Manifest.Workspace.Folder
	if target == "" && source == model.Default {
		return nil, "", http.StatusBadRequest, "the bundle holds the default workspace, _: name the folder it imports as"
	}
	if target == "" {
		target = source
	}
	return bundle, target, 0, ""
}

// rollback undoes the steps of an import that failed, the last first.
type rollback struct {
	steps []func()
}

func (u *rollback) add(step func()) { u.steps = append(u.steps, step) }

func (u *rollback) run() {
	for i := len(u.steps) - 1; i >= 0; i-- {
		u.steps[i]()
	}
}

// restore makes the workspace of a bundle under the target folder: its
// directory, its row, its keys and the rows of what it holds, each
// undone on a failure. It returns the workspace and the projects to
// clone.
func (h *Handlers) restore(ctx context.Context, bundle *workspace.Bundle, target string, undo *rollback) (*model.Workspace, []model.Project, error) {
	manifest := bundle.Manifest
	source := manifest.Workspace
	dir := layout.WorkspaceDir(h.opts.Root, target)
	if err := h.opts.Keys.CheckBundle(ctx, bundle.Keys); err != nil {
		return nil, nil, err
	}
	if err := os.Mkdir(dir, 0o755); err != nil {
		return nil, nil, err
	}
	undo.add(func() { os.RemoveAll(dir) })
	if err := layout.Ensure(h.opts.Root, target); err != nil {
		return nil, nil, err
	}

	row := &model.Workspace{
		Folder: target, Title: source.Title, Description: source.Description,
		Sort: source.Sort, Shallow: source.Shallow,
		HomeURL: manifest.Home.URL, HomeBranch: manifest.Home.Branch,
	}
	if source.Folder == model.Default {
		row.Title = target
	}
	if err := h.opts.Workspaces.Create(ctx, row); err != nil {
		return nil, nil, err
	}
	undo.add(func() { _ = h.opts.Workspaces.Delete(context.WithoutCancel(ctx), row.ID) })
	scope := keystore.ScopeWorkspace(row.ID)
	undo.add(func() { _ = h.opts.Keys.RemoveScope(context.WithoutCancel(ctx), scope) })
	roles, users := grants(manifest.Grants)
	if _, err := h.opts.Keys.Restore(ctx, scope, source.Folder, target, bundle.Keys, roles, users); err != nil {
		if errors.Is(err, keystore.ErrExists) {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("%w: %v", errImport, err)
	}
	if err := h.restoreEnvFiles(ctx, target, manifest.EnvFiles, undo); err != nil {
		return nil, nil, err
	}
	if err := h.restoreSecrets(ctx, row, manifest.Secrets, undo); err != nil {
		return nil, nil, err
	}
	projects, err := h.restoreProjects(ctx, target, manifest.Projects, undo)
	if err != nil {
		return nil, nil, err
	}
	return row, projects, nil
}

// importedFolder returns the folder a relative one of a bundle names
// under the target, refusing one that is not a folder ID inside it, or
// that is the workspace's home.
func importedFolder(target, relative string) (string, error) {
	id := workspace.Absolute(target, relative)
	if !folder.ValidID(id) || (id != target && !strings.HasPrefix(id, target+"/")) || layout.IsHome([]string{target}, id) {
		return "", fmt.Errorf("%w names the folder %q, which is not one inside %s", errImport, relative, target)
	}
	return id, nil
}

// grants returns a bundle's grants as the roles and the users of each
// SSH key, by name.
func grants(list []model.WorkspaceKeyGrant) (map[string][]string, map[string][]string) {
	roles, users := map[string][]string{}, map[string][]string{}
	for _, grant := range list {
		if grant.Role != "" {
			roles[grant.KeyName] = append(roles[grant.KeyName], grant.Role)
		}
		if grant.UserName != "" {
			users[grant.KeyName] = append(users[grant.KeyName], grant.UserName)
		}
	}
	return roles, users
}

// restoreEnvFiles adds the bundle's env files under the target. A row
// a removed workspace left at the same folder and path takes the
// bundle's content; a rollback puts its own back.
func (h *Handlers) restoreEnvFiles(ctx context.Context, target string, files []model.EnvFile, undo *rollback) error {
	if h.opts.Env == nil {
		return nil
	}
	for _, file := range files {
		id, err := importedFolder(target, file.Folder)
		if err != nil {
			return err
		}
		if _, err := envfile.SafeRelative(file.Path); err != nil {
			return fmt.Errorf("%w: env file %s: %v", errImport, file.Path, err)
		}
		row := &model.EnvFile{Folder: id, Path: file.Path, Content: file.Content}
		previous, err := h.opts.Env.Managed.Get(ctx, *row)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return err
		}
		if err := h.opts.Env.Managed.Set(ctx, row); err != nil {
			return err
		}
		undo.add(func() {
			if previous != nil {
				_ = h.opts.Env.Managed.Update(context.WithoutCancel(ctx), previous)
				return
			}
			_ = h.opts.Env.Managed.Delete(context.WithoutCancel(ctx), row.ID)
		})
	}
	return nil
}

// restoreSecrets adds the bundle's secrets under the target, which it
// carries plain inside its passphrase, encrypted with this instance's
// encryption key as they go into the vault. A row left at the same
// scope and name takes the bundle's, as an env file's does.
func (h *Handlers) restoreSecrets(ctx context.Context, row *model.Workspace, secrets []model.Secret, undo *rollback) error {
	if len(secrets) == 0 {
		return nil
	}
	if h.opts.Secrets == nil || h.opts.Secrets.Workspaces == nil {
		return fmt.Errorf("%w holds secrets, and this instance keeps none per workspace", errImport)
	}
	for _, s := range secrets {
		scope := workspace.AbsoluteScope(row.Folder, s.Scope)
		if s.Scope != storage.ScopeWorkspace {
			id, err := importedFolder(row.Folder, s.Scope)
			if err != nil {
				return err
			}
			scope = id
		}
		if !storage.ValidName(s.Name) {
			return fmt.Errorf("%w: %q is not a variable name", errImport, s.Name)
		}
		content, err := h.opts.Secrets.Vault.Encrypt(s.Content)
		if err != nil {
			return err
		}
		added := &model.Secret{Scope: scope, Name: s.Name, Content: content}
		previous, err := h.opts.Secrets.Secrets.Get(ctx, *added)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return err
		}
		if err := h.opts.Secrets.Secrets.Set(ctx, added); err != nil {
			return err
		}
		undo.add(func() {
			if previous != nil {
				_ = h.opts.Secrets.Secrets.Update(context.WithoutCancel(ctx), previous)
				return
			}
			_ = h.opts.Secrets.Secrets.Delete(context.WithoutCancel(ctx), added.ID)
		})
	}
	return nil
}

// restoreProjects adds the bundle's project rows under the target,
// parents before the projects inside them, the order they clone in.
func (h *Handlers) restoreProjects(ctx context.Context, target string, projects []model.Project, undo *rollback) ([]model.Project, error) {
	result := make([]model.Project, 0, len(projects))
	for _, project := range projects {
		id, err := importedFolder(target, project.Folder)
		if err != nil {
			return nil, err
		}
		if id == target {
			return nil, fmt.Errorf("%w holds a project at the workspace's own folder", errImport)
		}
		if _, err := h.opts.Projects.GetByFolder(ctx, id); err == nil {
			return nil, fmt.Errorf("%w: a project named %s exists already", errImport, id)
		}
		project.ID, project.Folder, project.CreatedAt, project.UpdatedAt = "", id, nil, nil
		if err := h.opts.Projects.Create(ctx, &project); err != nil {
			return nil, err
		}
		id = project.ID
		undo.add(func() { _ = h.opts.Projects.Delete(context.WithoutCancel(ctx), id) })
		result = append(result, project)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Folder < result[j].Folder })
	return result, nil
}

// cloneImported clones each imported project's repository into the
// workspace's src, or makes the directory of one without. A failure
// keeps the row and reports git's output.
func (h *Handlers) cloneImported(ctx context.Context, projects []model.Project) []importedProject {
	result := make([]importedProject, 0, len(projects))
	for _, project := range projects {
		dir := h.dirOf(ctx, project.Folder)
		item := importedProject{Folder: project.Folder}
		if project.SourceURL == "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				item.Output = err.Error()
			}
			result = append(result, item)
			continue
		}
		output, err := h.opts.Git.Clone(ctx, project.SourceURL, dir, project.Branch)
		item.Cloned, item.Output = err == nil, output
		if err != nil && output == "" {
			item.Output = err.Error()
		}
		result = append(result, item)
	}
	return result
}

// writeImportedEnv writes each imported env file into its folder, with
// the secrets appended, as a save from the UI does, once the clones
// made the folders: a folder that is not there, such as a project's
// whose clone failed, gets its file before its next run instead.
func (h *Handlers) writeImportedEnv(ctx context.Context, target string, files []model.EnvFile) {
	if h.opts.Env == nil {
		return
	}
	for _, file := range files {
		id := workspace.Absolute(target, file.Folder)
		dir := h.dirOf(ctx, id)
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			continue
		}
		_ = h.opts.Env.Save(ctx, &folder.Folder{ID: id, Dir: dir}, envfile.File{Path: file.Path, Content: file.Content})
	}
}

// unknownUsers returns the users the grants name that this instance
// does not have: their grants stay, and reach no run until a user
// of the name signs in.
func (h *Handlers) unknownUsers(ctx context.Context, grants []model.WorkspaceKeyGrant) []string {
	unknown := []string{}
	for _, grant := range grants {
		if grant.UserName == "" || slices.Contains(unknown, grant.UserName) {
			continue
		}
		if _, err := h.opts.Users.GetByName(ctx, grant.UserName); err != nil {
			unknown = append(unknown, grant.UserName)
		}
	}
	sort.Strings(unknown)
	return unknown
}
