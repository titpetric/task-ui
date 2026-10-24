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
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
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
// on: the server's own, or in a run of task-ui, where TASK_UI_WORKSPACE
// names the run's workspace, with the root at TASK_UI_WORKSPACE_DIR.
func maintenanceConfig() (*app.Config, error) {
	dir := os.Getenv("TASK_UI_WORKSPACE_DIR")
	if os.Getenv("TASK_UI_RUN_ID") == "" || dir == "" {
		return app.NewConfig(nil)
	}
	return app.NewConfig([]string{"--workspace", dir})
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
	cleaner := app.NewCleaner(db, cfg)
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
	named, homes, err := layoutRows(ctx, db)
	if err != nil {
		return err
	}
	moves, err := layout.Plan(cfg.Root, named, homes)
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

// layoutRows reads the named workspaces and, by workspace, the name of
// the directory of its home project, from a database before or after the
// migration that replaced the home project with the home repository: a
// database already migrated names no home project.
func layoutRows(ctx context.Context, db storage.DB) ([]string, map[string]string, error) {
	var named []string
	if err := db.SelectContext(ctx, &named, "SELECT folder FROM workspace WHERE folder <> ? ORDER BY folder", model.Default); err != nil {
		return nil, nil, err
	}
	homes := map[string]string{}
	rows, err := db.QueryContext(ctx, "SELECT w.folder, p.folder FROM workspace w JOIN project p ON p.id = w.home_project_id WHERE w.home_project_id <> ''")
	if err != nil {
		return named, homes, nil
	}
	defer rows.Close()
	for rows.Next() {
		var workspace, project string
		if err := rows.Scan(&workspace, &project); err != nil {
			return nil, nil, err
		}
		homes[workspace] = path.Base(project)
	}
	return named, homes, rows.Err()
}
