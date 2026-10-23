package folder

import (
	"strings"
)

// Resolve finds the folder a path under /f names, and the rest: "", a job ID, or env/<file>.
// The longest folder ID that starts the path wins, as folder IDs nest.
func Resolve(c Catalogue, path string) (*Folder, string) {
	path = strings.Trim(path, "/")
	var best *Folder
	rest := ""
	for _, f := range c.List() {
		if path == f.ID {
			return f, ""
		}
		if strings.HasPrefix(path, f.ID+"/") && (best == nil || len(f.ID) > len(best.ID)) {
			best, rest = f, path[len(f.ID)+1:]
		}
	}
	return best, rest
}
