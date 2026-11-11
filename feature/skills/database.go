package skills

import (
	"os"
	"path/filepath"
)

// AtkinsDB returns the ATKINS_DB of a run with HOME at home: the home's
// history, then the instance's mirror under serverHome, when there is one.
func AtkinsDB(home, serverHome string) string {
	own := filepath.Join(home, ".atkins", "atkins.db")
	if serverHome == "" {
		return own
	}
	return own + string(os.PathListSeparator) + filepath.Join(serverHome, ".atkins", "atkins.db")
}
