package diskusage_test

import (
	"testing"

	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestUsage_Percent(t *testing.T) {
	t.Parallel()
	u := diskusage.Usage{Files: 25, History: 75}
	assert.Equal(t, 25.0, u.FilesPercent())
	assert.Equal(t, 75.0, u.HistoryPercent())
	assert.Equal(t, 0.0, diskusage.Usage{}.FilesPercent())
	assert.Equal(t, 0.0, diskusage.Usage{}.HistoryPercent())
}
