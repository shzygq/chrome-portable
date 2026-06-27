package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shzygq/chrome-portable/internal/bundle"
)

const extensionInstallResultsFile = "extension-install-results.json"

type extensionInstallResult struct {
	Name      string `json:"name"`
	StoreID   string `json:"storeId"`
	ProfileID string `json:"profileId"`
}

type extensionManifest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Key     string `json:"key"`
}

func extensionInstallResultsPath(layout bundle.Layout) string {
	return filepath.Join(layout.Data, extensionInstallResultsFile)
}

// extensionInstalled reports whether a bundled extension is present in the profile.
func extensionInstalled(layout bundle.Layout, bundleID string) bool {
	if id, ok := profileIDForStore(layout, bundleID); ok {
		if profileExtensionPresent(layout, id) {
			return true
		}
	}
	return extensionInstalledByManifest(layout, bundleID)
}

func profileIDForStore(layout bundle.Layout, storeID string) (string, bool) {
	results, err := readExtensionInstallResults(layout)
	if err != nil {
		return "", false
	}
	for _, r := range results {
		if r.StoreID == storeID && r.ProfileID != "" {
			return r.ProfileID, true
		}
	}
	return "", false
}

func readExtensionInstallResults(layout bundle.Layout) ([]extensionInstallResult, error) {
	data, err := os.ReadFile(extensionInstallResultsPath(layout))
	if err != nil {
		return nil, err
	}
	var results []extensionInstallResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func profileExtensionPresent(layout bundle.Layout, profileID string) bool {
	dir := filepath.Join(layout.Data, "Default", "Extensions", profileID)
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

func extensionInstalledByManifest(layout bundle.Layout, bundleID string) bool {
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
	if a.Version == "" || a.Version != b.Version {
		return false
	}
	if a.Key != "" && b.Key != "" {
		return a.Key == b.Key
	}
	return a.Name != "" && a.Name == b.Name
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

func missingExtensions(layout bundle.Layout) []string {
	var missing []string
	for _, ext := range bundle.BundledExtensions {
		if !extensionInstalled(layout, ext.ID) {
			missing = append(missing, fmt.Sprintf("%s (%s)", ext.Name, ext.ID))
		}
	}
	return missing
}
