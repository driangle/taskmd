//go:build !embed_web

package web

import (
	"io/fs"
	"testing/fstest"
)

// StaticFiles returns an empty filesystem when web assets are not embedded.
func StaticFiles() fs.FS {
	return fstest.MapFS{}
}
