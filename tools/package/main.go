// Package main copies installed Chrome into a portable folder (CI / local build).
package main

import (
	"flag"
	"log"

	"github.com/shzygq/chrome-portable/internal/bundle"
	"github.com/shzygq/chrome-portable/internal/install"
)

func main() {
	root := flag.String("root", ".", "portable folder (contains Chrome.exe)")
	flag.Parse()

	layout := bundle.NewLayout(*root)
	if err := install.Setup(layout); err != nil {
		log.Fatal(err)
	}
}
