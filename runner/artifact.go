package runner

import (
	"context"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/model"
)

// collectArtifacts records every file a run left under its artifact
// directory. It removes an empty directory. The files stay where the run
// wrote them: the row says where.
func (m *Manager) collectArtifacts(ctx context.Context, session *Session) {
	dir := m.artifactDir(session.Run)
	if dir == "" || m.opts.Artifacts == nil {
		return
	}
	if _, err := os.Stat(dir); err != nil {
		return
	}
	found := 0
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		artifact, err := describeArtifact(session.Run, dir, path, entry)
		if err != nil {
			return err
		}
		found++
		return m.opts.Artifacts.Create(ctx, artifact)
	})
	if err != nil {
		slog.Warn("artifacts not collected", "run", session.Run.ID, "err", err)
	}
	if found == 0 {
		_ = os.Remove(dir)
	}
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

// OpenArtifact opens an artifact's file for reading.
func (m *Manager) OpenArtifact(artifact *model.Artifact) (io.ReadCloser, error) {
	return os.Open(m.ArtifactPath(artifact))
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
