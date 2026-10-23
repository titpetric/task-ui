package view

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"regexp"
	"sync"
)

//go:embed all:assets
var assets embed.FS

// AssetPrefix is the path the server serves the assets under.
const AssetPrefix = "/static/"

// AssetsFS returns the embedded assets: the stylesheet, the script, the icon and the vendored terminal.
// The server serves any file dropped in the assets folder.
func AssetsFS() fs.FS {
	sub, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}
	return sub
}

// assetVersion hashes the embedded assets: the URL changes when a file
// does, so a browser that keeps them for an hour never keeps an old one
// past a deploy. The build version cannot do it, a `go install` build is
// "dev" every time.
var assetVersion = sync.OnceValue(func() string {
	h := sha256.New()
	_ = fs.WalkDir(assets, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := assets.ReadFile(path)
		if err != nil {
			return err
		}
		h.Write([]byte(path))
		h.Write(data)
		return nil
	})
	return hex.EncodeToString(h.Sum(nil))[:12]
})

// AssetVersion is the hash of the embedded assets that their URLs carry as ?v=.
func AssetVersion() string {
	return assetVersion()
}

// importURL matches the url of an @import in the stylesheet.
var importURL = regexp.MustCompile(`@import url\("(css/[^"?]+\.css)"\)`)

// stylesheet versions every import of task-ui.css, so a deploy that
// changes one file changes that import's URL with it.
var stylesheet = sync.OnceValue(func() []byte {
	data, err := assets.ReadFile("assets/task-ui.css")
	if err != nil {
		panic(err)
	}
	return importURL.ReplaceAll(data, []byte(`@import url("$1?v=`+AssetVersion()+`")`))
})

// Stylesheet is task-ui.css as served, each import of css/ carrying the ?v= of AssetVersion.
func Stylesheet() []byte {
	return stylesheet()
}
