// Package recording writes the terminal output of a run as an asciicast
// v2 file, which the run dialog plays back. A row in storage describes
// the run; the recording holds only what it showed.
package recording

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/titpetric/task-ui/model"
)

// Store keeps the recordings under <root>/<folder>/<job>/<id>.cast.
type Store struct {
	Root string
}

// Path returns where a run's recording is, as the run names it.
func (s *Store) Path(run *model.Run) (string, error) {
	if !model.ValidID(run.ID) {
		return "", fmt.Errorf("invalid run id %q", run.ID)
	}
	return filepath.Join(s.Root, filepath.Base(run.Folder), strings.ReplaceAll(run.Job, "/", "_"), run.ID+".cast"), nil
}

// Relative returns the recording path relative to the root, the form the
// run row keeps.
func (s *Store) Relative(path string) string {
	rel, err := filepath.Rel(s.Root, path)
	if err != nil {
		return path
	}
	return rel
}

// Read opens the recording a run row names, for playback.
func (s *Store) Read(run *model.Run) (*os.File, error) {
	if run.CastPath == "" {
		return nil, os.ErrNotExist
	}
	return os.Open(filepath.Join(s.Root, filepath.FromSlash(run.CastPath)))
}

// Open starts a recording for a run.
func (s *Store) Open(run *model.Run) (*Recorder, error) {
	path, err := s.Path(run)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	started := time.Now()
	if run.StartedAt != nil {
		started = *run.StartedAt
	}
	r := &Recorder{path: path, file: file, buf: bufio.NewWriter(file), start: started}
	header, _ := json.Marshal(map[string]any{
		"version":   2,
		"width":     run.TermCols,
		"height":    run.TermRows,
		"timestamp": started.Unix(),
		"title":     run.Folder + " / " + run.Job,
	})
	r.buf.Write(header)
	r.buf.WriteByte('\n')
	return r, r.buf.Flush()
}
