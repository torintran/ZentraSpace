package main

import (
	"embed"
	"io/fs"
)

//go:embed web/app.html web/assets
var webFiles embed.FS

func webAssets() fs.FS {
	assets, err := fs.Sub(webFiles, "web")
	if err != nil {
		panic(err)
	}
	return assets
}
