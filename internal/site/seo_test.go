package site_test

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"image/png"
	"regexp"
	"strings"
	"testing"

	"github.com/jaysaha/embedded-visual-guide/internal/content"
	"github.com/jaysaha/embedded-visual-guide/internal/site"
)

func TestProductionMetadata(t *testing.T) {
	for _, base := range []string{"", "/guide"} {
		files, err := site.GenerateWithOptions(site.Options{BasePath: base, SiteURL: "https://example.com/", VisitorBadge: true})
		if err != nil {
			t.Fatal(err)
		}
		root := "https://example.com" + base
		paths := map[string]string{"index.html": "/"}
		for _, ch := range content.Chapters {
			paths["chapter/"+ch.Slug+"/index.html"] = "/chapter/" + ch.Slug + "/"
		}
		titles, descriptions := map[string]bool{}, map[string]bool{}
		for file, path := range paths {
			body := string(files[file])
			for _, marker := range []string{
				`rel="canonical" href="` + root + path + `"`,
				`property="og:image" content="` + root + `/static/social-card.png"`,
				`name="twitter:card" content="summary_large_image"`,
				`https://hits.sh/example.com` + base + `.svg?`,
				`referrerpolicy="no-referrer"`,
			} {
				if !strings.Contains(body, marker) {
					t.Errorf("%s missing %s", file, marker)
				}
			}
			for pattern, seen := range map[string]map[string]bool{
				`<title>(.*?)</title>`:                        titles,
				`<meta name="description" content="([^"]*)">`: descriptions,
			} {
				match := regexp.MustCompile(pattern).FindStringSubmatch(body)
				if len(match) != 2 || seen[match[1]] {
					t.Fatalf("%s missing or duplicate metadata: %s", file, pattern)
				}
				seen[match[1]] = true
			}
			match := regexp.MustCompile(`<script type="application/ld\+json">(.*?)</script>`).FindStringSubmatch(body)
			var structured map[string]any
			if len(match) != 2 || json.Unmarshal([]byte(match[1]), &structured) != nil || structured["url"] != root+path {
				t.Errorf("%s invalid structured metadata", file)
			}
		}
		var sitemap struct {
			URLs []struct {
				Loc string `xml:"loc"`
			} `xml:"url"`
		}
		if err := xml.Unmarshal(files["sitemap.xml"], &sitemap); err != nil {
			t.Fatal(err)
		}
		if len(sitemap.URLs) != len(paths) {
			t.Fatal("sitemap should contain exactly home and all chapters")
		}
		for _, entry := range sitemap.URLs {
			if !strings.HasPrefix(entry.Loc, root+"/") || strings.Contains(entry.Loc, "404") {
				t.Errorf("invalid sitemap URL %s", entry.Loc)
			}
		}
		if !strings.Contains(string(files["robots.txt"]), "Sitemap: "+root+"/sitemap.xml") {
			t.Error("robots missing deployment-aware sitemap")
		}
		errorPage := string(files["404.html"])
		if !strings.Contains(errorPage, `content="noindex, follow"`) || strings.Contains(errorPage, `rel="canonical"`) || strings.Contains(errorPage, "hits.sh/") {
			t.Error("404 must be noindex without canonical or counter")
		}
		image, err := png.DecodeConfig(bytes.NewReader(files["static/social-card.png"]))
		if err != nil || image.Width != 1200 || image.Height != 630 {
			t.Errorf("invalid social image: %+v, %v", image, err)
		}
	}
}

func TestSEOConfiguration(t *testing.T) {
	for _, origin := range []string{"http://example.com", "https://example.com/guide", "https://u:p@example.com", "https://example.com?a=b", "https://example.com?", "https://example.com/%2F", "https://example.com#frag", "//example.com", "https://"} {
		if _, err := site.GenerateWithOptions(site.Options{SiteURL: origin}); err == nil {
			t.Errorf("accepted invalid site origin %q", origin)
		}
	}
	if _, err := site.GenerateWithOptions(site.Options{VisitorBadge: true}); err == nil {
		t.Error("counter without public URL should fail")
	}
	files, err := site.Generate("")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"sitemap.xml", "robots.txt"} {
		if _, exists := files[name]; exists {
			t.Errorf("local build should omit %s", name)
		}
	}
	if strings.Contains(string(files["index.html"]), "hits.sh/") || strings.Contains(string(files["index.html"]), `rel="canonical"`) {
		t.Error("local builds should not claim a production URL or count visits")
	}
}
