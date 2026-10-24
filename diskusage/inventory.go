package diskusage

import (
	"io/fs"
	"path/filepath"
	"sync"
	"time"
)

// Item is one place an instance keeps data, as the Storage table names
// it: a database with its journal files, a directory of recordings, a
// workspace's checkouts. Paths are the files or directories whose
// regular files add up to the item, symbolic links not followed; a path
// that is not there counts as nothing.
type Item struct {
	Name  string
	Where string
	Paths []string
}

// Row is an item measured: the bytes its regular files take and how
// many there are, or Pending while the measure is not in yet.
type Row struct {
	Name    string
	Where   string
	Size    int64
	Files   int64
	Pending bool
}

// Table is the Storage table of the Instance page: a row per item and
// their total. Pending says a measure is under way and the rows show
// the last one; the page prints an ellipsis where there is none.
type Table struct {
	Rows    []Row
	Total   Row
	Pending bool
}

// Inventory measures the items of an instance for the Storage table,
// keeping the result for TTL and waiting at most Wait for a fresh one,
// so the Instance page answers at once with the last measure, or with
// pending rows, while a large tree walks in the background.
type Inventory struct {
	// Items names what to measure, read on every fresh measure: the
	// workspaces and their homes change.
	Items func() []Item
	TTL   time.Duration
	Wait  time.Duration

	mu      sync.Mutex
	table   Table
	at      time.Time
	running chan struct{}
}

// Table returns the measure of the items: the kept one while younger
// than TTL, else a fresh one when it comes within Wait, else the
// kept one marked pending, or pending rows when there is none yet.
func (i *Inventory) Table() Table {
	i.mu.Lock()
	if !i.at.IsZero() && time.Since(i.at) < i.TTL {
		defer i.mu.Unlock()
		return i.table
	}
	if i.running == nil {
		i.running = make(chan struct{})
		go i.measure(i.running)
	}
	done := i.running
	i.mu.Unlock()

	select {
	case <-done:
	case <-time.After(i.Wait):
	}
	i.mu.Lock()
	if i.running == nil {
		defer i.mu.Unlock()
		return i.table
	}
	table := Table{Pending: true, Total: Row{Name: "Total", Pending: true}}
	if !i.at.IsZero() {
		table.Rows, table.Total = i.table.Rows, i.table.Total
		i.mu.Unlock()
		return table
	}
	i.mu.Unlock()
	for _, item := range i.Items() {
		table.Rows = append(table.Rows, Row{Name: item.Name, Where: item.Where, Pending: true})
	}
	return table
}

// measure walks every item and keeps the table, closing done after.
func (i *Inventory) measure(done chan struct{}) {
	table := Table{Total: Row{Name: "Total"}}
	for _, item := range i.Items() {
		row := Row{Name: item.Name, Where: item.Where}
		for _, path := range item.Paths {
			size, files := Count(path)
			row.Size += size
			row.Files += files
		}
		table.Total.Size += row.Size
		table.Total.Files += row.Files
		table.Rows = append(table.Rows, row)
	}
	i.mu.Lock()
	i.table, i.at, i.running = table, time.Now(), nil
	i.mu.Unlock()
	close(done)
}

// Count returns the bytes and the number of the regular files at a path.
func Count(path string) (size, files int64) {
	_ = filepath.WalkDir(path, func(_ string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return nil
		}
		size += info.Size()
		files++
		return nil
	})
	return size, files
}
