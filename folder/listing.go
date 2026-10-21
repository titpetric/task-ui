package folder

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/titpetric/task-ui/detector"
)

// Listing returns what a directory holds: the directories first, then the
// files, each by name regardless of case, with sizes and times. Listing
// keeps dot entries; the sealed ones say so. A symlink lists as what it
// points to, and Listing seals it when it seals that. A directory
// Listing cannot read lists nothing.
func Listing(dir string) []Item {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	items := make([]Item, 0, len(entries))
	for _, entry := range entries {
		item := Item{Name: entry.Name(), Dir: entry.IsDir()}
		full := filepath.Join(dir, entry.Name())
		info, err := entry.Info()
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			if target, err := os.Stat(full); err == nil {
				info, item.Dir = target, target.IsDir()
			}
			item.Sealed = sealedTarget(full, item.Dir)
		}
		item.Sealed = item.Sealed || Sealed(item.Name, item.Dir)
		if err == nil {
			item.ModTime = info.ModTime()
			if !item.Dir {
				item.Size = info.Size()
				item.Exec = !item.Sealed && info.Mode().IsRegular() && (info.Mode()&0o111 != 0 || detector.Shebang(full) != nil)
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

// sealedTarget reports whether a symlink leads to a sealed entry: a dot
// file, .git, or something inside .git. dir says the target is a
// directory.
func sealedTarget(link string, dir bool) bool {
	target, err := filepath.EvalSymlinks(link)
	if err != nil {
		return false
	}
	target = filepath.ToSlash(target)
	return Sealed(path.Base(target), dir) || strings.Contains(target+"/", "/"+repository+"/")
}
