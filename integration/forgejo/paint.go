package forgejo

import (
	"io"
	"os"
)

// Paint colours the report of a Sync with ANSI escapes: the plan
// symbols, the names by kind and the dim rows nothing changes. Off, every
// method returns its text plain, so a report to a file or a pipe stays
// plain. Colored says when a writer wants it.
type Paint bool

// ANSI escapes Paint writes.
const (
	ansiReset   = "\x1b[0m"
	ansiBold    = "\x1b[1m"
	ansiDim     = "\x1b[2m"
	ansiRed     = "\x1b[31m"
	ansiGreen   = "\x1b[32m"
	ansiYellow  = "\x1b[33m"
	ansiBlue    = "\x1b[34m"
	ansiMagenta = "\x1b[35m"
	ansiCyan    = "\x1b[36m"
)

// Colored reports whether a report to w wants colour: when w is a
// terminal or the command runs in a run of task-ui, TASK_UI_RUN_ID set,
// whose pty is one, and never with NO_COLOR in the environment.
func Colored(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TASK_UI_RUN_ID") != "" {
		return true
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// Bold is for the owner headers and the footer's head.
func (p Paint) Bold(s string) string { return p.wrap(ansiBold, s) }

// Dim is for a row in place or skipped.
func (p Paint) Dim(s string) string { return p.wrap(ansiDim, s) }

// Red is for "!", a row left alone or failed.
func (p Paint) Red(s string) string { return p.wrap(ansiRed, s) }

// Green is for "+", a row sync creates.
func (p Paint) Green(s string) string { return p.wrap(ansiGreen, s) }

// Yellow is for "~", a row sync brings current.
func (p Paint) Yellow(s string) string { return p.wrap(ansiYellow, s) }

// Blue is for a project name.
func (p Paint) Blue(s string) string { return p.wrap(ansiBlue, s) }

// Magenta is for a user login.
func (p Paint) Magenta(s string) string { return p.wrap(ansiMagenta, s) }

// Cyan is for an organisation and a workspace name, its home too.
func (p Paint) Cyan(s string) string { return p.wrap(ansiCyan, s) }

// wrap puts s between an escape and the reset, or returns it plain.
func (p Paint) wrap(code, s string) string {
	if !p || s == "" {
		return s
	}
	return code + s + ansiReset
}
