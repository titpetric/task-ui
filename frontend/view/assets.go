package view

import (
	"embed"
	"io/fs"
)

//go:embed all:assets
var assets embed.FS

// AssetPrefix is where the assets are served.
const AssetPrefix = "/static/"

// AssetsFS returns the embedded assets: the stylesheet, the script, the
// icon and the vendored terminal. Dropping a file in the folder is all it
// takes to serve it.
func AssetsFS() fs.FS {
	sub, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}
	return sub
}
