package bundle

import (
	"fmt"
	"path/filepath"
)

func (l Layout) portableFlags() []string {
	return []string{
		fmt.Sprintf("--user-data-dir=%s", chromePath(l.Data)),
		fmt.Sprintf("--disk-cache-dir=%s", chromePath(l.Cache)),
		fmt.Sprintf("--media-cache-dir=%s", chromePath(l.Media)),
		fmt.Sprintf("--gpu-disk-cache-dir=%s", chromePath(l.GPUCache)),
		fmt.Sprintf("--crash-dumps-dir=%s", chromePath(l.Crash)),
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-logging",
		"--disable-breakpad",
		// Reduce background connections to Google (offline / no Google access).
		"--disable-sync",
		"--disable-background-networking",
		"--disable-component-update",
		"--disable-domain-reliability",
		"--disable-client-side-phishing-detection",
		"--disable-features=TranslateUI",
	}
}

// ChromeArgs returns flags for normal browser launch.
func (l Layout) ChromeArgs() []string {
	return l.portableFlags()
}

// chromePath normalizes a filesystem path for Chrome command-line flags on Windows.
func chromePath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	return filepath.ToSlash(abs)
}
