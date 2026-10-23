// Package hours decides when a job may run: the rules an admin writes
// for a folder, such as "mon-fri 9-17" and "fri 9-13", and whether a
// moment falls inside them. Outside the hours a run needs Force and a
// reason.
package hours

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ErrInvalid reports a rule that does not parse, wrapped with
// what is wrong.
var ErrInvalid = errors.New("invalid hours")

// dayNames maps the names a rule may use to weekdays.
var dayNames = map[string]time.Weekday{
	"sun": time.Sunday, "sunday": time.Sunday,
	"mon": time.Monday, "monday": time.Monday,
	"tue": time.Tuesday, "tues": time.Tuesday, "tuesday": time.Tuesday,
	"wed": time.Wednesday, "wednesday": time.Wednesday,
	"thu": time.Thursday, "thur": time.Thursday, "thurs": time.Thursday, "thursday": time.Thursday,
	"fri": time.Friday, "friday": time.Friday,
	"sat": time.Saturday, "saturday": time.Saturday,
}

// Rule allows a span of a day on some weekdays: From inclusive, To
// exclusive, in minutes from midnight.
type Rule struct {
	Days [7]bool
	From int
	To   int
}

// Matches reports whether the rule covers the weekday of a time.
func (r Rule) Matches(t time.Time) bool {
	return r.Days[t.Weekday()]
}

// Allows reports whether a time falls inside the rule.
func (r Rule) Allows(t time.Time) bool {
	minute := t.Hour()*60 + t.Minute()
	return r.Matches(t) && minute >= r.From && minute < r.To
}

// String writes the rule back the way Parse reads it, whole hours as hours and finer spans as hh:mm.
func (r Rule) String() string {
	padded := r.From%60 != 0 || r.To%60 != 0
	return days(r.Days) + " " + clock(r.From, padded) + "-" + clock(r.To, padded)
}

// Rules are the rules of a folder or a job, in order. The last rule that
// covers a weekday decides it, so "mon-fri 9-17" followed by "fri 9-13"
// closes Friday afternoons. A day no rule covers stays closed. No rules at
// all means always.
type Rules []Rule

// Allows reports whether a time is inside the hours.
func (rules Rules) Allows(t time.Time) bool {
	if len(rules) == 0 {
		return true
	}
	decided := false
	allowed := false
	for _, rule := range rules {
		if rule.Matches(t) {
			decided, allowed = true, rule.Allows(t)
		}
	}
	return decided && allowed
}

// String writes the rules one per line.
func (rules Rules) String() string {
	parts := make([]string, 0, len(rules))
	for _, rule := range rules {
		parts = append(parts, rule.String())
	}
	return strings.Join(parts, "\n")
}

// Parse reads rules, one per line or comma separated: days as names or ranges, then a span, such as:
//
//	mon-fri 9-17
//	sat,sun 09:30-17:00
//	daily 0-24
//
// Empty text is no rules.
func Parse(text string) (Rules, error) {
	var rules Rules
	for _, line := range strings.FieldsFunc(text, func(r rune) bool { return r == '\n' || r == ';' }) {
		line, _, _ = strings.Cut(line, "#")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		rule, err := parseRule(line)
		if err != nil {
			return nil, fmt.Errorf("%w: %q: %w", ErrInvalid, line, err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// parseRule reads one rule: the days, then the span.
func parseRule(line string) (Rule, error) {
	fields := strings.Fields(strings.ToLower(line))
	if len(fields) < 2 {
		return Rule{}, errors.New("a rule is days and a span, like mon-fri 9-17")
	}
	days, err := parseDays(strings.Join(fields[:len(fields)-1], ","))
	if err != nil {
		return Rule{}, err
	}
	from, to, err := parseSpan(fields[len(fields)-1])
	if err != nil {
		return Rule{}, err
	}
	return Rule{Days: days, From: from, To: to}, nil
}

// parseDays reads a comma separated list of names and ranges.
func parseDays(text string) ([7]bool, error) {
	var days [7]bool
	for _, part := range strings.Split(text, ",") {
		part = strings.TrimSpace(part)
		switch part {
		case "":
			continue
		case "daily", "everyday", "every", "all", "any":
			days = [7]bool{true, true, true, true, true, true, true}
			continue
		case "weekdays", "workdays":
			for d := time.Monday; d <= time.Friday; d++ {
				days[d] = true
			}
			continue
		case "weekends", "weekend":
			days[time.Saturday], days[time.Sunday] = true, true
			continue
		}
		first, last, found := strings.Cut(part, "-")
		if !found {
			last = first
		}
		from, ok := dayNames[first]
		to, ok2 := dayNames[last]
		if !ok || !ok2 {
			return days, fmt.Errorf("unknown day %q", part)
		}
		for d := from; ; d = (d + 1) % 7 {
			days[d] = true
			if d == to {
				break
			}
		}
	}
	return days, nil
}

// parseSpan reads "9-17" or "09:30-17:00". The end may be 24.
func parseSpan(text string) (int, int, error) {
	first, last, found := strings.Cut(text, "-")
	if !found {
		return 0, 0, errors.New("a span is from-to, like 9-17")
	}
	from, err := parseClock(first)
	if err != nil {
		return 0, 0, err
	}
	to, err := parseClock(last)
	if err != nil {
		return 0, 0, err
	}
	if to <= from {
		return 0, 0, errors.New("the span ends before it starts")
	}
	return from, to, nil
}

// parseClock reads "9", "09", "9:30" or "17:00" as minutes from midnight.
func parseClock(text string) (int, error) {
	hour, minute, _ := strings.Cut(text, ":")
	h, err := strconv.Atoi(hour)
	if err != nil || h < 0 || h > 24 {
		return 0, fmt.Errorf("bad hour %q", text)
	}
	m := 0
	if minute != "" {
		if m, err = strconv.Atoi(minute); err != nil || m < 0 || m > 59 {
			return 0, fmt.Errorf("bad minute %q", text)
		}
	}
	if h == 24 && m > 0 {
		return 0, fmt.Errorf("bad hour %q", text)
	}
	return h*60 + m, nil
}

// clock formats minutes from midnight.
func clock(minutes int, padded bool) string {
	if !padded {
		return strconv.Itoa(minutes / 60)
	}
	return fmt.Sprintf("%02d:%02d", minutes/60, minutes%60)
}

// days names a weekday set the short way: a range when the set forms
// one, the names otherwise.
func days(set [7]bool) string {
	var names []string
	for d := time.Monday; ; d = (d + 1) % 7 {
		if set[d] {
			names = append(names, strings.ToLower(d.String()[:3]))
		}
		if d == time.Sunday {
			break
		}
	}
	switch len(names) {
	case 0:
		return "never"
	case 7:
		return "daily"
	}
	if len(names) > 2 && contiguous(set) {
		return names[0] + "-" + names[len(names)-1]
	}
	return strings.Join(names, ",")
}

// contiguous reports whether the set is one unbroken run starting Monday.
func contiguous(set [7]bool) bool {
	started, ended := false, false
	for d := time.Monday; ; d = (d + 1) % 7 {
		switch {
		case set[d] && ended:
			return false
		case set[d]:
			started = true
		case started:
			ended = true
		}
		if d == time.Sunday {
			break
		}
	}
	return true
}
