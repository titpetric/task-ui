package skills

import (
	"path/filepath"
	"strings"
)

// When is the condition a skill or one of its jobs holds to, as atkins
// reports it: any of Files, looked for in the directory and its parents,
// or in the directory alone with Here, and with Contains the file must
// match that regular expression. Matched is the file that met it.
type When struct {
	Files    []string `json:"files,omitempty"`
	Here     bool     `json:"here,omitempty"`
	Contains string   `json:"contains,omitempty"`
	Matched  string   `json:"matched,omitempty"`
}

// String says the condition in a line, for the title of a chip:
// "compose.yml here, matched compose.yml".
func (w *When) String() string {
	if w == nil || len(w.Files) == 0 {
		return ""
	}
	line := strings.Join(w.Files, " or ")
	if w.Here {
		line += " here"
	}
	if w.Contains != "" {
		line += " containing " + w.Contains
	}
	if w.Matched != "" {
		line += ", matched " + filepath.Base(w.Matched)
	}
	return line
}
