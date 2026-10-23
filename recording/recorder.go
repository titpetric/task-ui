package recording

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
)

// Recorder writes one run's output as asciicast events. As an
// io.Writer, it takes exactly what the terminal showed.
type Recorder struct {
	mu    sync.Mutex
	path  string
	file  *os.File
	buf   *bufio.Writer
	start time.Time
}

// Path returns the file the recorder writes.
func (r *Recorder) Path() string {
	return r.path
}

// Write records terminal output.
func (r *Recorder) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	event, _ := json.Marshal([]any{time.Since(r.start).Seconds(), "o", string(p)})
	r.buf.Write(event)
	r.buf.WriteByte('\n')
	return len(p), r.buf.Flush()
}

// Mark records a marker event, what happened to the run besides its
// output, such as who took input on it. Replays skip it.
func (r *Recorder) Mark(label string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	event, _ := json.Marshal([]any{time.Since(r.start).Seconds(), "m", label})
	r.buf.Write(event)
	r.buf.WriteByte('\n')
	return r.buf.Flush()
}

// Close finishes the recording.
func (r *Recorder) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return errors.Join(r.buf.Flush(), r.file.Close())
}
