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
	flag.Parse()
	files, err := site.Generate(*base)
	if err != nil {
		log.Fatal(err)
	}
	if err := site.Write(*output, files); err != nil {
		log.Fatal(err)
	}
	log.Printf("Generated %d files in %s", len(files), *output)
}
