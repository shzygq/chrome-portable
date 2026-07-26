package bundle

import (
	"fmt"
	"path/filepath"
)

func (l Layout) portableFlags() []string {
	return []string{
		fmt.Sprintf("--user-data-dir=%s", chromePath(l.Data)),
		fmt.Sprintf("--profile-directory=%s", WorkProfile),
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
		"--disable-features=TranslateUI,MediaRouter,OptimizationHints,InterestFeedContentSuggestions",
		// Less throttling for background tabs / occluded windows (uses more CPU).
		"--disable-backgrounding-occluded-windows",
		"--disable-renderer-backgrounding",
		"--disable-background-timer-throttling",
		"--disable-hang-monitor",
		"--disable-ipc-flooding-protection",
	}
}

// ChromeArgs returns flags for normal browser launch.
func (l Layout) ChromeArgs() []string {
	return l.portableFlags()
}

// WarmupArgs is used at packaging time to pre-create the profile.
func (l Layout) WarmupArgs() []string {
	return append(l.portableFlags(), "--headless=new", "--disable-gpu", "about:blank")
}

// chromePath normalizes a filesystem path for Chrome command-line flags on Windows.
func chromePath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	return filepath.ToSlash(abs)
}
