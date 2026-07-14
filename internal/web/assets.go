package web

import (
	"embed"
	"io/fs"
)

//go:embed static/*
var embeddedStatic embed.FS

func staticFiles() fs.FS {
	static, err := fs.Sub(embeddedStatic, "static")
	if err != nil {
		panic(err)
	}
	return static
}
