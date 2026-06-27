package install

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/shzygq/chrome-portable/internal/bundle"
)

type extensionManifest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// extensionInstalled reports whether a bundled extension is present in the profile.
// Unpacked installs may use a different Chrome extension id than the Web Store id.
func extensionInstalled(layout bundle.Layout, bundleID string) bool {
	want, err := readExtensionManifest(filepath.Join(layout.ExtensionDir(bundleID), "manifest.json"))
	if err != nil {
		return false
	}

	profileRoot := filepath.Join(layout.Data, "Default", "Extensions")
	ids, err := os.ReadDir(profileRoot)
	if err != nil {
		return false
	}
	for _, idEntry := range ids {
		if !idEntry.IsDir() {
			continue
		}
		versions, err := os.ReadDir(filepath.Join(profileRoot, idEntry.Name()))
		if err != nil {
			continue
		}
		for _, ver := range versions {
			if !ver.IsDir() {
				continue
			}
			got, err := readExtensionManifest(filepath.Join(profileRoot, idEntry.Name(), ver.Name(), "manifest.json"))
			if err != nil {
				continue
			}
			if manifestsMatch(want, got) {
				return true
			}
		}
	}
	return false
}

func readExtensionManifest(path string) (extensionManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return extensionManifest{}, err
	}
	var m extensionManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return extensionManifest{}, err
	}
	return m, nil
}

func manifestsMatch(a, b extensionManifest) bool {
	return a.Name != "" && a.Name == b.Name && a.Version != "" && a.Version == b.Version
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
