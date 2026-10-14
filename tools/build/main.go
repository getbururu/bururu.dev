// Command build makes the site: it turns content/ into HTML pages in
// public/, for Cloudflare Pages to serve.
//
// From the bururu.dev folder:
//
//	go run ./tools/build
//	go run ./tools/build -content content -static static -out public
//
// Each version's pages go to /<version>/<page>/. A broken link or a link
// to a heading that is not there stops the build.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	content := flag.String("content", "content", "the content folder")
	static := flag.String("static", "static", "the folder copied into the site as it is")
	out := flag.String("out", "public", "the folder to write the site into; it is emptied first")
	base := flag.String("base", "https://bururu.dev", "the address the site is served at, for canonical links and the sitemap")
	flag.Parse()

	s, err := load(*content)
	if err != nil {
		log.Fatal(err)
	}
	s.base = *base
	for _, w := range s.warnings {
		log.Print(w)
	}
	n, err := s.write(*out, *static)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stdout, "wrote %d pages into %s\n", n, *out)
}
