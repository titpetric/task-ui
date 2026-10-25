package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/forgejo"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/storage"
)

// dryRun reads the arguments of a maintenance command: --dry-run or -n,
// and -- to end the flags; anything else is the usage error.
func dryRun(args []string, usage string) (bool, error) {
	dry := false
	for _, arg := range args {
		switch arg {
		case "--":
		case "--dry-run", "-n":
			dry = true
		default:
			return false, errors.New(usage)
		}
	}
	return dry, nil
}

// maintenanceConfig reads the configuration a maintenance command works
// on: the server's own, TASK_UI_DATA and TASK_UI_WORKSPACE, or in a run
// of task-ui, where TASK_UI_WORKSPACE names the run's workspace when it
// has one, the data directory the run inherits, TASK_UI_DATA, with its
// root at <data>/workspaces.
func maintenanceConfig() (*app.Config, error) {
	return app.NewConfig(nil)
}

// cleanup runs task-ui cleanup, the root's @cleanup: the Data retention
// rules applied to the data of the configuration, read as
// maintenanceConfig does. --dry-run prints what would go and changes
// nothing.
func cleanup(ctx context.Context, args []string, stdout io.Writer) error {
	dry, err := dryRun(args, "usage: task-ui cleanup [--dry-run]")
	if err != nil {
		return err
	}
	cfg, err := maintenanceConfig()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cfg.DB), 0o755); err != nil {
		return err
	}
	db, err := storage.OpenAndMigrate(ctx, cfg.DB)
	if err != nil {
		return err
	}
	defer db.Close()
	cleaner := app.NewCleaner(db, cfg, os.Getenv("HOME"))
	when := "now"
	if dry {
		when = "a dry run: nothing changes"
	}
	fmt.Fprintf(stdout, "task-ui cleanup of %s, %s\n", cfg.DataDir, when)
	return cleaner.Clean(ctx, dry, stdout)
}

// migrateLayout runs task-ui migrate-layout: the moves that bring the
// workspace folder to the home and src layout, read from the
// configuration and the database, and prints each; --dry-run
// prints them and moves nothing. A root laid out already prints nothing
// and changes nothing, so the command runs again without harm. Run it
// with the server stopped.
func migrateLayout(ctx context.Context, args []string, stdout io.Writer) error {
	dry, err := dryRun(args, "usage: task-ui migrate-layout [--dry-run]")
	if err != nil {
		return err
	}
	cfg, err := maintenanceConfig()
	if err != nil {
		return err
	}
	db, err := storage.Open(cfg.DB)
	if err != nil {
		return err
	}
	defer db.Close()
	workspaces, homes, err := layoutRows(ctx, db)
	if err != nil {
		return err
	}
	moves, err := layout.Plan(cfg.Root, workspaces, homes)
	if err != nil {
		return err
	}
	if len(moves) == 0 {
		_, err := fmt.Fprintf(stdout, "%s is laid out already: nothing to move\n", cfg.Root)
		return err
	}
	for _, move := range moves {
		fmt.Fprintln(stdout, move)
	}
	if dry {
		_, err := fmt.Fprintln(stdout, "a dry run: nothing moved")
		return err
	}
	return layout.Apply(moves)
}

// layoutRows reads the workspaces, the default one first when the
// instance has it, and, by workspace, the name of the directory of its
// home project, from a database before or after the migration that
// replaced the home project with the home repository: a database
// already migrated names no home project.
func layoutRows(ctx context.Context, db storage.DB) ([]string, map[string]string, error) {
	var workspaces []string
	if err := db.SelectContext(ctx, &workspaces, "SELECT folder FROM workspace ORDER BY CASE WHEN folder = ? THEN 0 ELSE 1 END, folder", model.Default); err != nil {
		return nil, nil, err
	}
	homes := map[string]string{}
	rows, err := db.QueryContext(ctx, "SELECT w.folder, p.folder FROM workspace w JOIN project p ON p.id = w.home_project_id WHERE w.home_project_id <> ''")
	if err != nil {
		return workspaces, homes, nil
	}
	defer rows.Close()
	for rows.Next() {
		var workspace, project string
		if err := rows.Scan(&workspace, &project); err != nil {
			return nil, nil, err
		}
		homes[workspace] = path.Base(project)
	}
	return workspaces, homes, rows.Err()
}

// errForgejoOff refuses a sync while the Forgejo integration lacks its
// URL or token.
var errForgejoOff = errors.New("the Forgejo integration is not set up: set its URL and token in Settings, Integrations")

// forgejoSync runs task-ui forgejo-sync, the root's @forgejo-sync: the
// Forgejo integration's organisations as workspaces and projects under
// the root of the configuration, read as maintenanceConfig does, with
// the integration's settings from the database and the key folder.
// --dry-run prints what it would do and changes nothing.
func forgejoSync(ctx context.Context, args []string, stdout io.Writer) error {
	dry, err := dryRun(args, "usage: task-ui forgejo-sync [--dry-run]")
	if err != nil {
		return err
	}
	cfg, err := maintenanceConfig()
	if err != nil {
		return err
	}
	if cfg.SSHDir == "" {
		cfg.SSHDir = filepath.Join(cfg.DataDir, "keys")
	}
	db, err := storage.OpenAndMigrate(ctx, cfg.DB)
	if err != nil {
		return err
	}
	defer db.Close()
	keys := &keystore.Local{Dir: cfg.SSHDir, Rows: storage.NewKeyInfoStorage(db), Settings: storage.NewSettingStorage(db)}
	secrets := secret.New(storage.NewSecretStorage(db), secret.NewVault(keys.Vault(storage.ScopeInstance)))
	values, err := (&integration.Store{Secrets: secrets}).Values(ctx, integration.Forgejo)
	if err != nil {
		return err
	}
	if !integration.Configured(*integration.KindOf(integration.Forgejo), values) {
		return errForgejoOff
	}
	sync := &forgejo.Sync{
		Client:     &forgejo.Client{URL: values["url"], Token: values["token"]},
		Git:        &git.Git{Keys: keys},
		Workspaces: storage.NewWorkspaceStorage(db),
		Projects:   storage.NewProjectStorage(db),
		Root:       cfg.Root,
		Secret:     values["secret"],
		HTTPS:      values["clone"] == "https",
		DryRun:     dry,
	}
	if cfg.URL != "" {
		sync.HookURL = cfg.URL + integration.WebhookPath
	}
	fmt.Fprintf(stdout, "task-ui forgejo-sync of %s from %s\n", cfg.Root, values["url"])
	return sync.Run(ctx, stdout)
}
