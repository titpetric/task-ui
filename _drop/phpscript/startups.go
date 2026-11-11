package phpscript

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/titpetric/phpscript/annotations"
)

// Startups returns the module's other @startup files, by path, without
// its modinfo.php: that one runs first and on its own, because nothing
// else may run before a module has registered.
//
// A module puts its migrations in one of these rather than requiring the
// file from modinfo.php, so a startup step is a file with an annotation
// and the list of them is the directory.
func Startups(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if entry != nil && entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			if path != root && skipped[entry.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".php" || entry.Name() == ModInfo {
			return nil
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if annotations.HasStartup(source) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}
