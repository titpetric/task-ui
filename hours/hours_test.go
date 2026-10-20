package hours_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/hours"
	"github.com/titpetric/task-ui/internal/assert"
)

// at returns a weekday at a clock time, in the week of 2026-10-12.
func at(day time.Weekday, hour, minute int) time.Time {
	monday := time.Date(2026, time.October, 12, 0, 0, 0, 0, time.UTC)
	return monday.AddDate(0, 0, int(day+6)%7).Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)
}

func TestParse(t *testing.T) {
	t.Parallel()

	rules, err := hours.Parse("mon-fri 9-17\nfri 9-13")
	assert.NoError(t, err)
	assert.Len(t, rules, 2)
	assert.Equal(t, "mon-fri 9-17\nfri 9-13", rules.String())

	rules, err = hours.Parse("Mon,Wed 09:30-17:45; weekends 10-14 # comment free\n# a comment\n")
	assert.NoError(t, err)
	assert.Len(t, rules, 2)
	assert.Equal(t, "mon,wed 09:30-17:45\nsat,sun 10-14", rules.String())

	rules, err = hours.Parse("")
	assert.NoError(t, err)
	assert.Nil(t, rules)
	assert.True(t, rules.Allows(time.Now()))

	daily, err := hours.Parse("daily 0-24")
	assert.NoError(t, err)
	assert.Equal(t, "daily 0-24", daily.String())

	wrap, err := hours.Parse("sat-mon 8-12")
	assert.NoError(t, err)
	assert.Equal(t, "mon,sat,sun 8-12", wrap.String())

	for _, bad := range []string{"mon-fri", "someday 9-17", "mon 17-9", "mon 9:60-10", "mon 25-26", "mon 24:30-24:40", "mon x-y"} {
		_, err := hours.Parse(bad)
		assert.ErrorIs(t, err, hours.ErrInvalid, bad)
	}
}

func TestRules_Allows(t *testing.T) {
	t.Parallel()

	rules, err := hours.Parse("mon-fri 9-17\nfri 9-13")
	assert.NoError(t, err)

	assert.True(t, rules.Allows(at(time.Monday, 9, 0)))
	assert.True(t, rules.Allows(at(time.Wednesday, 16, 59)))
	assert.False(t, rules.Allows(at(time.Wednesday, 17, 0)))
	assert.False(t, rules.Allows(at(time.Monday, 8, 59)))
	// The later, narrower rule closes Friday afternoon.
	assert.True(t, rules.Allows(at(time.Friday, 12, 59)))
	assert.False(t, rules.Allows(at(time.Friday, 14, 0)))
	// Once any rule exists, days no rule names stay closed.
	assert.False(t, rules.Allows(at(time.Saturday, 10, 0)))
	assert.False(t, rules.Allows(at(time.Sunday, 10, 0)))

	// A later rule can open a day the earlier one closed.
	open, err := hours.Parse("weekdays 9-17\nsat 10-12")
	assert.NoError(t, err)
	assert.True(t, open.Allows(at(time.Saturday, 11, 0)))
	assert.False(t, open.Allows(at(time.Saturday, 13, 0)))
}

func TestRule(t *testing.T) {
	t.Parallel()

	rule := hours.Rule{From: 9 * 60, To: 17 * 60}
	rule.Days[time.Monday] = true
	assert.True(t, rule.Matches(at(time.Monday, 3, 0)))
	assert.False(t, rule.Matches(at(time.Tuesday, 10, 0)))
	assert.True(t, rule.Allows(at(time.Monday, 10, 0)))
	assert.False(t, rule.Allows(at(time.Monday, 3, 0)))
	assert.Equal(t, "mon 9-17", rule.String())
	assert.Equal(t, "mon 09:00-17:30", hours.Rule{Days: rule.Days, From: 9 * 60, To: 17*60 + 30}.String())

	assert.Equal(t, "never 9-17", hours.Rule{From: 9 * 60, To: 17 * 60}.String())
}
