package layout

// Move is one step of a layout migration: a directory that goes from
// From to To, or a directory made when From is empty. A migration moves
// the directory of a workspace's home project, by workspace folder, to
// home, every other directory of the workspace, task-ui's own names and
// the workspaces at the root left alone, under src, and makes home and
// src where they are not; Plan lists the moves and Apply makes them.
type Move struct {
	From string
	To   string
}

// String prints a move as the migration lists it.
func (m Move) String() string {
	if m.From == "" {
		return "mkdir " + m.To
	}
	return "mv " + m.From + " " + m.To
}
