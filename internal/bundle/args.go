package bundle

import (
	"fmt"
	"os"
	"path/filepath"
)

func (l Layout) portableFlags() []string {
	return []string{
		fmt.Sprintf("--user-data-dir=%s", chromePath(l.Data)),
		"--no-default-browser-check",
		"--disable-logging",
		"--disable-breakpad",
	}
}

// ChromeArgs returns flags for normal browser launch.
func (l Layout) ChromeArgs() []string {
	args := l.portableFlags()
	return append(args, l.extensionArgs()...)
}

func (l Layout) extensionArgs() []string {
	var args []string
	for _, ext := range BundledExtensions {
		root := l.ExtensionDir(ext.ID)
		if _, err := os.Stat(filepath.Join(root, "manifest.json")); err != nil {
			continue
		}
		args = append(args, "--load-extension="+chromePath(root))
	}
	return args
}

// WarmupArgs uses the same launch flags as production plus headless mode.
func (l Layout) WarmupArgs() []string {
	return append(l.ChromeArgs(), "--headless=new", "--disable-gpu", "about:blank")
}

// chromePath normalizes a filesystem path for Chrome command-line flags on Windows.
func chromePath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	return filepath.ToSlash(abs)
}
