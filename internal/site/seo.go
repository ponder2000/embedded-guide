package site

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"html/template"
	"net/url"
	"strings"
)

// Options configures deployment metadata. SiteURL is the public origin only;
// BasePath supplies any project prefix. An empty origin omits absolute SEO URLs.
type Options struct {
	BasePath     string
	SiteURL      string
	VisitorBadge bool
}

type metadata struct {
	Title, Description, Canonical, Image, Topbar, BadgeURL string
	NoIndex                                                bool
	StructuredData                                         template.JS
}

func plainTitle(title string) string {
	return html.UnescapeString(strings.NewReplacer("<em>", "", "</em>", "").Replace(title))
}

func siteOrigin(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	u, err := url.Parse(value)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || (u.Path != "" && u.Path != "/") || u.RawPath != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Opaque != "" {
		return "", fmt.Errorf("site URL must be an HTTPS origin without a path, query, or fragment: %q", value)
	}
	return strings.TrimSuffix(u.String(), "/"), nil
}

func pageMetadata(root, path, title, description, topbar string, badge bool) metadata {
	m := metadata{Title: title, Description: description, Topbar: topbar}
	if root == "" {
		return m
	}
	m.Canonical = root + path
	m.Image = root + "/static/social-card.png"
	if badge {
		// One shared namespace across all chapters. No unique-person claim.
		m.BadgeURL = "https://hits.sh/" + strings.TrimPrefix(root, "https://") + ".svg?label=visits&color=b8521e&labelColor=292720&style=flat"
	}
	kind := "LearningResource"
	if path == "/" {
		kind = "WebSite"
	}
	data, _ := json.Marshal(map[string]any{
		"@context": "https://schema.org", "@type": kind,
		"name": title, "description": description, "url": m.Canonical,
		"inLanguage": "en", "isAccessibleForFree": true,
		"author": map[string]string{"@type": "Person", "name": "Jay", "url": "https://x.com/chotathanos"},
	})
	// json.Marshal escapes HTML delimiters; only first-party metadata enters here.
	m.StructuredData = template.JS(data)
	return m
}

func discoveryFiles(root string, paths []string) (map[string][]byte, error) {
	type entry struct {
		Location string `xml:"loc"`
	}
	doc := struct {
		XMLName xml.Name `xml:"urlset"`
		XMLNS   string   `xml:"xmlns,attr"`
		URLs    []entry  `xml:"url"`
	}{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	for _, path := range paths {
		doc.URLs = append(doc.URLs, entry{root + path})
	}
	data, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}
	return map[string][]byte{
		"sitemap.xml": append([]byte(xml.Header), data...),
		"robots.txt":  []byte("User-agent: *\nAllow: /\n\nSitemap: " + root + "/sitemap.xml\n"),
	}, nil
}
