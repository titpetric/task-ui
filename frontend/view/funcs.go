package view

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/titpetric/task-ui/hours"
	"github.com/titpetric/task-ui/model"
)

// Funcs returns the functions the templates use.
func Funcs() template.FuncMap {
	return template.FuncMap{
		"ago":       Ago,
		"stamp":     Stamp,
		"duration":  Duration,
		"size":      Size,
		"join":      strings.Join,
		"lower":     strings.ToLower,
		"add":       func(a, b int) int { return a + b },
		"initials":  func(name string) string { return (&model.User{Name: name}).Initials() },
		"mask":      Mask,
		"states":    model.States,
		"roles":     model.Roles,
		"list":      func(items ...any) []any { return items },
		"inHours":   InHours,
		"hoursLine": HoursLine,
	}
}

// InHours reports whether the hours allow a run now. Hours that do not
// parse are not the page's problem: they read as open.
func InHours(text string) bool {
	rules, err := hours.Parse(text)
	if err != nil {
		return true
	}
	return rules.Allows(time.Now())
}

// HoursLine writes hours on one line.
func HoursLine(text string) string {
	rules, err := hours.Parse(text)
	if err != nil {
		return strings.TrimSpace(text)
	}
	return strings.ReplaceAll(rules.String(), "\n", ", ")
}

// when returns the time a value stands for: a time, a pointer to one, or
// nothing.
func when(v any) (time.Time, bool) {
	switch t := v.(type) {
	case time.Time:
		return t, !t.IsZero()
	case *time.Time:
		if t == nil || t.IsZero() {
			return time.Time{}, false
		}
		return *t, true
	}
	return time.Time{}, false
}

// Ago says how long ago a time was, in the coarsest unit that is not
// zero. A missing time is a dash.
func Ago(v any) string {
	t, ok := when(v)
	if !ok {
		return "-"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}

// Stamp formats a time for a table cell.
func Stamp(v any) string {
	t, ok := when(v)
	if !ok {
		return "-"
	}
	return t.Local().Format("2006-01-02 15:04:05")
}

// Duration formats a duration the way a stopwatch shows it.
func Duration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	d = d.Round(100 * time.Millisecond)
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%02ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%02dm", int(d.Hours()), int(d.Minutes())%60)
}

// Size formats a byte count.
func Size(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	value, suffix := float64(n)/unit, "KB"
	if value >= unit {
		value, suffix = value/unit, "MB"
	}
	if value >= unit {
		value, suffix = value/unit, "GB"
	}
	return fmt.Sprintf("%.1f %s", value, suffix)
}

// Mask hides a value, keeping the first characters of a long one.
func Mask(value string) string {
	if len(value) <= 6 {
		return strings.Repeat("*", 8)
	}
	return value[:3] + strings.Repeat("*", 8)
}
