package recording_test

import (
	"os"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/recording"
)

func TestRecorder_Write(t *testing.T) {
	t.Parallel()
	s := &recording.Store{Root: t.TempDir()}
	rec, err := s.Open(&model.Run{ID: model.NewID(), Folder: "f", Job: "j"})
	assert.NoError(t, err)
	n, err := rec.Write([]byte("hello\r\n"))
	assert.NoError(t, err)
	assert.Equal(t, 7, n)
	assert.NoError(t, rec.Close())

	data, err := os.ReadFile(rec.Path())
	assert.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.Len(t, lines, 2)
	assert.Contains(t, lines[1], `"o","hello\r\n"`)
}

func TestRecorder_Close(t *testing.T) {
	t.Parallel()
	s := &recording.Store{Root: t.TempDir()}
	rec, err := s.Open(&model.Run{ID: model.NewID(), Folder: "f", Job: "j"})
	assert.NoError(t, err)
	assert.NoError(t, rec.Close())
	assert.Error(t, rec.Close(), "closing twice")
}
