package folder

import (
	"path/filepath"
	"slices"
	"testing"
)

// A listing has the folders first, then the files, each by name
// regardless of case; dot entries are in it, the sealed ones marked.
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
	if want := []string{".git", "Alpha", "zeta", ".env", "b.txt", "Makefile"}; !slices.Equal(names, want) {
		t.Fatalf("listing: got %v, want %v", names, want)
	}
	if want := []bool{true, true, true, false, false, false}; !slices.Equal(dirs, want) {
		t.Fatalf("dirs: got %v, want %v", dirs, want)
	}
	sealed := []bool{}
	for _, item := range Listing(dir) {
		sealed = append(sealed, item.Sealed)
	}
	if want := []bool{true, false, false, true, false, false}; !slices.Equal(sealed, want) {
		t.Fatalf("sealed: got %v, want %v", sealed, want)
	}
	if Listing(filepath.Join(dir, "nope")) != nil {
		t.Fatal("a missing directory lists something")
	}
}
