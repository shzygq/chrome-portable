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

// WarmupProfile pre-initializes the profile and installs bundled extensions via CDP.
// Chrome 137+ removed --load-extension; Extensions.loadUnpacked over
// --remote-debugging-pipe persists extensions in the profile like "Load unpacked".
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

	if _, err := os.Stat(filepath.Join(layout.Data, "Default", "Preferences")); err != nil {
		if err := bootstrapProfile(layout); err != nil {
			return err
		}
	}

	fmt.Println("Installing bundled extensions into profile (CDP)...")
	if err := InstallExtensionsViaCDP(layout); err != nil {
		return err
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
	fmt.Println("Bootstrapping profile (headless)...")
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
		if _, err := os.Stat(prefs); err == nil {
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
		return fmt.Errorf("warmup failed: profile not created")
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
