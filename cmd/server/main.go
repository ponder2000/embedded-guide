// Command server previews an already-generated site using only net/http.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jaysaha/embedded-visual-guide/internal/preview"
)

func main() {
	dir := flag.String("dir", "dist", "generated site directory")
	base := flag.String("base-path", "", "deployment URL prefix (same as build)")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := flag.String("addr", ":"+port, "listen address")
	flag.Parse()
	handler, err := preview.New(*dir, *base)
	if err != nil {
		log.Fatal(err)
	}
	server := &http.Server{Addr: *addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("Static preview: http://localhost%s%s/", *addr, *base)
	log.Fatal(server.ListenAndServe())
}
