package runner

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// collectArtifacts records every file a run left under its artifact
// directory. With an artifact store, the files go into its tree as one
// commit, the artifacts name that commit, and the run's directory goes;
// the store squashes its history when it grows past what it keeps.
// Without one, or when the store fails, the files stay where the run wrote
// them. It removes an empty directory.
func (m *Manager) collectArtifacts(ctx context.Context, session *Session) {
	run := session.Run
	dir := m.artifactDir(run)
	if dir == "" || m.opts.Artifacts == nil {
		return
	}
	if _, err := os.Stat(dir); err != nil {
		return
	}
	var found []*model.Artifact
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !entry.Type().IsRegular() {
			return err
		}
		artifact, err := describeArtifact(run, dir, path, entry)
		if err != nil {
			return err
		}
		found = append(found, artifact)
		return nil
	})
	if err != nil {
		slog.Warn("artifacts not collected", "run", run.ID, "err", err)
	}
	if len(found) == 0 {
		_ = os.RemoveAll(dir)
		return
	}
	record := func(commit string) error {
		for _, artifact := range found {
			if commit != "" {
				artifact.Path, artifact.CommitID = artifact.Name, commit
			}
			if err := m.opts.Artifacts.Create(ctx, artifact); err != nil {
				slog.Warn("artifact not recorded", "run", run.ID, "name", artifact.Name, "err", err)
			}
		}
		return nil
	}
	if m.storeArtifacts(ctx, run, dir, record) == "" {
		_ = record("")
		return
	}
	_ = os.RemoveAll(dir)
	m.squashArtifacts(ctx)
}

// squashArtifacts squashes the store's history when it grows past what
// it keeps, and the records with it.
func (m *Manager) squashArtifacts(ctx context.Context) {
	_, err := m.opts.ArtifactStore.SquashThen(ctx, func(head string) error {
		if err := m.opts.Artifacts.Squash(ctx, head); err != nil {
			slog.Warn("artifact versions not updated after the squash", "err", err)
		}
		return nil
	})
	if err != nil {
		slog.Warn("artifact history not squashed", "err", err)
	}
}

// WriteArtifact puts a file into the artifact tree at a path, as a user
// and with no run, over the API or MCP: the store commits it under the
// user's name and keeps the version it replaces, and the record names the
// user where a run's names its run, folder and job.
func (m *Manager) WriteArtifact(ctx context.Context, rel string, data []byte, user string) (*model.Artifact, error) {
	if m.opts.ArtifactStore == nil || m.opts.Artifacts == nil {
		return nil, ErrNotFound
	}
	artifact := &model.Artifact{
		Name:        rel,
		Path:        rel,
		ByteSize:    int64(len(data)),
		ContentType: ContentType(rel),
		UserName:    user,
	}
	_, err := m.opts.ArtifactStore.WriteThen(ctx, rel, data, user, func(commit string) error {
		artifact.CommitID = commit
		return m.opts.Artifacts.Create(ctx, artifact)
	})
	if err != nil {
		return nil, err
	}
	m.squashArtifacts(ctx)
	return artifact, nil
}

// storeArtifacts commits a run's directory into the artifact store,
// records the artifacts with record while it holds the tree, and returns
// the commit, or "" without a store or when it fails.
func (m *Manager) storeArtifacts(ctx context.Context, run *model.Run, dir string, record func(commit string) error) string {
	if m.opts.ArtifactStore == nil {
		return ""
	}
	author := run.UserName
	if author == "" {
		author = "task-ui"
	}
	commit, err := m.opts.ArtifactStore.IngestThen(ctx, dir, run.Folder+"/"+run.Job+" run "+run.ID, author, record)
	if err != nil || commit == "" {
		slog.Warn("artifacts not stored", "run", run.ID, "err", err)
		return ""
	}
	return commit
}

// describeArtifact reads one file as an artifact row: its path relative
// to the artifact root, its size and a content type from its name.
func describeArtifact(run *model.Run, dir, path string, entry fs.DirEntry) (*model.Artifact, error) {
	info, err := entry.Info()
	if err != nil {
		return nil, err
	}
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return nil, err
	}
	return &model.Artifact{
		RunID:       run.ID,
		Folder:      run.Folder,
		Job:         run.Job,
		Name:        filepath.ToSlash(rel),
		Path:        filepath.ToSlash(filepath.Join(run.ID, rel)),
		ByteSize:    info.Size(),
		ContentType: ContentType(rel),
	}, nil
}

// ContentType guesses a media type from a file name, text/plain when the
// name says nothing.
func ContentType(name string) string {
	if kind := mime.TypeByExtension(strings.ToLower(filepath.Ext(name))); kind != "" {
		return kind
	}
	return "application/octet-stream"
}

// ArtifactPath returns where an artifact's file is under the root.
func (m *Manager) ArtifactPath(artifact *model.Artifact) string {
	return filepath.Join(m.opts.ArtifactRoot, filepath.FromSlash(artifact.Path))
}

// OpenArtifact opens an artifact's file for reading: the version its
// commit holds in the artifact store, or the file in its run's directory.
func (m *Manager) OpenArtifact(artifact *model.Artifact) (io.ReadCloser, error) {
	if artifact.CommitID != "" {
		if m.opts.ArtifactStore == nil {
			return nil, ErrNotFound
		}
		data, err := m.opts.ArtifactStore.Show(context.Background(), artifact.CommitID, artifact.Path)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(bytes.NewReader(data)), nil
	}
	return os.Open(m.ArtifactPath(artifact))
}

// EmptyArtifacts deletes every artifact: the store's tree and history, the
// files kept in finished runs' directories, and the records. A running
// run's directory stays, so the run can still write to it.
func (m *Manager) EmptyArtifacts(ctx context.Context) error {
	if m.opts.Artifacts == nil {
		return nil
	}
	all, err := m.opts.Artifacts.List(ctx, storage.ArtifactFilter{})
	if err != nil {
		return err
	}
	for _, artifact := range all {
		if artifact.CommitID == "" && m.opts.ArtifactRoot != "" {
			_ = os.RemoveAll(filepath.Join(m.opts.ArtifactRoot, artifact.RunID))
		}
	}
	if m.opts.ArtifactStore != nil {
		if err := m.opts.ArtifactStore.Empty(); err != nil {
			return err
		}
	}
	return m.opts.Artifacts.DeleteAll(ctx)
}

// crlfWriter turns the newlines of a plain command into the line ends a
// terminal expects, which a pty would otherwise do.
type crlfWriter struct {
	w io.Writer
}

func (c crlfWriter) Write(p []byte) (int, error) {
	_, err := c.w.Write([]byte(strings.ReplaceAll(string(p), "\n", "\r\n")))
	return len(p), err
}
