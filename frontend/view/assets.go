package view

import (
	"io/fs"
	"sync"

	"github.com/titpetric/task-ui/internal/assets"
	"github.com/titpetric/task-ui/phpfront"
)

// AssetPrefix is the path the server serves the assets under.
//
// TODO: delete, replaced with phpscript, the theme serves them at
// internal/assets.Prefix.
const AssetPrefix = "/static/"

// theme is the static tree over the embedded theme, read once.
var theme = sync.OnceValue(func() *assets.Tree {
	return assets.New(map[string]fs.FS{assets.Theme: phpfront.ThemeAssets()})
})

// AssetsFS returns the theme's static files.
func AssetsFS() fs.FS {
	return phpfront.ThemeAssets()
}

// AssetVersion is the hash of the assets that their URLs carry as ?v=.
func AssetVersion() string {
	return theme().Version()
}

// Stylesheet is task-ui.css as served, its imports versioned.
func Stylesheet() []byte {
	return theme().CSS()
}
