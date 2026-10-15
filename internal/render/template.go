// Package render renders the guide's templates without an HTTP runtime.
package render

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
)

// Renderer holds a parsed template set.
type Renderer struct {
	templates *template.Template
}

// New parses every .html file in the given filesystem into one template set.
func New(files fs.FS, basePath string) (*Renderer, error) {
	r := &Renderer{}

	funcs := template.FuncMap{
		"url": func(path string) string { return basePath + path },
		// safeHTML lets trusted markup render without escaping. Only ever
		// applied to first-party content (our own titles/SVG).
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		// seq returns [1..n], handy for rendering step dots in templates.
		"seq": func(n int) []int {
			out := make([]int, n)
			for i := range out {
				out[i] = i + 1
			}
			return out
		},
		// partial executes a named template (chosen at runtime, e.g. the
		// chapter's partial) and returns its HTML. {{template}} can't take a
		// variable name, so chapters use {{ partial .Chapter.Template . }}.
		"partial": func(name string, data any) (template.HTML, error) {
			var buf bytes.Buffer
			if err := r.templates.ExecuteTemplate(&buf, name, data); err != nil {
				return "", fmt.Errorf("partial %q: %w", name, err)
			}
			return template.HTML(buf.String()), nil
		},
	}

	tmpl := template.New("").Funcs(funcs)
	tmpl, err := tmpl.ParseFS(files,
		"templates/*.html",
		"templates/chapters/*.html",
	)
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	r.templates = tmpl
	return r, nil
}

// Render executes a top-level template.
func (r *Renderer) Render(w io.Writer, name string, data any) error {
	return r.templates.ExecuteTemplate(w, name, data)
}
