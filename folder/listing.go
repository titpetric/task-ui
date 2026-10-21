package folder

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/titpetric/task-ui/detector"
)

// Listing returns what a directory holds: the directories first, then the
// files, each by name regardless of case, with sizes and times. Listing
// keeps dot entries; the sealed ones say so. A symlink lists as what it
// points to. A directory Listing cannot read lists nothing.
func Listing(dir string) []Item {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	items := make([]Item, 0, len(entries))
	for _, entry := range entries {
		item := Item{Name: entry.Name(), Dir: entry.IsDir(), Sealed: Sealed(entry.Name())}
		full := filepath.Join(dir, entry.Name())
		info, err := entry.Info()
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			if target, err := os.Stat(full); err == nil {
				info, item.Dir = target, target.IsDir()
			}
		}
		if err == nil {
			item.ModTime = info.ModTime()
			if !item.Dir {
				item.Size = info.Size()
				item.Exec = info.Mode().IsRegular() && (info.Mode()&0o111 != 0 || detector.Shebang(full) != nil)
			}
		}
		items = append(items, item)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Dir != items[j].Dir {
			return items[i].Dir
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
	return items
}
