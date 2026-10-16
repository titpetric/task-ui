package repository

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/titpetric/task-ui/server/model"
)

// History stores run recordings as asciicast v2 files, one folder per
// project and job: <root>/<project>/<job>/<id>.cast, with the run
// outcome next to it in <id>.json.
type History struct {
	Root string
}

var validHistoryID = regexp.MustCompile(`^[0-9]{8}T[0-9]{9}-[0-9a-f]{6}$`)

// ValidID reports whether id is a history ID this package generates. It is
// checked before an ID is used in a path.
func ValidID(id string) bool {
	return validHistoryID.MatchString(id)
}

func (h *History) dir(project, job string) string {
	// Job names may contain ':' (namespaced tasks), which is fine on
	// the file systems this runs on, but never a path separator.
	return filepath.Join(h.Root, filepath.Base(project), strings.ReplaceAll(job, "/", "_"))
}

// CastPath returns the recording path for a run.
func (h *History) CastPath(project, job, id string) (string, error) {
	if !ValidID(id) {
		return "", fmt.Errorf("invalid history id %q", id)
	}
	return filepath.Join(h.dir(project, job), id+".cast"), nil
}

// List returns the recorded runs of a job, newest first.
func (h *History) List(project, job string) []model.HistoryRecord {
	files, _ := filepath.Glob(filepath.Join(h.dir(project, job), "*.json"))
	result := make([]model.HistoryRecord, 0, len(files))
	for _, file := range files {
		if record, err := readRecord(file); err == nil {
			result = append(result, record)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Started.After(result[j].Started) })
	return result
}

// Recent returns the newest runs across all projects.
func (h *History) Recent(limit int) []model.HistoryRecord {
	files, _ := filepath.Glob(filepath.Join(h.Root, "*", "*", "*.json"))
	// IDs start with the timestamp, so the newest files sort last.
	sort.Slice(files, func(i, j int) bool { return filepath.Base(files[i]) > filepath.Base(files[j]) })
	result := []model.HistoryRecord{}
	for _, file := range files {
		if len(result) >= limit {
			break
		}
		if record, err := readRecord(file); err == nil {
			result = append(result, record)
		}
	}
	return result
}

// Last returns the most recent run of a job.
func (h *History) Last(project, job string) (model.HistoryRecord, bool) {
	files, _ := filepath.Glob(filepath.Join(h.dir(project, job), "*.json"))
	sort.Strings(files)
	for i := len(files) - 1; i >= 0; i-- {
		if record, err := readRecord(files[i]); err == nil {
			return record, true
		}
	}
	return model.HistoryRecord{}, false
}

func readRecord(path string) (model.HistoryRecord, error) {
	var record model.HistoryRecord
	data, err := os.ReadFile(path)
	if err != nil {
		return record, err
	}
	err = json.Unmarshal(data, &record)
	return record, err
}

// Recorder writes one run as an asciicast v2 stream.
type Recorder struct {
	mu     sync.Mutex
	file   *os.File
	buf    *bufio.Writer
	meta   string
	start  time.Time
	Record model.HistoryRecord
}

// NewID returns a sortable, unique history ID.
func NewID(t time.Time) string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	// Millisecond resolution keeps runs started in the same second in order.
	stamp := t.UTC().Format("20060102T150405.000")
	return strings.Replace(stamp, ".", "", 1) + "-" + hex.EncodeToString(b)
}

// Start opens a recording for a run. The record is written as running, so a
// run is visible in the history while it happens.
func (h *History) Start(record model.HistoryRecord, cols, rows int) (*Recorder, error) {
	dir := h.dir(record.Project, record.Job)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	if record.ID == "" {
		record.ID = NewID(record.Started)
	}
	if record.State == "" {
		record.State = model.StateRunning
	}
	file, err := os.Create(filepath.Join(dir, record.ID+".cast"))
	if err != nil {
		return nil, err
	}
	r := &Recorder{
		file:   file,
		buf:    bufio.NewWriter(file),
		meta:   filepath.Join(dir, record.ID+".json"),
		start:  record.Started,
		Record: record,
	}
	header, _ := json.Marshal(map[string]any{
		"version":   2,
		"width":     cols,
		"height":    rows,
		"timestamp": record.Started.Unix(),
		"title":     record.Project + " / " + record.Job,
	})
	r.buf.Write(header)
	r.buf.WriteByte('\n')
	return r, r.writeMeta()
}

// Write records terminal output.
func (r *Recorder) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	event, _ := json.Marshal([]any{time.Since(r.start).Seconds(), "o", string(p)})
	r.buf.Write(event)
	r.buf.WriteByte('\n')
	return len(p), nil
}

// Finish records the outcome and closes the recording.
func (r *Recorder) Finish(state string, exitCode int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Record.State = state
	r.Record.ExitCode = exitCode
	r.Record.Duration = time.Since(r.start)
	return errors.Join(r.buf.Flush(), r.file.Close(), r.writeMeta())
}

func (r *Recorder) writeMeta() error {
	data, err := json.MarshalIndent(r.Record, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.meta, data, 0o644)
}
