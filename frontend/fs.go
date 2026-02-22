package frontend

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var FS embed.FS

// SubFS returns a sub-filesystem rooted at the given directory within FS.
func SubFS(dir string) (fs.FS, error) {
	return fs.Sub(FS, dir)
}
