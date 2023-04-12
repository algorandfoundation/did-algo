package ui

import (
	"embed"
	"io/fs"
)

// AppContents contains a static build of the local graphical
// client.
var AppContents fs.FS

//go:embed local-app/dist
var dist embed.FS

func init() {
	AppContents, _ = fs.Sub(dist, "local-app/dist")
}
