package view_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestFuncs(t *testing.T) {
	t.Parallel()

	funcs := view.Funcs()
	for _, name := range []string{"ago", "stamp", "duration", "size", "join", "lower", "add", "initials", "mask", "states", "roles", "list"} {
		assert.NotNil(t, funcs[name], name)
	}
	assert.Equal(t, 3, funcs["add"].(func(a, b int) int)(1, 2))
	assert.Equal(t, "OP", funcs["initials"].(func(string) string)("ops"))
	assert.Len(t, funcs["list"].(func(...any) []any)(1, "a"), 2)
}

func TestAgo(t *testing.T) {
	t.Parallel()

	now := time.Now()
	assert.Equal(t, "-", view.Ago(nil))
	assert.Equal(t, "-", view.Ago((*time.Time)(nil)))
	assert.Equal(t, "-", view.Ago(time.Time{}))
	assert.Equal(t, "just now", view.Ago(now))
	assert.Equal(t, "just now", view.Ago(&now))
	minutes := now.Add(-5 * time.Minute)
	assert.Equal(t, "5m ago", view.Ago(&minutes))
	assert.Equal(t, "3h ago", view.Ago(now.Add(-3*time.Hour)))
	assert.Equal(t, "4d ago", view.Ago(now.Add(-4*24*time.Hour)))
}

func TestStamp(t *testing.T) {
	t.Parallel()

	at := time.Date(2026, time.October, 17, 9, 30, 0, 0, time.Local)
	assert.Equal(t, "2026-10-17 09:30:00", view.Stamp(at))
	assert.Equal(t, "2026-10-17 09:30:00", view.Stamp(&at))
	assert.Equal(t, "-", view.Stamp(nil))
}

func TestDuration(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "0.0s", view.Duration(-time.Second))
	assert.Equal(t, "1.5s", view.Duration(1500*time.Millisecond))
	assert.Equal(t, "2m05s", view.Duration(125*time.Second))
	assert.Equal(t, "1h01m", view.Duration(61*time.Minute))
}

func TestSize(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "512 B", view.Size(512))
	assert.Equal(t, "1.5 KB", view.Size(1536))
	assert.Equal(t, "2.0 MB", view.Size(2<<20))
	assert.Equal(t, "3.0 GB", view.Size(3<<30))
}

func TestMask(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "********", view.Mask("short"))
	assert.Equal(t, "hun********", view.Mask("hunter2secret"))
}

func TestInHours(t *testing.T) {
	t.Parallel()

	assert.True(t, view.InHours(""))
	assert.True(t, view.InHours("daily 0-24"))
	assert.False(t, view.InHours("mon 0:00-0:01"))
	assert.True(t, view.InHours("not hours at all"))
}

func TestHoursLine(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "mon-fri 9-17, fri 9-13", view.HoursLine("mon-fri 9-17\nfri 9-13"))
	assert.Equal(t, "", view.HoursLine(""))
	assert.Equal(t, "broken", view.HoursLine(" broken "))
}
