package preview_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jaysaha/embedded-visual-guide/internal/content"
	"github.com/jaysaha/embedded-visual-guide/internal/preview"
	"github.com/jaysaha/embedded-visual-guide/internal/site"
)

func TestPreview(t *testing.T) {
	for _, base := range []string{"", "/visual-guide", "/nested/project"} {
		t.Run(base, func(t *testing.T) {
			files, err := site.Generate(base)
			if err != nil {
				t.Fatal(err)
			}
			dir, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			if err := site.Write(dir, files); err != nil {
				t.Fatal(err)
			}
			h, err := preview.New(dir, base)
			if err != nil {
				t.Fatal(err)
			}
			get := func(path string) *httptest.ResponseRecorder {
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
				return rec
			}
			if rec := get(base + "/"); rec.Code != 200 || rec.Body.String() != string(files["index.html"]) {
				t.Fatal("home differs from generated page")
			}
			for _, ch := range content.Chapters {
				url := base + "/chapter/" + ch.Slug
				if rec := get(url + "/"); rec.Code != 200 || rec.Body.String() != string(files["chapter/"+ch.Slug+"/index.html"]) {
					t.Errorf("chapter %s failed: %d", ch.Slug, rec.Code)
				}
				if rec := get(url); rec.Code != 301 || rec.Header().Get("Location") != ch.Slug+"/" {
					t.Errorf("chapter slash redirect: %d %q", rec.Code, rec.Header().Get("Location"))
				}
			}
			for _, name := range []string{"static/css/app.css", "static/js/app.js", "static/js/alpine.min.js", "404.html"} {
				if rec := get(base + "/" + name); rec.Code != 200 || rec.Body.String() != string(files[name]) {
					t.Errorf("file %s failed", name)
				}
			}
			for _, suffix := range []string{"missing", "chapter/not-real/", "static/", ".visual-guide-manifest.json", "../index.html"} {
				rec := get(base + "/" + suffix)
				if rec.Code != 404 || rec.Body.String() != string(files["404.html"]) {
					t.Errorf("missing %s got %d", suffix, rec.Code)
				}
			}
			if base != "" {
				if rec := get(base); rec.Code != 301 || rec.Header().Get("Location") != base+"/" {
					t.Error("missing base slash redirect")
				}
				if rec := get("/"); rec.Code != 404 {
					t.Error("served site outside base prefix")
				}
				if rec := get(base + "-other/"); rec.Code != 404 {
					t.Error("prefix collision")
				}
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodHead, base+"/missing", nil))
			if rec.Code != 404 || rec.Body.Len() != 0 {
				t.Error("HEAD 404 has body or wrong status")
			}
			rec = httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, base+"/", strings.NewReader("x")))
			if rec.Code != 405 {
				t.Error("POST should be rejected")
			}
		})
	}
}

func TestPreviewRequiresBuild(t *testing.T) {
	if _, err := preview.New(t.TempDir(), ""); err == nil {
		t.Fatal("accepted empty output")
	}
	if _, err := preview.New(t.TempDir(), "https://evil"); err == nil {
		t.Fatal("accepted invalid base")
	}
}
