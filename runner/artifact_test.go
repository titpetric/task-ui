package runner_test

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

func TestContentType(t *testing.T) {
	t.Parallel()
	assert.Contains(t, runner.ContentType("out.txt"), "text/plain")
	assert.Equal(t, "application/json", runner.ContentType("data.JSON"))
	assert.Equal(t, "application/octet-stream", runner.ContentType("binary"))
	// The preview window plays what answers an audio type, whether or not
	// the image this runs in holds a mime database.
	assert.Equal(t, "audio/mpeg", runner.ContentType("dictation.MP3"))
	assert.Contains(t, runner.ContentType("take.wav"), "audio/")
	assert.Contains(t, runner.ContentType("take.flac"), "audio/")
}

// Every path that answers a file of the artifact tree sandboxes it,
// because a run wrote the bytes: the page, the frame it opens in and the
// ?raw a client reads.
func TestSandboxFile(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	runner.SandboxFile(rec)
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	policy := rec.Header().Get("Content-Security-Policy")
	assert.Contains(t, policy, "default-src 'none'")
	assert.Contains(t, policy, "sandbox")
}

// A write with no artifact store is not found: the store is where a file
// of the tree lives.
func TestManager_WriteArtifact(t *testing.T) {
	t.Parallel()
	_, err := runner.New(runner.Options{}).WriteArtifact(context.Background(), "upload/a.txt", []byte("x"), "miguel")
	assert.True(t, errors.Is(err, runner.ErrNotFound), err)
}

// An artifact a run left opens from the run's directory; one a commit
// holds needs the store that commit is in.
func TestManager_OpenArtifact(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	assert.NoError(t, os.MkdirAll(filepath.Join(root, "run1"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(root, "run1", "out.txt"), []byte("hello"), 0o600))
	m := runner.New(runner.Options{ArtifactRoot: root})
	file, err := m.OpenArtifact(&model.Artifact{Path: "run1/out.txt"})
	assert.NoError(t, err)
	data, err := io.ReadAll(file)
	assert.NoError(t, err)
	assert.NoError(t, file.Close())
	assert.Equal(t, "hello", string(data))
	_, err = m.OpenArtifact(&model.Artifact{Path: "run1/out.txt", CommitID: "9f1c2d3"})
	assert.True(t, errors.Is(err, runner.ErrNotFound), err)
}

// Emptying the artifacts with no records to read does nothing.
func TestManager_EmptyArtifacts(t *testing.T) {
	t.Parallel()
	assert.NoError(t, runner.New(runner.Options{}).EmptyArtifacts(context.Background()))
}

func TestManager_ArtifactPath(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	m := runner.New(runner.Options{ArtifactRoot: root})
	artifact := &model.Artifact{Path: "run1/sub/out.txt"}
	assert.Equal(t, filepath.Join(root, "run1", "sub", "out.txt"), m.ArtifactPath(artifact))
}
