package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/shzygq/chrome-portable/internal/bundle"
)

const warmupTimeout = 90 * time.Second

// WarmupProfile pre-initializes the profile and installs bundled extensions.
// Chrome 137+ removed --load-extension; extensions are registered as local
// external CRX files and loaded into the profile on the first headless launch.
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

	fmt.Println("Registering bundled extensions (external CRX)...")
	if err := InstallExtensionsExternal(layout); err != nil {
		return err
	}

	fmt.Println("Bootstrapping profile and installing extensions (headless)...")
	if err := bootstrapProfile(layout); err != nil {
		return err
	}

	if !allExtensionsInstalled(layout) {
		return fmt.Errorf("extension install failed: not present under Data/Default/Extensions after warmup")
	}

	fmt.Println("Profile ready")
	return nil
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

func bootstrapProfile(layout bundle.Layout) error {
	browser, err := filepath.Abs(layout.BrowserExe())
	if err != nil {
		return fmt.Errorf("warmup: resolve chrome.exe: %w", err)
	}

	cmd := exec.Command(browser, layout.WarmupArgs()...)
	cmd.Dir = filepath.Dir(browser)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("warmup start: %w", err)
	}

	prefs := filepath.Join(layout.Data, "Default", "Preferences")
	deadline := time.Now().Add(warmupTimeout)
	ready := false
	for time.Now().Before(deadline) {
		if _, err := os.Stat(prefs); err == nil && allExtensionsInstalled(layout) {
			time.Sleep(5 * time.Second)
			ready = true
			break
		}
		time.Sleep(time.Second)
	}

	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	_ = cmd.Wait()

	if !ready {
		if _, err := os.Stat(prefs); err != nil {
			return fmt.Errorf("warmup failed: profile not created")
		}
		return fmt.Errorf("warmup failed: extensions not installed into profile")
	}
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
