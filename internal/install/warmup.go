package install

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shzygq/chrome-portable/internal/bundle"
)

// WarmupProfile pre-initializes the profile and installs bundled extensions.
// Chrome 137+ removed --load-extension; build uses CDP Extensions.loadUnpacked
// via Node chrome-launcher (Go cannot pass pipe fds on Windows).
func WarmupProfile(layout bundle.Layout) error {
	if err := layout.EnsureDataDirs(); err != nil {
		return err
	}
	if err := verifyBundledExtensions(layout); err != nil {
		return err
	}

	if allExtensionsInstalled(layout) {
		fmt.Println("Profile and extensions already ready")
		return nil
	}

	fmt.Println("Installing bundled extensions into profile (CDP via node)...")
	if err := InstallExtensionsViaNode(layout); err != nil {
		return err
	}

	if !allExtensionsInstalled(layout) {
		return fmt.Errorf("extension install failed: not present under Data/Default/Extensions after CDP install")
	}

	fmt.Println("Profile ready")
	return nil
}

func verifyBundledExtensions(layout bundle.Layout) error {
	for _, ext := range bundle.BundledExtensions {
		if _, err := os.Stat(filepath.Join(layout.ExtensionDir(ext.ID), "manifest.json")); err != nil {
			return fmt.Errorf("extension %s (%s) missing: %w", ext.Name, ext.ID, err)
		}
	}
	return nil
}
