//go:build embed_web

package web

import (
	"embed"
	"io/fs"
)

//go:embed static/dist
var embeddedFiles embed.FS

// StaticFiles returns the embedded static filesystem.
func StaticFiles() fs.FS {
	return embeddedFiles
}
