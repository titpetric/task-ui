// Package stt turns speech into words on this machine, with whisper.cpp.
//
// The instance holds no model and builds no binary: the stt folder's
// setup job does both, and the integration says where it put them. See
// docs/guide/voice.md.
package stt

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// The reasons a clip is not transcribed.
var (
	ErrOff   = errors.New("speech to text is not set up")
	ErrEmpty = errors.New("the clip holds no audio")
	ErrLarge = errors.New("the clip is too long to transcribe in one call")
	ErrBusy  = errors.New("the transcriber is busy: this clip was not read")
)

// MaxClip bounds a clip at what an artifact write takes, about eight
// minutes of the 16 kHz mono the recorder writes.
const MaxClip = 16 << 20

// Calls is how many clips one instance transcribes at once, for a service
// that names no other number. Each is a whisper.cpp process of up to 32
// threads, and a page listening for a name posts one every two seconds.
const Calls = 2

// Config is the integration's settings, read.
type Config struct {
	Bin     string
	Model   string
	Threads int
}

// Service transcribes a clip with the binary and model the integration names.
type Service struct {
	// Settings reads the integration's settings, decrypted.
	Settings func(ctx context.Context) (map[string]string, error)
	// Timeout caps one call; zero is two minutes.
	Timeout time.Duration
	// Run executes the transcriber, and New is what sets it; nil
	// transcribes nothing and answers ErrOff.
	Run Transcriber
	// Limit is how many clips are read at once; zero is Calls.
	Limit int

	opens sync.Once
	slots chan struct{}
}

// Transcriber runs the transcriber over the clip at in, leaving out.txt beside it.
type Transcriber func(ctx context.Context, bin string, args, env []string) error

// New returns a Service that runs whisper.cpp over the settings.
func New(settings func(ctx context.Context) (map[string]string, error)) *Service {
	return &Service{Settings: settings, Run: whisper}
}

// Config reads the integration's settings: ErrOff where the binary or the model is not a file.
func (s *Service) Config(ctx context.Context) (Config, error) {
	if s == nil || s.Settings == nil {
		return Config{}, ErrOff
	}
	values, err := s.Settings(ctx)
	if err != nil {
		return Config{}, err
	}
	c := Config{
		Bin:     strings.TrimSpace(values["bin"]),
		Model:   strings.TrimSpace(values["model"]),
		Threads: threadsOf(values["threads"]),
	}
	if c.Bin == "" || c.Model == "" || !isFile(c.Bin) || !isFile(c.Model) {
		return Config{}, ErrOff
	}
	return c, nil
}

// Ready reports whether a clip would be transcribed, for a page that draws the button.
func (s *Service) Ready(ctx context.Context) bool {
	_, err := s.Config(ctx)
	return err == nil
}

// Text returns the words of a WAV clip, empty where the transcriber heard none.
func (s *Service) Text(ctx context.Context, clip []byte) (string, error) {
	switch {
	case len(clip) == 0:
		return "", ErrEmpty
	case len(clip) > MaxClip:
		return "", ErrLarge
	}
	c, err := s.Config(ctx)
	if err != nil {
		return "", err
	}
	done, err := s.hold()
	if err != nil {
		return "", err
	}
	defer done()
	dir, err := os.MkdirTemp("", "stt-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)
	in := filepath.Join(dir, "clip.wav")
	if err := os.WriteFile(in, clip, 0o600); err != nil {
		return "", err
	}
	out := filepath.Join(dir, "clip")
	timeout := s.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := s.run(ctx, c, in, out); err != nil {
		return "", err
	}
	said, err := os.ReadFile(out + ".txt")
	if err != nil {
		return "", err
	}
	return clean(string(said)), nil
}

// hold takes one of the call slots and returns what frees it, ErrBusy
// where every slot is taken. A clip nobody read is dropped by its caller:
// a listener's piece is two seconds of a room, and queueing one behind a
// whisper run lands its words after the sentence has moved on.
func (s *Service) hold() (func(), error) {
	s.opens.Do(func() {
		n := s.Limit
		if n < 1 {
			n = Calls
		}
		s.slots = make(chan struct{}, n)
	})
	select {
	case s.slots <- struct{}{}:
		return func() { <-s.slots }, nil
	default:
		return nil, ErrBusy
	}
}

// run transcribes in, leaving out.txt beside it.
func (s *Service) run(ctx context.Context, c Config, in, out string) error {
	if s.Run == nil {
		return ErrOff
	}
	args := []string{
		"-m", c.Model,
		"-f", in,
		"-of", out,
		"-otxt",
		"-nt",
		"-np",
		"-t", strconv.Itoa(c.Threads),
	}
	// whisper.cpp builds its libraries beside the binary and links them by
	// soname, so a call from anywhere else finds none.
	return s.Run(ctx, c.Bin, args, append(os.Environ(), "LD_LIBRARY_PATH="+filepath.Dir(c.Bin)))
}

// whisper runs whisper-cli, answering with the exit status on a failure.
// Its output is logged and not returned: it names the model path and the
// temp directory, and a caller holding prompt/edit reads what is returned.
func whisper(ctx context.Context, bin string, args, env []string) error {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = env
	said, err := cmd.CombinedOutput()
	if err != nil {
		slog.Warn("the transcriber failed", "err", err, "output", strings.TrimSpace(string(said)))
		return errors.New("the transcriber failed: " + err.Error())
	}
	return nil
}

// clean is the transcript as one line, with the markers whisper writes for silence dropped.
func clean(said string) string {
	out := []string{}
	for _, line := range strings.Split(said, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "[") || strings.HasPrefix(line, "(") {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, " ")
}

// threadsOf is how many threads a call takes, bounded, or four.
func threadsOf(value string) int {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || n < 1 {
		return 4
	}
	return min(n, 32)
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
