// Package site builds the static guide from embedded templates and assets.
package site

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"

	guide "github.com/jaysaha/embedded-visual-guide"
	"github.com/jaysaha/embedded-visual-guide/internal/content"
	"github.com/jaysaha/embedded-visual-guide/internal/render"
)

// BasePath normalizes a root-relative deployment prefix, without a final slash.
// URL metacharacters, dot segments, and empty segments are deliberately rejected.
func BasePath(value string) (string, error) {
	if value == "" || value == "/" {
		return "", nil
	}
	if !strings.HasPrefix(value, "/") {
		return "", fmt.Errorf("base path must start with /: %q", value)
	}
	value = strings.TrimSuffix(value, "/")
	for _, segment := range strings.Split(value[1:], "/") {
		if segment == "" || segment == "." || segment == ".." {
			return "", fmt.Errorf("invalid base path %q", value)
		}
		for _, c := range segment {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c == '.') {
				return "", fmt.Errorf("invalid base path %q", value)
			}
		}
	}
	return value, nil
}

// Generate renders every page before touching output. The returned keys are
// relative to the deployment root (the base path is only used in URLs).
func Generate(base string) (map[string][]byte, error) {
	base, err := BasePath(base)
	if err != nil {
		return nil, err
	}
	templates, err := guide.Templates()
	if err != nil {
		return nil, err
	}
	r, err := render.New(templates, base)
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{".nojekyll": {}}
	sectionCounts := make(map[string]int, len(content.Chapters))
	for _, chapter := range content.Chapters {
		var partial bytes.Buffer
		if err := r.Render(&partial, chapter.Template, nil); err != nil {
			return nil, fmt.Errorf("count sections in %s: %w", chapter.Slug, err)
		}
		sectionCounts[chapter.Slug] = strings.Count(partial.String(), "<section>")
	}
	page := func(path, name string, data any) error {
		var b bytes.Buffer
		if err := r.Render(&b, name, data); err != nil {
			return fmt.Errorf("render %s: %w", path, err)
		}
		files[path] = b.Bytes()
		return nil
	}
	if err := page("index.html", "home.html", struct {
		Parts       []content.Part
		ChapterNum  int
		CurrentSlug string
		SectionCounts map[string]int
	}{content.Parts(), len(content.Chapters), "", sectionCounts}); err != nil {
		return nil, err
	}
	for i, ch := range content.Chapters {
		if !fs.ValidPath(ch.Slug) || strings.Contains(ch.Slug, "/") {
			return nil, fmt.Errorf("invalid chapter slug %q", ch.Slug)
		}
		prev, next := content.Neighbours(ch.Slug)
		data := struct {
			Chapter       content.Chapter
			Number, Total int
			Prev, Next    content.Chapter
			Percent       int
			Parts         []content.Part
			CurrentSlug   string
			SectionCounts map[string]int
		}{ch, i + 1, len(content.Chapters), prev, next, (i + 1) * 100 / len(content.Chapters), content.Parts(), ch.Slug, sectionCounts}
		if err := page("chapter/"+ch.Slug+"/index.html", "chapter.html", data); err != nil {
			return nil, err
		}
	}
	if err := page("404.html", "error.html", map[string]any{"Code": 404}); err != nil {
		return nil, err
	}
	assets, err := guide.Static()
	if err != nil {
		return nil, err
	}
	err = fs.WalkDir(assets, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(assets, path)
		if err == nil {
			files["static/"+path] = data
		}
		return err
	})
	return files, err
}
