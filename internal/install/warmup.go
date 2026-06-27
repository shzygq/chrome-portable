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

// WarmupProfile launches Chrome headless with the same args as production to pre-initialize the profile.
func WarmupProfile(layout bundle.Layout) error {
	if err := layout.EnsureDataDirs(); err != nil {
		return err
	}
	if err := verifyBundledExtensions(layout); err != nil {
		return err
	}

	fmt.Println("Warming up profile (headless)...")
	browser, err := filepath.Abs(layout.BrowserExe())
	if err != nil {
		return fmt.Errorf("warmup: resolve chrome.exe: %w", err)
	}
	if _, err := os.Stat(browser); err != nil {
		return fmt.Errorf("warmup: chrome.exe not found at %s: %w", browser, err)
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
			time.Sleep(10 * time.Second)
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
