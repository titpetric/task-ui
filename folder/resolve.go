package folder

import (
	"strings"
)

// Resolve finds the folder a path under /f names, and what follows it:
// nothing for the folder itself, a job ID, or env/<file>. Folder IDs can
// be nested, so the longest folder that starts the path wins:
// "services/api/deploy" is the job deploy of services/api when that is
// a folder, and the folder services/api/deploy when that is one.
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
