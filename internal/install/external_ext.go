package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shzygq/chrome-portable/internal/bundle"
)

type externalExtensionSpec struct {
	ExternalCRX      string `json:"external_crx"`
	ExternalVersion  string `json:"external_version"`
}

// InstallExtensionsExternal registers bundled CRX files under Data/External Extensions/.
// Chrome loads these on the next launch and persists them in the profile.
// Used on Windows where Go cannot pass CRT fds 3/4 for --remote-debugging-pipe.
func InstallExtensionsExternal(layout bundle.Layout) error {
	dir := filepath.Join(layout.Data, "External Extensions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir external extensions: %w", err)
	}

	for _, ext := range bundle.BundledExtensions {
		if extensionInstalled(layout, ext.ID) {
			continue
		}

		crxPath, err := filepath.Abs(filepath.Join(layout.ExtensionDir(ext.ID), "extension.crx"))
		if err != nil {
			return err
		}
		if _, err := os.Stat(crxPath); err != nil {
			return fmt.Errorf("extension %s: %w", ext.Name, err)
		}

		version, err := extensionManifestVersion(layout, ext.ID)
		if err != nil {
			return fmt.Errorf("extension %s: %w", ext.Name, err)
		}

		spec := externalExtensionSpec{
			ExternalCRX:     filepath.ToSlash(crxPath),
			ExternalVersion: version,
		}
		data, err := json.Marshal(spec)
		if err != nil {
			return err
		}

		jsonPath := filepath.Join(dir, ext.ID+".json")
		if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", jsonPath, err)
		}
		fmt.Printf("Registered external extension %s (%s) version %s\n", ext.Name, ext.ID, version)
	}
	return nil
}

func extensionManifestVersion(layout bundle.Layout, id string) (string, error) {
	data, err := os.ReadFile(filepath.Join(layout.ExtensionDir(id), "manifest.json"))
	if err != nil {
		return "", err
	}
	var manifest struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return "", err
	}
	if manifest.Version == "" {
		return "", fmt.Errorf("manifest.json: missing version")
	}
	return manifest.Version, nil
}

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
