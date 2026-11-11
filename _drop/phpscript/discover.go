package phpscript

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/titpetric/phpscript/annotations"
)

// ModInfo is the file a module is found by.
const ModInfo = "modinfo.php"

// skipped are the directory names a walk never descends into: composer writes
// the first, a package manager the second, and the third is a checkout's own
// history.
var skipped = map[string]bool{"vendor": true, "node_modules": true, ".git": true}

// Found is a modinfo.php a walk reached, and the directory holding it, which
// is the module's root. Startup reports whether the file holds the @startup
// annotation; a file without it is not run, and the Manager says so.
type Found struct {
	Root    string
	File    string
	Startup bool
}

// Discover walks each root for a modinfo.php and returns what it found, by
// path, so two starts register the modules in the same order.
//
// A root that is not there is not an error: an instance with no modules
// directory yet has no modules, which is the state of a fresh data folder
// before the embedded tree is extracted. A root it cannot read is.
//
// vendor, node_modules and .git are never descended into, so a composer
// dependency cannot publish a module of its own into the instance.
func Discover(roots ...string) ([]Found, error) {
	var found []Found
	seen := map[string]bool{}
	for _, root := range roots {
		if root == "" {
			continue
		}
		abs, err := filepath.Abs(root)
		if err != nil {
			return nil, err
		}
		if _, err := os.Stat(abs); errors.Is(err, fs.ErrNotExist) {
			continue
		}
		walked, err := walk(abs, seen)
		if err != nil {
			return nil, err
		}
		found = append(found, walked...)
	}
	sort.Slice(found, func(i, j int) bool { return found[i].Root < found[j].Root })
	return found, nil
}

// walk reads one root and reports every modinfo.php under it. A directory
// already seen through another root is left out, because two roots that
// overlap would otherwise register the same module twice.
func walk(root string, seen map[string]bool) ([]Found, error) {
	var found []Found
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			// A directory the server may not read is not a module and not a
			// reason to start without the modules that are readable.
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
		if entry.Name() != ModInfo {
			return nil
		}
		dir := filepath.Dir(path)
		if seen[dir] {
			return nil
		}
		seen[dir] = true
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		found = append(found, Found{Root: dir, File: path, Startup: annotations.HasStartup(source)})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return found, nil
}
