package stt_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/feature/orb/stt"
	"github.com/titpetric/task-ui/internal/assert"
)

// New holds the whisper.cpp runner, which a Service built by hand does
// not: it reads its settings and transcribes nothing.
func TestNew(t *testing.T) {
	t.Parallel()
	bin, model := touch(t, "whisper-cli"), touch(t, "ggml-small.en.bin")
	values := map[string]string{"bin": bin, "model": model}
	s := stt.New(settings(values))
	assert.NotNil(t, s.Run)
	var got stt.Config = mustConfig(t, s)
	assert.Equal(t, bin, got.Bin)
	assert.Equal(t, model, got.Model)

	bare := &stt.Service{Settings: settings(values)}
	assert.Nil(t, bare.Run)
	_, err := bare.Text(context.Background(), []byte("RIFFclip"))
	assert.ErrorIs(t, err, stt.ErrOff)
}

// Transcriber is the seam: a test supplies its own and whisper.cpp stays shut.
func TestTranscriber(t *testing.T) {
	t.Parallel()
	var run stt.Transcriber = func(context.Context, string, []string, []string) error { return nil }
	assert.NoError(t, run(context.Background(), "", nil, nil))
}

func mustConfig(t *testing.T, s *stt.Service) stt.Config {
	t.Helper()
	got, err := s.Config(context.Background())
	assert.NoError(t, err)
	return got
}

// A configuration takes a binary and a model that are both files, and
// reads the thread count with four as the default and 32 as the cap.
func TestConfig(t *testing.T) {
	t.Parallel()
	bin, model := touch(t, "whisper-cli"), touch(t, "ggml-small.en.bin")
	for _, test := range []struct {
		name    string
		values  map[string]string
		threads int
		err     error
	}{
		{name: "nothing set", values: map[string]string{}, err: stt.ErrOff},
		{name: "no model", values: map[string]string{"bin": bin}, err: stt.ErrOff},
		{name: "model is not there", values: map[string]string{"bin": bin, "model": model + ".missing"}, err: stt.ErrOff},
		{name: "a directory is not a binary", values: map[string]string{"bin": filepath.Dir(bin), "model": model}, err: stt.ErrOff},
		{name: "default threads", values: map[string]string{"bin": bin, "model": model}, threads: 4},
		{name: "threads read", values: map[string]string{"bin": bin, "model": model, "threads": "8"}, threads: 8},
		{name: "threads capped", values: map[string]string{"bin": bin, "model": model, "threads": "900"}, threads: 32},
		{name: "threads refused", values: map[string]string{"bin": bin, "model": model, "threads": "nope"}, threads: 4},
	} {
		s := &stt.Service{Settings: settings(test.values)}
		got, err := s.Config(context.Background())
		if test.err != nil {
			assert.Error(t, err, test.name)
			assert.False(t, s.Ready(context.Background()), test.name)
			continue
		}
		assert.NoError(t, err, test.name)
		assert.Equal(t, test.threads, got.Threads, test.name)
		assert.True(t, s.Ready(context.Background()), test.name)
	}

	// A service with no settings at all is off, and so is the nil one a
	// server without the integration holds.
	_, err := (&stt.Service{}).Config(context.Background())
	assert.ErrorIs(t, err, stt.ErrOff)
	var off *stt.Service
	assert.False(t, off.Ready(context.Background()))
}

// A clip is written beside the transcript the binary leaves, and the
// words come back as one line with whisper's silence markers dropped.
func TestService_Text(t *testing.T) {
	t.Parallel()
	bin, model := touch(t, "whisper-cli"), touch(t, "ggml-small.en.bin")
	var ran []string
	var env []string
	s := &stt.Service{
		Settings: settings(map[string]string{"bin": bin, "model": model, "threads": "2"}),
		Run: func(_ context.Context, name string, args, environ []string) error {
			ran = append([]string{name}, args...)
			env = environ
			out := args[5]
			return os.WriteFile(out+".txt", []byte("[BLANK_AUDIO]\n Deploy the api.\n(silence)\n Then tell me.\n"), 0o600)
		},
	}
	said, err := s.Text(context.Background(), []byte("RIFFclip"))
	assert.NoError(t, err)
	assert.Equal(t, "Deploy the api. Then tell me.", said)
	assert.Equal(t, bin, ran[0])
	assert.Contains(t, strings.Join(ran, " "), "-m "+model)
	assert.Contains(t, strings.Join(ran, " "), "-t 2")
	assert.Contains(t, strings.Join(ran, " "), "-otxt")
	assert.Contains(t, strings.Join(env, " "), "LD_LIBRARY_PATH="+filepath.Dir(bin))

	// A clip with nothing in it, and one past the endpoint's own limit,
	// are refused before the transcriber is asked.
	_, err = s.Text(context.Background(), nil)
	assert.ErrorIs(t, err, stt.ErrEmpty)
	_, err = s.Text(context.Background(), make([]byte, stt.MaxClip+1))
	assert.ErrorIs(t, err, stt.ErrLarge)

	// A transcriber that writes nothing is an error and not an empty answer.
	quiet := &stt.Service{
		Settings: settings(map[string]string{"bin": bin, "model": model}),
		Timeout:  time.Second,
		Run:      func(context.Context, string, []string, []string) error { return nil },
	}
	_, err = quiet.Text(context.Background(), []byte("RIFFclip"))
	assert.Error(t, err)
}

// One instance reads as many clips at once as Limit says and refuses the
// next: a page listening for a name posts a piece every two seconds, and
// each read is a whisper process of its own.
func TestService_TextBusy(t *testing.T) {
	t.Parallel()
	bin, model := touch(t, "whisper-cli"), touch(t, "ggml-small.en.bin")
	held := make(chan struct{})
	ran := make(chan struct{}, 2)
	s := &stt.Service{
		Settings: settings(map[string]string{"bin": bin, "model": model}),
		Limit:    1,
		Run: func(_ context.Context, _ string, args, _ []string) error {
			ran <- struct{}{}
			<-held
			return os.WriteFile(args[5]+".txt", []byte(" Deploy the api.\n"), 0o600)
		},
	}
	over := make(chan struct{})
	go func() {
		_, _ = s.Text(context.Background(), []byte("RIFFclip"))
		close(over)
	}()
	<-ran

	_, err := s.Text(context.Background(), []byte("RIFFclip"))
	assert.ErrorIs(t, err, stt.ErrBusy)

	// The slot frees as the call ends, and the next clip is read.
	close(held)
	<-over
	said, err := s.Text(context.Background(), []byte("RIFFclip"))
	assert.NoError(t, err)
	assert.Equal(t, "Deploy the api.", said)
}

func settings(values map[string]string) func(context.Context) (map[string]string, error) {
	return func(context.Context) (map[string]string, error) { return values, nil }
}

// touch writes an empty file of the name in the test's own directory.
func touch(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	assert.NoError(t, os.WriteFile(path, nil, 0o600))
	return path
}
