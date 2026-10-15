// Package guide holds the embedded web assets (templates + static files).
//
// The embed directives must live in a package whose directory is an ancestor
// of web/, so they live here at the module root rather than in cmd/build.
package guide

import (
	"embed"
	"io/fs"
)

//go:embed web/templates/*.html web/templates/chapters/*.html
var templatesFS embed.FS

//go:embed web/static
var staticFS embed.FS

// Templates returns the template filesystem rooted at "templates/".
func Templates() (fs.FS, error) { return fs.Sub(templatesFS, "web") }

// Static returns the static-asset filesystem rooted at the static dir.
func Static() (fs.FS, error) { return fs.Sub(staticFS, "web/static") }
