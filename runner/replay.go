package runner

// Replay is a recording played back: the screen it ended on, and its
// whole output, the same output in a terminal OutputRows tall without
// its empty rows at the bottom.
type Replay struct {
	Screen Frame
	Output Frame
}

// Lines counts the lines of the whole output, less the empty one the
// cursor waits on after the last.
func (r Replay) Lines() int {
	lines := r.Output.Rows
	if lines > 0 && blank(r.Output.Lines[lines-1]) {
		lines--
	}
	return lines
}

// Complete reports whether the screen already shows the whole output:
// nothing scrolled off it.
func (r Replay) Complete() bool {
	return r.Screen.Trim(1).Text() == r.Output.Text()
}
