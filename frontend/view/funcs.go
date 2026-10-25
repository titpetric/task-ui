package view

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/titpetric/task-ui/claude"
	"github.com/titpetric/task-ui/hours"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
)

// Funcs returns the functions the templates use.
func Funcs() template.FuncMap {
	return template.FuncMap{
		"ago":          Ago,
		"stamp":        Stamp,
		"duration":     Duration,
		"size":         Size,
		"join":         strings.Join,
		"backupName":   keystore.BackupName,
		"split":        strings.Split,
		"lower":        strings.ToLower,
		"add":          func(a, b int) int { return a + b },
		"sub":          func(a, b int) int { return a - b },
		"initials":     func(name string) string { return (&model.User{Name: name}).Initials() },
		"states":       model.States,
		"roles":        model.Roles,
		"list":         func(items ...any) []any { return items },
		"hoursLine":    HoursLine,
		"tone":         Tone,
		"assetVersion": AssetVersion,
		"folderDir":    FolderDir,
		"groups":       Groups,
		"hasPrefix":    strings.HasPrefix,
		"message":      Message,
		"markdown":     Markdown,
		"excerpt":      Excerpt,
		"json":         JSON,
		"ended":        model.Finished,
		"argLine":      ArgLine,
		// menu, item, danger, detail and sep build a [...] menu for the
		// menu template.
		"menu":   NewMenu,
		"item":   NewMenuItem,
		"danger": NewDangerItem,
		"detail": NewMenuDetail,
		"sep":    func() MenuItem { return MenuItem{Sep: true} },
		// head and primary build the head of a block for the blockhead
		// template: its title, its empty line and its buttons, the main
		// one from primary.
		"head":    NewHead,
		"primary": NewPrimaryItem,
		// who builds the "who gets it" element for the whofield template.
		"who": NewWho,
		// removeConfirm is what Remove checkout of a project asks.
		"removeConfirm": RemoveConfirm,
		// deleteWorkspaceConfirm is what Delete of a workspace asks.
		"deleteWorkspaceConfirm": DeleteWorkspaceConfirm,
		// deleteBlocked is the menu line of a workspace whose secrets stop its delete.
		"deleteBlocked": DeleteBlocked,
		// bytes and megabytes word a size in bytes, or in MB as Beszel
		// counts memory, as size does.
		"bytes":     func(n uint64) string { return Size(int64(n)) },
		"megabytes": func(v float64) string { return Size(int64(v * (1 << 20))) },
		"rate":      func(v float64) string { return Size(int64(v)) + "/s" },
		// summaryPrompt is the short summary instruction, which a
		// workflow's summary step sets beside its question,
		// workflowSummaryQuestion; newStep is a blank step of a kind,
		// for the editor's Add step.
		"summaryPrompt":           func() string { return claude.RunSummary },
		"workflowSummaryQuestion": func() string { return note.WorkflowSummaryQuestion },
		"newStep":                 NewStep,
		// summaryQuestion is the question of Run and summarize, which
		// Summarize on a run that ended asks as well.
		"summaryQuestion": func() string { return note.SummaryQuestion },
	}
}

// ArgLine returns the args a schedule or a workflow step keeps, a JSON list, as the line its form takes back:
// the words, quoted where a word needs it, or the command unchanged for a job of @host:<host>.
func ArgLine(job, args string) string {
	list := model.DecodeList(args)
	if strings.HasPrefix(job, model.HostPrefix) {
		return strings.Join(list, " ")
	}
	words := make([]string, len(list))
	for i, word := range list {
		words[i] = word
		if word == "" || strings.ContainsAny(word, " \t'\"\\") {
			words[i] = "'" + strings.ReplaceAll(word, "'", `'\''`) + "'"
		}
	}
	return strings.Join(words, " ")
}

// JSON encodes a value for an attribute a script reads; a value that
// does not encode is null.
func JSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return string(data)
}

// Tone names the badge class of a run state: ok, signal while running,
// bad for a failed or killed run, plain for a stopped one.
func Tone(state string) string {
	switch state {
	case model.StateOK:
		return "ok"
	case model.StateRunning:
		return "signal"
	case model.StateFailed, model.StateKilled:
		return "bad"
	}
	return ""
}

// InHours reports whether the hours allow a run at a moment. Hours that
// do not parse are not the page's problem: they read as open.
func InHours(text string, now time.Time) bool {
	rules, err := hours.Parse(text)
	if err != nil {
		return true
	}
	return rules.Allows(now)
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

// Ago says how much time has passed since a time, in the coarsest unit
// that is not zero. A missing time is a dash.
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
	if value >= unit {
		value, suffix = value/unit, "TB"
	}
	return fmt.Sprintf("%.1f %s", value, suffix)
}

// FolderDir is the sidebar heading of a folder under top: "dockerfiles" for deploy/dockerfiles/blog in deploy.
// FolderDir returns "" for deploy/dockerfiles and for deploy itself.
// In a workspace the project is two segments deep: malta/deploy/web/blog under malta/deploy sits in "web".
func FolderDir(id, top string) string {
	rest, ok := strings.CutPrefix(id, top+"/")
	if !ok {
		return ""
	}
	if i := strings.LastIndex(rest, "/"); i >= 0 {
		return rest[:i]
	}
	return ""
}
