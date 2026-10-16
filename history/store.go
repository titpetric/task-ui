package history

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Store keeps the recordings as files, one folder per folder and job:
// <root>/<folder>/<job>/<id>.cast, with the record next to it in
// <id>.json.
type Store struct {
	Root string
}

func (s *Store) dir(folder, job string) string {
	// Job IDs may contain ':' (kind:name), which is fine on the file
	// systems this runs on, but never a path separator.
	return filepath.Join(s.Root, filepath.Base(folder), strings.ReplaceAll(job, "/", "_"))
}

// CastPath returns the recording path of a run.
func (s *Store) CastPath(record Record) (string, error) {
	if !ValidID(record.ID) {
		return "", fmt.Errorf("invalid history id %q", record.ID)
	}
	return filepath.Join(s.dir(record.Folder, record.Job), record.ID+".cast"), nil
}

// List returns the recorded runs of a job, newest first.
func (s *Store) List(folder, job string) []Record {
	files, _ := filepath.Glob(filepath.Join(s.dir(folder, job), "*.json"))
	result := make([]Record, 0, len(files))
	for _, file := range files {
		if record, err := readRecord(file); err == nil {
			result = append(result, record)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Started.After(result[j].Started) })
	return result
}

// Recent returns the newest runs across all folders.
func (s *Store) Recent(limit int) []Record {
	files, _ := filepath.Glob(filepath.Join(s.Root, "*", "*", "*.json"))
	// IDs start with the timestamp, so the newest files sort last.
	sort.Slice(files, func(i, j int) bool { return filepath.Base(files[i]) > filepath.Base(files[j]) })
	result := []Record{}
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
func (s *Store) Last(folder, job string) (Record, bool) {
	files, _ := filepath.Glob(filepath.Join(s.dir(folder, job), "*.json"))
	sort.Strings(files)
	for i := len(files) - 1; i >= 0; i-- {
		if record, err := readRecord(files[i]); err == nil {
			return record, true
		}
	}
	return Record{}, false
}

func readRecord(path string) (Record, error) {
	var record Record
	data, err := os.ReadFile(path)
	if err != nil {
		return record, err
	}
	err = json.Unmarshal(data, &record)
	return record, err
}

// Start opens a recording for a run. The record is written as running, so a
// run is visible in the history while it happens.
func (s *Store) Start(record Record, cols, rows int) (*Recorder, error) {
	dir := s.dir(record.Folder, record.Job)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	if record.ID == "" {
		record.ID = NewID(record.Started)
	}
	if record.State == "" {
		record.State = StateRunning
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
		"title":     record.Folder + " / " + record.Job,
	})
	r.buf.Write(header)
	r.buf.WriteByte('\n')
	return r, r.writeMeta()
}
