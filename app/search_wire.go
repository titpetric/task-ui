package app

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/feature/scheduler"
	"github.com/titpetric/task-ui/feature/search"
	"github.com/titpetric/task-ui/internal/folder"
)

// newSearchIndexer returns the index over the workspaces' projects and the catalogue's jobs.
func newSearchIndexer(cfg *Config, store *AppStorage, catalogue *Catalogue) *search.Indexer {
	index := search.NewIndexer(store.Search, cfg.Root)
	index.Sources = func(ctx context.Context) ([]search.Source, error) {
		workspaces, err := store.Workspaces.List(ctx)
		if err != nil {
			return nil, err
		}
		var sources []search.Source
		for _, workspace := range workspaces {
			entries, err := os.ReadDir(folder.SrcDir(cfg.Root, workspace.Folder))
			if err != nil {
				continue
			}
			for _, entry := range entries {
				if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
					continue
				}
				sources = append(sources, search.Source{
					Workspace: workspace.Folder, Project: entry.Name(),
					FolderID: folder.Join(workspace.Folder, entry.Name()),
					Dir:      folder.ProjectDir(cfg.Root, workspace.Folder, entry.Name()),
				})
			}
		}
		return sources, nil
	}
	index.Targets = func(context.Context) ([]search.Target, error) {
		var targets []search.Target
		for _, f := range catalogue.Every() {
			workspace, project := splitFolderID(f.ID)
			for _, job := range f.Jobs {
				if job.Source == "" {
					continue
				}
				rel, err := filepath.Rel(cfg.Root, filepath.Join(f.Dir, job.Source))
				if err != nil || strings.HasPrefix(rel, "..") {
					continue
				}
				targets = append(targets, search.Target{
					FolderID: f.ID, FolderTitle: f.Title, Workspace: workspace, Project: project,
					Dir: job.Dir, ID: job.ID, Name: job.Name, Kind: job.Kind, SourceFile: filepath.ToSlash(rel),
					Description: job.Description, Summary: job.Summary, Steps: job.Steps, Aliases: job.Aliases, Command: job.Command,
				})
			}
		}
		return targets, nil
	}
	index.After = func(ctx context.Context) error {
		added, err := (&scheduler.CIReconciler{Search: store.Search, Schedules: store.Schedules}).Run(ctx)
		if err != nil {
			return err
		}
		if added > 0 {
			slog.Info("ci:daily schedules registered", "count", added)
		}
		return nil
	}
	return index
}

// splitFolderID names the workspace and project a folder ID belongs to.
func splitFolderID(id string) (workspace, project string) {
	parts := strings.Split(model.Canonical(id), "/")
	switch len(parts) {
	case 0:
		return model.Default, ""
	case 1:
		return model.Default, parts[0]
	default:
		return parts[0], parts[1]
	}
}
