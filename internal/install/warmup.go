package install

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/shzygq/chrome-portable/internal/bundle"
)

const warmupTimeout = 90 * time.Second

// WarmupProfile pre-creates the Work profile via a short headless launch,
// then sets its display name to "Work".
func WarmupProfile(layout bundle.Layout) error {
	if err := layout.EnsureDataDirs(); err != nil {
		return err
	}

	prefs := filepath.Join(layout.ProfileDir(), "Preferences")
	if _, err := os.Stat(prefs); err != nil {
		if err := bootstrapWorkProfile(layout); err != nil {
			return err
		}
	}

	if err := nameWorkProfile(layout); err != nil {
		return err
	}

	fmt.Println("Profile ready: Work")
	return nil
}

func bootstrapWorkProfile(layout bundle.Layout) error {
	fmt.Println("Bootstrapping Work profile (headless)...")
	browser, err := filepath.Abs(layout.BrowserExe())
	if err != nil {
		return fmt.Errorf("warmup: resolve chrome.exe: %w", err)
	}

	cmd := exec.Command(browser, layout.WarmupArgs()...)
	cmd.Dir = filepath.Dir(browser)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("warmup start: %w", err)
	}

	prefs := filepath.Join(layout.ProfileDir(), "Preferences")
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
		return fmt.Errorf("warmup failed: Work profile not created")
	}
	return nil
}

func nameWorkProfile(layout bundle.Layout) error {
	if err := patchJSON(filepath.Join(layout.ProfileDir(), "Preferences"), func(root map[string]any) {
		profile, _ := root["profile"].(map[string]any)
		if profile == nil {
			profile = map[string]any{}
			root["profile"] = profile
		}
		profile["name"] = bundle.WorkProfile
		profile["is_using_default_name"] = false
	}); err != nil {
		return fmt.Errorf("set Preferences profile name: %w", err)
	}

	if err := patchJSON(filepath.Join(layout.Data, "Local State"), func(root map[string]any) {
		profile, _ := root["profile"].(map[string]any)
		if profile == nil {
			profile = map[string]any{}
			root["profile"] = profile
		}
		info, _ := profile["info_cache"].(map[string]any)
		if info == nil {
			info = map[string]any{}
			profile["info_cache"] = info
		}
		entry, _ := info[bundle.WorkProfile].(map[string]any)
		if entry == nil {
			entry = map[string]any{}
			info[bundle.WorkProfile] = entry
		}
		entry["name"] = bundle.WorkProfile
		entry["is_using_default_name"] = false
		profile["last_used"] = bundle.WorkProfile
		profile["last_active_profiles"] = []any{bundle.WorkProfile}
	}); err != nil {
		return fmt.Errorf("set Local State profile name: %w", err)
	}
	return nil
}

func patchJSON(path string, mutate func(map[string]any)) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}
	mutate(root)
	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}
