package model

import (
	"regexp"
	"strings"
)

// ChannelNotes is the channel of the notes board, where a note can be an
// instruction to the agent.
const ChannelNotes = "notes"

// The states of a note to the agent.
const (
	NoteOpen    = "open"
	NoteWorking = "working"
	NoteDone    = "done"
	NoteFailed  = "failed"
)

// mention matches @name at the start of the text or after a space or an
// opening mark: a user name of letters, digits, dot, dash or underscore.
var mention = regexp.MustCompile(`(?:^|[\s(\[{,;])@([A-Za-z0-9][A-Za-z0-9._-]*)`)

// Mentions returns the user names a text names with @, once each, in the
// order they appear. Mentions drops a trailing dot or dash, the punctuation.
func Mentions(text string) []string {
	var names []string
	seen := map[string]bool{}
	for _, m := range mention.FindAllStringSubmatch(text, -1) {
		name := strings.TrimRight(m[1], ".-")
		if name != "" && !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	return names
}

// EncodeMentions keeps mentions in a column, between spaces, so a query
// finds one with LIKE '% name %'.
func EncodeMentions(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return " " + strings.Join(names, " ") + " "
}

// MentionList returns the users a note names.
func (n *Note) MentionList() []string {
	return strings.Fields(n.Mentions)
}

// Mentioned reports whether a note names the user.
func (n *Note) Mentioned(name string) bool {
	return name != "" && strings.Contains(n.Mentions, " "+name+" ")
}

// Finished reports whether the agent ended the work of the note.
func (n *Note) Finished() bool {
	return n.State == NoteDone || n.State == NoteFailed
}
