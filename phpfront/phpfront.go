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
// that, so the mount is one Go line, app.FrontBase, and no template knows of
// it. It is the empty string: these pages are what the instance serves, and
// the Go pages being replaced are the ones under a prefix now.
//
// Static files never reach the PHP: assets.go serves them from Go, so an
// @route is always a page or an endpoint.
//
// The plugins are the other half: a module is a directory with a modinfo.php
// in it, under /m/<vendor>/<name>, and a Manager finds the modules under its
// roots, runs each modinfo.php once to learn what it registers and keeps a
// pool of runtimes per module to serve requests from. Four host symbols reach
// that PHP: the Task\Module constructor a modinfo.php builds, Task::register,
// Task::migrate, which applies the module's own schema, and Task::dispatch,
// which starts a Claude run. Which module is calling sits on the runtime's
// context, so the bindings are the same for every module, and the handler
// rewrites no path: REQUEST_URI holds what the browser asked for, prefix and
// all. Writing one is docs/modules.md.
package phpfront

import (
	"embed"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
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

// Migrations returns the embedded module tree, never the checkout Env names.
func Migrations() fs.FS {
	return sub(path.Join("app", "modules"))
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
// The tree is the binary's: a file this build does not ship is deleted, and a
// module dropped from it answers its routes forever otherwise.
//
// Env short-circuits the whole thing, so a checkout is never pruned.
func Extract(dir string) (string, error) {
	if checkout := os.Getenv(Env); checkout != "" {
		return filepath.Abs(checkout)
	}
	root := filepath.Join(dir, Dir)
	shipped := map[string]bool{}
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
		shipped[name] = true
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		return "", err
	}
	if err := prune(root, shipped); err != nil {
		return "", err
	}
	return root, nil
}

// prune deletes every file under root that shipped does not name.
func prune(root string, shipped map[string]bool) error {
	var dirs []string
	err := filepath.WalkDir(root, func(target string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name, err := filepath.Rel(root, target)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			dirs = append(dirs, target)
			return nil
		}
		if shipped[filepath.ToSlash(name)] {
			return nil
		}
		return os.Remove(target)
	})
	if err != nil {
		return err
	}
	// Deepest first, and a directory still holding something refuses.
	slices.Reverse(dirs)
	for _, dir := range dirs {
		if dir != root {
			os.Remove(dir)
		}
	}
	return nil
}
