// Command build generates the deployable static site; no Go server is needed.
package main

import (
	"flag"
	"log"

	"github.com/jaysaha/embedded-visual-guide/internal/site"
)

func main() {
	output := flag.String("out", "dist", "output directory (empty or a previous verified build)")
	base := flag.String("base-path", "", "deployment URL prefix, e.g. /visual-guide/")
	origin := flag.String("site-url", "", "public HTTPS origin for SEO, e.g. https://example.com")
	badge := flag.Bool("visitor-badge", false, "show a third-party hits.sh site visits badge (requires -site-url)")
	flag.Parse()
	files, err := site.GenerateWithOptions(site.Options{BasePath: *base, SiteURL: *origin, VisitorBadge: *badge})
	if err != nil {
		log.Fatal(err)
	}
	if err := site.Write(*output, files); err != nil {
		log.Fatal(err)
	}
	log.Printf("Generated %d files in %s", len(files), *output)
}
