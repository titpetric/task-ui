// Package phpscripts holds the modules that ship with the instance, as a PHP
// source tree embedded in the binary.
//
// The tree is not served from the binary. Extract writes it to the instance's
// modules directory on start, and the Manager walks that directory with the
// rest: composer's autoloader reads a vendor directory, the filesystem
// bindings are rooted at a real path, and the compile directory has to be
// writable, none of which an embedded filesystem answers.
// docs/todo-correction.md records why.
package phpscripts

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:modules
var tree embed.FS

// Modules returns the embedded source tree, rooted at the modules directory,
// so a path in it reads as <vendor>/<plugin>/modinfo.php.
func Modules() fs.FS {
	sub, err := fs.Sub(tree, "modules")
	if err != nil {
		// modules is embedded above, so a failure here is a build that
		// cannot happen.
		panic(err)
	}
	return sub
}

// Extract writes the embedded tree under dir, one directory per module.
//
// A file already there is replaced, so an upgrade ships a changed module and
// the next start applies it. Anything the instance wrote beside the shipped
// files stays: a module's compile directory survives, and so does a module a
// checkout put there.
func Extract(dir string) error {
	return fs.WalkDir(Modules(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(dir, filepath.FromSlash(name))
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(Modules(), name)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, mode(name))
	})
}

// mode is the permission a shipped file lands with: a PHP script is read, not
// run as a program, so nothing in the tree is executable.
func mode(name string) os.FileMode {
	if filepath.Ext(name) == ".sh" {
		return 0o755
	}
	return 0o644
}
