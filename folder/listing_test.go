package folder

import (
	"path/filepath"
	"slices"
	"testing"
)

// A listing has the folders first, then the files, each by name
// regardless of case; dot entries stay out.
func TestListing(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "b.txt"), "b")
	write(t, filepath.Join(dir, "Makefile"), "all:")
	write(t, filepath.Join(dir, "zeta", "x"), "x")
	write(t, filepath.Join(dir, "Alpha", "y"), "y")
	write(t, filepath.Join(dir, ".git", "HEAD"), "ref")
	write(t, filepath.Join(dir, ".env"), "A=1")

	names, dirs := []string{}, []bool{}
	for _, item := range Listing(dir) {
		names = append(names, item.Name)
		dirs = append(dirs, item.Dir)
	}
	if want := []string{"Alpha", "zeta", "b.txt", "Makefile"}; !slices.Equal(names, want) {
		t.Fatalf("listing: got %v, want %v", names, want)
	}
	if want := []bool{true, true, false, false}; !slices.Equal(dirs, want) {
		t.Fatalf("dirs: got %v, want %v", dirs, want)
	}
	if Listing(filepath.Join(dir, "nope")) != nil {
		t.Fatal("a missing directory lists something")
	}
}

// An item is a name, and whether the name links into a directory.
func TestItem(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "sub", "x"), "x")
	write(t, filepath.Join(dir, "chart.SVG"), "<svg/>")
	items := Listing(dir)
	if len(items) != 2 || items[0].Name != "sub" || !items[0].Dir || items[0].Size != 0 || items[0].ModTime.IsZero() {
		t.Fatalf("items: %+v", items)
	}
	if items[1].Size != 6 {
		t.Fatalf("size: %+v", items[1])
	}
	if items[1].Preview() {
		t.Fatal("a file without a link has no preview")
	}
	items[1].Href = "/x"
	if !items[1].Preview() {
		t.Fatal("an svg with a link has a preview")
	}
	if (Item{Name: "a.txt", Href: "/a"}).Preview() || (Item{Name: "d.png", Dir: true, Href: "/d"}).Preview() {
		t.Fatal("text and directories have no preview")
	}
}
