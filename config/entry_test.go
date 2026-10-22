package config_test

import (
	"testing"

	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestEntry(t *testing.T) {
	t.Parallel()
	assert.Equal(t, config.Entry{Key: "note", Title: "Notes"}, config.Default().Sections[2])
}
