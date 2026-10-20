package folder

import (
	"os"
	"path"
	"sort"
	"strings"
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
}

// previewable are the image types a page opens in the image window.
var previewable = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".svg": true}

// Preview reports whether the item is an image a page shows in a window
// in place of a download.
func (i Item) Preview() bool {
	return !i.Dir && i.Href != "" && previewable[strings.ToLower(path.Ext(i.Name))]
}

// Listing returns what a directory holds: the directories first, then the
// files, each by name regardless of case, with sizes and times. Listing
// skips dot entries. A directory Listing cannot read lists nothing.
func Listing(dir string) []Item {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	items := make([]Item, 0, len(entries))
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		item := Item{Name: entry.Name(), Dir: entry.IsDir()}
		if info, err := entry.Info(); err == nil {
			item.ModTime = info.ModTime()
			if !item.Dir {
				item.Size = info.Size()
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
