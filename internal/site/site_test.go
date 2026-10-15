package site_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	guide "github.com/jaysaha/embedded-visual-guide"
	"github.com/jaysaha/embedded-visual-guide/internal/content"
	"github.com/jaysaha/embedded-visual-guide/internal/site"
)

func TestBasePath(t *testing.T) {
	for input, want := range map[string]string{"": "", "/": "", "/guide": "/guide", "/guide/": "/guide", "/org/guide/": "/org/guide"} {
		got, err := site.BasePath(input)
		if err != nil || got != want {
			t.Errorf("BasePath(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
	for _, input := range []string{"guide", "https://example.com", "//evil", "/../guide", "/a/./b", "/a//b", "/a//", "/a?b", "/a#b", "/a%20b", "/a\\b", "/a b"} {
		if _, err := site.Generate(input); err == nil {
			t.Errorf("accepted invalid base path %q", input)
		}
	}
}

func TestGeneratedSite(t *testing.T) {
	for _, base := range []string{"", "/", "/visual-guide/", "/nested/project"} {
		t.Run(base, func(t *testing.T) {
			prefix, _ := site.BasePath(base)
			files, err := site.Generate(base)
			if err != nil {
				t.Fatal(err)
			}
			home := string(files["index.html"])
			for _, ch := range content.Chapters {
				url := prefix + "/chapter/" + ch.Slug + "/"
				if !strings.Contains(home, `href="`+url+`"`) {
					t.Errorf("missing TOC link %s", url)
				}
				body := string(files["chapter/"+ch.Slug+"/index.html"])
				if !strings.Contains(body, `class="section-number"`) {
					t.Errorf("%s missing chapter partial", ch.Slug)
				}
				if ch.Interactive && !strings.Contains(body, "x-data") {
					t.Errorf("%s lost Alpine interaction", ch.Slug)
				}
				prev, next := content.Neighbours(ch.Slug)
				for _, neighbour := range []content.Chapter{prev, next} {
					if neighbour.Slug != "" && !strings.Contains(body, `href="`+prefix+"/chapter/"+neighbour.Slug+`/"`) {
						t.Errorf("%s missing navigation to %s", ch.Slug, neighbour.Slug)
					}
				}
			}
			if !strings.Contains(string(files["404.html"]), `href="`+prefix+`/"`) {
				t.Error("404 missing base-aware home link")
			}
			if _, ok := files[".nojekyll"]; !ok {
				t.Error("missing Pages marker")
			}
			// Every internal page/asset reference must resolve to generated output.
			links := regexp.MustCompile(`(?:href|src)="(/[^\"]*)"`)
			for name, body := range files {
				if !strings.HasSuffix(name, ".html") {
					continue
				}
				for _, match := range links.FindAllSubmatch(body, -1) {
					url := string(match[1])
					if !strings.HasPrefix(url, prefix+"/") {
						t.Errorf("%s has unprefixed URL %s", name, url)
						continue
					}
					key := strings.TrimPrefix(url, prefix+"/")
					if strings.Contains(key, "#") {
						key = strings.SplitN(key, "#", 2)[0]
					}
					if key == "" || strings.HasSuffix(key, "/") {
						key += "index.html"
					}
					if _, ok := files[key]; !ok {
						t.Errorf("%s links to missing %s", name, key)
					}
				}
			}
			assets, err := guide.Static()
			if err != nil {
				t.Fatal(err)
			}
			err = fs.WalkDir(assets, ".", func(path string, entry fs.DirEntry, err error) error {
				if err != nil || entry.IsDir() {
					return err
				}
				want, err := fs.ReadFile(assets, path)
				if err != nil {
					return err
				}
				if !bytes.Equal(files["static/"+path], want) {
					t.Errorf("asset changed or missing: %s", path)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(realTemp(t), "output")
			if err := site.Write(out, files); err != nil {
				t.Fatal(err)
			}
			if err := site.Write(out, files); err != nil {
				t.Fatalf("repeat build: %v", err)
			}
			for name, want := range files {
				got, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(name)))
				if err != nil || !bytes.Equal(got, want) {
					t.Errorf("written %s differs: %v", name, err)
				}
			}
		})
	}
}

func TestGuideLayout(t *testing.T) {
	for _, base := range []string{"", "/visual-guide"} {
		files, err := site.Generate(base)
		if err != nil {
			t.Fatal(err)
		}
		home := string(files["index.html"])
		for _, marker := range []string{"Why this guide exists", "Who it's for", "What you'll learn", `class="landing-hero"`} {
			if !strings.Contains(home, marker) {
				t.Errorf("home missing introduction: %s", marker)
			}
		}
		for _, marker := range []string{`class="circuit-board"`, `class="chapter-index"`, `data-answer="correct"`, `class="overall-progress"`} {
			if !strings.Contains(home, marker) {
				t.Errorf("home missing interactive feature %s", marker)
			}
		}
		if strings.Contains(home, `class="toc-list"`) {
			t.Error("home still renders the old chapter-list landing page")
		}
		pages := map[string]string{"index.html": ""}
		for _, chapter := range content.Chapters {
			pages["chapter/"+chapter.Slug+"/index.html"] = chapter.Slug
		}
		for name, active := range pages {
			body := string(files[name])
			for _, marker := range []string{`class="guide-layout"`, `class="chapter-menu" open`, `aria-label="Chapters"`, `id="main-content"`, `href="#main-content"`} {
				if !strings.Contains(body, marker) {
					t.Errorf("%s missing layout element %s", name, marker)
				}
			}
			if count := strings.Count(body, `aria-current="page"`); count != 1 {
				t.Errorf("%s has %d current pages; want 1", name, count)
			}
			activeURL := base + "/"
			if active != "" {
				activeURL += "chapter/" + active + "/"
			}
			if !strings.Contains(body, `href="`+activeURL+`" aria-current="page"`) {
				t.Errorf("%s missing correct active link", name)
			}
			for _, chapter := range content.Chapters {
				if !strings.Contains(body, `href="`+base+"/chapter/"+chapter.Slug+`/"`) {
					t.Errorf("%s missing sidebar chapter %s", name, chapter.Slug)
				}
			}
			for _, marker := range []string{`id="guide-sidebar"`, `aria-controls="guide-sidebar"`, `id="theme-choice"`, `id="sound-choice"`, `id="tracking-choice"`, `id="reset-progress"`, `/static/css/experience.css`, `/static/favicon.svg`} {
				if !strings.Contains(body, marker) {
					t.Errorf("%s missing experience feature %s", name, marker)
				}
			}
		}
		for _, slug := range []string{"boot-sequence", "clocks", "interrupts", "timers-pwm", "uart"} {
			body := string(files["chapter/"+slug+"/index.html"])
			if !strings.Contains(body, `data-simulation=`) || !strings.Contains(body, `/static/js/simulations.js`) {
				t.Errorf("%s missing simulation player", slug)
			}
		}
	}
}

// macOS's default temp path may itself contain the /var symlink.
func realTemp(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestOutputSafety(t *testing.T) {
	t.Run("hard links are replaced not truncated", func(t *testing.T) {
		out := realTemp(t)
		if err := site.Write(out, map[string][]byte{"index.html": []byte("original")}); err != nil {
			t.Fatal(err)
		}
		outside := filepath.Join(realTemp(t), "saved.html")
		if err := os.Link(filepath.Join(out, "index.html"), outside); err != nil {
			t.Fatal(err)
		}
		if err := site.Write(out, map[string][]byte{"index.html": []byte("updated")}); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(outside)
		if err != nil || string(got) != "original" {
			t.Fatal("changed external hard link")
		}
	})
	t.Run("unrecognized file in previous build", func(t *testing.T) {
		out := realTemp(t)
		if err := site.Write(out, map[string][]byte{"index.html": []byte("old")}); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, "notes.txt"), []byte("user work"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := site.Write(out, map[string][]byte{"index.html": []byte("new")}); err == nil {
			t.Fatal("accepted unknown file")
		}
		got, _ := os.ReadFile(filepath.Join(out, "index.html"))
		if string(got) != "old" {
			t.Fatal("updated output before completing validation")
		}
	})
	t.Run("manifest cannot remove directories", func(t *testing.T) {
		out := realTemp(t)
		if err := os.Mkdir(filepath.Join(out, "user-directory"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, ".visual-guide-manifest.json"), []byte(`{"user-directory":"hash"}`), 0644); err != nil {
			t.Fatal(err)
		}
		if err := site.Write(out, nil); err == nil {
			t.Fatal("accepted directory in manifest")
		}
		if info, err := os.Stat(filepath.Join(out, "user-directory")); err != nil || !info.IsDir() {
			t.Fatal("removed user directory")
		}
	})
	t.Run("unknown file", func(t *testing.T) {
		out := realTemp(t)
		path := filepath.Join(out, "notes.txt")
		if err := os.WriteFile(path, []byte("user work"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := site.Write(out, map[string][]byte{"index.html": []byte("generated")}); err == nil {
			t.Fatal("accepted non-build directory")
		}
		got, _ := os.ReadFile(path)
		if string(got) != "user work" {
			t.Fatal("overwrote user data")
		}
		if _, err := os.Stat(filepath.Join(out, "index.html")); !os.IsNotExist(err) {
			t.Fatal("wrote before validation")
		}
	})
	t.Run("modified generated file", func(t *testing.T) {
		out := realTemp(t)
		files := map[string][]byte{"index.html": []byte("generated")}
		if err := site.Write(out, files); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, "index.html"), []byte("user edit"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := site.Write(out, files); err == nil {
			t.Fatal("overwrote user edits")
		}
		got, _ := os.ReadFile(filepath.Join(out, "index.html"))
		if string(got) != "user edit" {
			t.Fatal("modified file was changed")
		}
	})
	t.Run("stale owned files", func(t *testing.T) {
		out := realTemp(t)
		if err := site.Write(out, map[string][]byte{"old/index.html": []byte("old")}); err != nil {
			t.Fatal(err)
		}
		if err := site.Write(out, map[string][]byte{"index.html": []byte("new")}); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(out, "old/index.html")); !os.IsNotExist(err) {
			t.Fatal("stale generated file remains")
		}
	})
	for _, kind := range []string{"root", "child", "ancestor"} {
		t.Run("symlink "+kind, func(t *testing.T) {
			root := realTemp(t)
			target := realTemp(t)
			link := filepath.Join(root, "link")
			if err := os.Symlink(target, link); err != nil {
				t.Fatal(err)
			}
			out := root
			if kind == "root" {
				out = link
			}
			if kind == "ancestor" {
				out = filepath.Join(link, "output")
			}
			if err := site.Write(out, map[string][]byte{"index.html": []byte("new")}); err == nil {
				t.Fatal("accepted symlink")
			}
			entries, err := os.ReadDir(target)
			if err != nil || len(entries) != 0 {
				t.Fatal("modified symlink target")
			}
		})
	}
	t.Run("unsafe paths", func(t *testing.T) {
		for _, name := range []string{"../escape", "/absolute", ".", "a/../b", ".visual-guide-manifest.json"} {
			if err := site.Write(realTemp(t), map[string][]byte{name: []byte("bad")}); err == nil {
				t.Errorf("accepted %s", name)
			}
		}
	})
	t.Run("unsafe manifest", func(t *testing.T) {
		out := realTemp(t)
		if err := os.WriteFile(filepath.Join(out, ".visual-guide-manifest.json"), []byte(`{"../escape":"hash"}`), 0644); err != nil {
			t.Fatal(err)
		}
		if err := site.Write(out, nil); err == nil {
			t.Fatal("accepted unsafe manifest")
		}
	})
}
