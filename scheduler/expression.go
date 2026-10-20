// Package scheduler starts jobs on their own: cron expressions, when they
// fire next, and the loop that starts the due ones as detached runs.
package scheduler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Expression is a parsed cron expression: five fields, minute to day of
// week, each a set of allowed values.
type Expression struct {
	text   string
	minute uint64
	hour   uint64
	dom    uint64
	month  uint64
	dow    uint64
}

// field describes one position of the expression.
type field struct {
	name     string
	min, max int
	names    []string
}

var fields = []field{
	{"minute", 0, 59, nil},
	{"hour", 0, 23, nil},
	{"day of month", 1, 31, nil},
	{"month", 1, 12, []string{"jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec"}},
	{"day of week", 0, 6, []string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}},
}

// aliases are the named schedules.
var aliases = map[string]string{
	"@yearly":   "0 0 1 1 *",
	"@annually": "0 0 1 1 *",
	"@monthly":  "0 0 1 * *",
	"@weekly":   "0 0 * * 0",
	"@daily":    "0 0 * * *",
	"@midnight": "0 0 * * *",
	"@hourly":   "0 * * * *",
}

// ErrInvalid reports an expression that does not parse.
var ErrInvalid = errors.New("invalid cron expression")

// Parse reads a cron expression: five fields of `*`, `*/n`, `a`, `a-b`,
// `a-b/n` and lists of those, or one of the @ aliases.
func Parse(text string) (*Expression, error) {
	text = strings.TrimSpace(strings.ToLower(text))
	if alias, ok := aliases[text]; ok {
		text = alias
	}
	parts := strings.Fields(text)
	if len(parts) != 5 {
		return nil, fmt.Errorf("%w: %q has %d fields, five are needed", ErrInvalid, text, len(parts))
	}
	expr := &Expression{text: text}
	targets := []*uint64{&expr.minute, &expr.hour, &expr.dom, &expr.month, &expr.dow}
	for i, part := range parts {
		set, err := parseField(fields[i], part)
		if err != nil {
			return nil, err
		}
		*targets[i] = set
	}
	// 7 is Sunday too, as crontab allows.
	if expr.dow&(1<<7) != 0 {
		expr.dow |= 1
	}
	return expr, nil
}

// parseField reads one field into a set of values.
func parseField(f field, text string) (uint64, error) {
	var set uint64
	for _, item := range strings.Split(text, ",") {
		bits, err := parseRange(f, item)
		if err != nil {
			return 0, err
		}
		set |= bits
	}
	return set, nil
}

// parseRange reads `*`, `a`, `a-b`, each with an optional `/n` step.
func parseRange(f field, item string) (uint64, error) {
	step := 1
	if base, after, ok := strings.Cut(item, "/"); ok {
		n, err := strconv.Atoi(after)
		if err != nil || n <= 0 {
			return 0, fmt.Errorf("%w: step %q in %s", ErrInvalid, after, f.name)
		}
		item, step = base, n
	}
	lo, hi := f.min, f.max
	if item != "*" {
		var err error
		lo, hi, err = parseBounds(f, item)
		if err != nil {
			return 0, err
		}
		if !strings.Contains(item, "-") && step == 1 {
			hi = lo
		}
	}
	var set uint64
	for v := lo; v <= hi; v += step {
		set |= 1 << uint(v)
	}
	return set, nil
}

// parseBounds reads `a` or `a-b`, numbers or names.
func parseBounds(f field, item string) (int, int, error) {
	from, to, isRange := strings.Cut(item, "-")
	lo, err := parseValue(f, from)
	if err != nil {
		return 0, 0, err
	}
	hi := lo
	if isRange {
		if hi, err = parseValue(f, to); err != nil {
			return 0, 0, err
		}
	}
	if hi < lo {
		return 0, 0, fmt.Errorf("%w: %q runs backwards in %s", ErrInvalid, item, f.name)
	}
	return lo, hi, nil
}

// parseValue reads one number or name of the field.
func parseValue(f field, text string) (int, error) {
	for i, name := range f.names {
		if text == name {
			return f.min + i, nil
		}
	}
	v, err := strconv.Atoi(text)
	if err != nil || v < f.min || v > f.max+boolInt(f.name == "day of week") {
		return 0, fmt.Errorf("%w: %q is not a %s", ErrInvalid, text, f.name)
	}
	return v, nil
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// String returns the expression as parsed.
func (e *Expression) String() string {
	return e.text
}

// Matches reports whether the expression fires at the minute of t.
func (e *Expression) Matches(t time.Time) bool {
	return e.minute&(1<<uint(t.Minute())) != 0 &&
		e.hour&(1<<uint(t.Hour())) != 0 &&
		e.dom&(1<<uint(t.Day())) != 0 &&
		e.month&(1<<uint(t.Month())) != 0 &&
		e.dow&(1<<uint(t.Weekday())) != 0
}

// Next returns the first minute after t at which the expression fires.
// A schedule that never fires, such as the 31st of February, returns the
// zero time after five years of looking.
func (e *Expression) Next(after time.Time) time.Time {
	t := after.Truncate(time.Minute).Add(time.Minute)
	limit := t.AddDate(5, 0, 0)
	for t.Before(limit) {
		switch {
		case e.month&(1<<uint(t.Month())) == 0:
			t = time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
		case e.dom&(1<<uint(t.Day())) == 0 || e.dow&(1<<uint(t.Weekday())) == 0:
			t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, t.Location())
		case e.hour&(1<<uint(t.Hour())) == 0:
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, t.Location())
		case e.minute&(1<<uint(t.Minute())) == 0:
			t = t.Add(time.Minute)
		default:
			return t
		}
	}
	return time.Time{}
}

// NextRun parses an expression and returns when it fires after t.
func NextRun(expression string, after time.Time) (time.Time, error) {
	expr, err := Parse(expression)
	if err != nil {
		return time.Time{}, err
	}
	next := expr.Next(after)
	if next.IsZero() {
		return time.Time{}, fmt.Errorf("%w: %q never fires", ErrInvalid, expression)
	}
	return next, nil
}
