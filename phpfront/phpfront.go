// Package phpfront serves task-ui's pages from a phpscript application: a
// theme of components, a module per view, and the task-ui data model as PHP
// classes over the instance's database.
//
// The tree under app/ is the application, embedded and extracted into the data
// folder on start, because composer's autoloader reads a real directory and
// MiniTPL writes compiled templates into one. TASK_UI_PHP_FRONT names a
// checkout instead.
//
// Routing is the phpscript annotation scanner. The @route paths carry no
// prefix; Go mounts the router where the instance serves it and chi strips
// that, so /development is one Go line and no template knows of it.
//
// Static files never reach the PHP: assets.go serves them from Go, so an
// @route is always a page or an endpoint.
package phpfront

import (
	"embed"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

//go:embed all:app
var tree embed.FS

// Dir is where the application is extracted, under the data folder.
const Dir = "phpfront"

// Env names a checkout served in place of the extracted tree.
const Env = "TASK_UI_PHP_FRONT"

// App returns the embedded tree, rooted at app/.
func App() fs.FS {
	return sub("app")
}

// ThemeAssets returns the theme's static files out of the embedded tree.
func ThemeAssets() fs.FS {
	return sub(path.Join("app", "theme", "assets"))
}

// sub roots the embedded tree at dir.
func sub(dir string) fs.FS {
	at, err := fs.Sub(tree, dir)
	if err != nil {
		panic(err)
	}
	return at
}

// Extract writes the embedded tree under dir and returns the root to serve.
// A shipped file is replaced; what the instance wrote beside it stays, so the
// compiled templates survive. Env short-circuits the whole thing.
func Extract(dir string) (string, error) {
	if checkout := os.Getenv(Env); checkout != "" {
		return filepath.Abs(checkout)
	}
	root := filepath.Join(dir, Dir)
	err := fs.WalkDir(App(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(root, filepath.FromSlash(name))
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(App(), name)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		return "", err
	}
	return root, nil
}
