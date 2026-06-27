package main

import (
	"os"
	"runtime"

	"github.com/shzygq/chrome-portable/internal/portable"
)

//go:generate go run ../../tools/genicon

func main() {
	if runtime.GOOS != "windows" {
		os.Exit(1)
	}
	if err := portable.Run(); err != nil {
		os.Exit(1)
	}
}
