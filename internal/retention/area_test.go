package retention_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/retention"
)

// The table has a row per kind in order, then the fixed rows, no key
// read log among them; a sized row says its bytes, the others their
// count alone.
func TestAreas(t *testing.T) {
	t.Parallel()
	titles := []string{}
	for _, area := range retention.Areas {
		titles = append(titles, area.Title)
		assert.NotEqual(t, "Key read log", area.Title)
	}
	assert.Equal(t, []string{"Run recordings", "Run history", "Prompt threads", "Workflow run history", "Artifact history", "Dictations", "Secret read log", "Expired sign-ins", "Tool history"}, titles)
	for _, kind := range retention.Kinds {
		assert.Equal(t, kind, retention.AreaOf(kind).Kind)
		assert.Equal(t, "", retention.AreaOf(kind).Fixed, kind)
	}
	assert.Equal(t, "Removed every cleanup, no setting.", retention.AreaOf(retention.Signins).Fixed)
	assert.True(t, retention.AreaOf(retention.Recordings).Limited())
	assert.False(t, retention.AreaOf(retention.Prompts).Limited())
	tooling := retention.AreaOf(retention.Tooling)
	assert.True(t, tooling.Sized, "the row shows the bytes")
	assert.False(t, tooling.Limited(), "but takes no size limit")
	assert.Equal(t, retention.Area{}, retention.AreaOf("key_reads"))

	// The voice window's own directory, and nothing that merely starts
	// with the same letters.
	assert.True(t, retention.Dictated("voice/2026-11-10T09-00-00-000.wav"))
	assert.True(t, retention.Dictated("voice/2026-11-10T09-00-00-000.1.reply.wav"))
	assert.False(t, retention.Dictated("voices/take.wav"))
	assert.False(t, retention.Dictated("reports/voice/take.wav"))

	usage := retention.Usage{Count: 2, Bytes: 2048}
	assert.Equal(t, "2 recordings, 2.0 KB", retention.AreaOf(retention.Recordings).Holds(usage))
	assert.Equal(t, "2 threads", retention.AreaOf(retention.Prompts).Holds(usage))
	assert.Equal(t, "2 rows", retention.AreaOf(retention.Signins).Holds(usage))
}
