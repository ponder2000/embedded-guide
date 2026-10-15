// Package preview serves generated files, never templates or application routes.
package preview

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/jaysaha/embedded-visual-guide/internal/site"
)

// New creates a static preview with directory indexes, redirects, and a branded 404.
func New(directory, base string) (http.Handler, error) {
	base, err := site.BasePath(base)
	if err != nil {
		return nil, err
	}
	files := os.DirFS(directory)
	for _, name := range []string{"index.html", "404.html"} {
		if _, err := fs.Stat(files, name); err != nil {
			return nil, fmt.Errorf("build the site first: %w", err)
		}
	}
	notFound, err := fs.ReadFile(files, "404.html")
	if err != nil {
		return nil, err
	}
	static := http.StripPrefix(base, http.FileServer(http.FS(files)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		if base != "" && r.URL.Path == base {
			http.Redirect(w, r, base+"/", http.StatusMovedPermanently)
			return
		}
		name := strings.TrimPrefix(r.URL.Path, base+"/")
		valid := strings.HasPrefix(r.URL.Path, base+"/")
		for _, segment := range strings.Split(name, "/") {
			if strings.HasPrefix(segment, ".") || strings.Contains(segment, "\\") {
				valid = false
			}
		}
		if name == "" {
			name = "index.html"
		}
		info, statErr := fs.Stat(files, strings.TrimSuffix(name, "/"))
		if valid && statErr == nil {
			if !info.IsDir() {
				static.ServeHTTP(w, r)
				return
			}
			if index, err := fs.Stat(files, path.Join(name, "index.html")); err == nil && !index.IsDir() {
				static.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		if r.Method != http.MethodHead {
			_, _ = w.Write(notFound)
		}
	}), nil
}
