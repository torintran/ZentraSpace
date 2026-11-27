package app

import (
	"embed"
	"io/fs"
)

//go:embed web/assets/icon.svg
var webFiles embed.FS

func webAssets() fs.FS {
	assets, err := fs.Sub(webFiles, "web")
	if err != nil {
		panic(err)
	}
	return assets
}
