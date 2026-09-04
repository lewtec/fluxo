package web

import (
	"embed"
	"io/fs"
)

//go:embed all:static
var staticFS embed.FS

// Static returns the embedded static files (CSS, JS).
func Static() (fs.FS, error) {
	return fs.Sub(staticFS, "static")
}
