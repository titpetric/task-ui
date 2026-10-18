package runner

// Cell is one character of a frame with its colours and attributes.
type Cell struct {
	Char    rune
	FG, BG  uint32
	Bold    bool
	Reverse bool
}
