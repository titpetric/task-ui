package view

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"sync"
)

//go:embed all:assets
var assets embed.FS

// AssetPrefix is the path the server serves the assets under.
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

// AssetVersion is a hash of the embedded assets, for the ?v= of their
// URLs: the URL changes when a file does, so a browser that keeps them for
// an hour never keeps an old one past a deploy. The build version could
// not do it, a `go install` build is "dev" every time.
func AssetVersion() string {
	return assetVersion()
}
