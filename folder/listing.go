package folder

import (
	"os"
	"sort"
	"strings"
)

// Item is an entry of a directory as the folder page lists it.
type Item struct {
	Name string
	// Dir says the entry is a directory, which the page links into.
	Dir bool
}

// Listing returns what a directory holds, for the folder page: the
// directories first, then the files, each by name regardless of case.
// Listing skips dot entries. A directory Listing cannot read lists
// nothing.
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
		items = append(items, Item{Name: entry.Name(), Dir: entry.IsDir()})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Dir != items[j].Dir {
			return items[i].Dir
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
	return items
}
