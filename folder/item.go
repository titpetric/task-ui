package folder

import (
	"time"
)

// Item is an entry of a directory as a page lists it.
type Item struct {
	Name string
	// Dir says the entry is a directory.
	Dir bool
	// Size is a file's size in bytes; a directory has none.
	Size int64
	// ModTime is when the entry last changed.
	ModTime time.Time
	// Href is where the name links to: the directory to go into, or the
	// file to open. Empty leaves the name plain.
	Href string
	// Path is the entry's path from the top of the tree a page lists, for
	// the actions on it; empty where a page has none.
	Path string
	// Archive is where a directory downloads as a .tgz; empty without.
	Archive string
	// Sealed says task-ui lists the entry but does not open it or go
	// into it: see Sealed.
	Sealed bool
	// Exec says the entry is a file that runs: executable, or with a
	// shebang on its first line.
	Exec bool
}
