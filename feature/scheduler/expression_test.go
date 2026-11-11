package scheduler_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/scheduler"
)

func at(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}

func TestParse(t *testing.T) {
	t.Parallel()
	for _, text := range []string{"* * * * *", "0 * * * *", "*/15 9-17 * * mon-fri", "0 0 1,15 * *", "30 4 * jan,jul *", "0 0 * * 7", "@hourly", "@daily", "@weekly", "@monthly", "@yearly", "@annually", "@midnight"} {
		expr, err := scheduler.Parse(text)
		assert.NoError(t, err, text)
		assert.NotNil(t, expr, text)
	}
	for _, bad := range []string{"", "* * * *", "60 * * * *", "* 24 * * *", "* * 0 * *", "* * * 13 *", "* * * * 8", "a * * * *", "*/0 * * * *", "5-1 * * * *", "* * * * mon-xyz", "@never"} {
		_, err := scheduler.Parse(bad)
		assert.ErrorIs(t, err, scheduler.ErrInvalid, bad)
	}
}

func TestExpression_String(t *testing.T) {
	t.Parallel()
	expr, err := scheduler.Parse(" @Hourly ")
	assert.NoError(t, err)
	assert.Equal(t, "0 * * * *", expr.String())
}

func TestExpression_Matches(t *testing.T) {
	t.Parallel()
	expr, err := scheduler.Parse("*/15 9-17 * * mon-fri")
	assert.NoError(t, err)
	assert.True(t, expr.Matches(at(2026, time.October, 19, 9, 0)), "Monday 09:00")
	assert.True(t, expr.Matches(at(2026, time.October, 19, 17, 45)))
	assert.False(t, expr.Matches(at(2026, time.October, 19, 9, 5)))
	assert.False(t, expr.Matches(at(2026, time.October, 19, 18, 0)))
	assert.False(t, expr.Matches(at(2026, time.October, 18, 9, 0)), "Sunday")

	sunday, err := scheduler.Parse("0 0 * * 7")
	assert.NoError(t, err)
	assert.True(t, sunday.Matches(at(2026, time.October, 18, 0, 0)), "7 is Sunday")
}

func TestExpression_Next(t *testing.T) {
	t.Parallel()
	cases := []struct {
		expr  string
		after time.Time
		want  time.Time
	}{
		{"* * * * *", at(2026, time.October, 17, 10, 25), at(2026, time.October, 17, 10, 26)},
		{"0 * * * *", at(2026, time.October, 17, 10, 25), at(2026, time.October, 17, 11, 0)},
		{"@daily", at(2026, time.October, 17, 10, 25), at(2026, time.October, 18, 0, 0)},
		{"30 4 1 * *", at(2026, time.October, 17, 10, 25), at(2026, time.November, 1, 4, 30)},
		{"0 9 * * mon", at(2026, time.October, 17, 10, 25), at(2026, time.October, 19, 9, 0)},
		{"0 0 29 2 *", at(2026, time.January, 1, 0, 0), at(2028, time.February, 29, 0, 0)},
		{"0 0 1 1 *", at(2026, time.December, 31, 23, 59), at(2027, time.January, 1, 0, 0)},
	}
	for _, c := range cases {
		expr, err := scheduler.Parse(c.expr)
		assert.NoError(t, err, c.expr)
		assert.Equal(t, c.want, expr.Next(c.after), c.expr)
	}
	never, err := scheduler.Parse("0 0 31 2 *")
	assert.NoError(t, err)
	assert.True(t, never.Next(at(2026, time.January, 1, 0, 0)).IsZero())
}

func TestNextRun(t *testing.T) {
	t.Parallel()
	next, err := scheduler.NextRun("@hourly", at(2026, time.October, 17, 10, 25))
	assert.NoError(t, err)
	assert.Equal(t, at(2026, time.October, 17, 11, 0), next)
	_, err = scheduler.NextRun("nope", time.Now())
	assert.ErrorIs(t, err, scheduler.ErrInvalid)
	_, err = scheduler.NextRun("0 0 31 2 *", time.Now())
	assert.ErrorIs(t, err, scheduler.ErrInvalid)
}
