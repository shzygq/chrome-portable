package install

import (
	"os"
	"path/filepath"

	"github.com/shzygq/chrome-portable/internal/bundle"
)

func extensionInstalled(layout bundle.Layout, id string) bool {
	dir := filepath.Join(layout.Data, "Default", "Extensions", id)
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			if _, err := os.Stat(filepath.Join(dir, e.Name(), "manifest.json")); err == nil {
				return true
			}
		}
	}
	return false
}

func allExtensionsInstalled(layout bundle.Layout) bool {
	if _, err := os.Stat(filepath.Join(layout.Data, "Default", "Preferences")); err != nil {
		return false
	}
	for _, ext := range bundle.BundledExtensions {
		if !extensionInstalled(layout, ext.ID) {
			return false
		}
	}
	return true
}
