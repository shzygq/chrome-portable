package install

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/shzygq/chrome-portable/internal/bundle"
)

// Setup copies Chrome and pre-warms the profile.
func Setup(layout bundle.Layout) error {
	if err := layout.EnsureDirs(); err != nil {
		return err
	}
	if err := FromLocal("", layout); err != nil {
		return err
	}
	if err := WritePrefs(layout); err != nil {
		return err
	}
	if err := BundleExtensions(layout); err != nil {
		return err
	}
	return WarmupProfile(layout)
}

func WritePrefs(layout bundle.Layout) error {
	prefs, _ := json.Marshal(map[string]any{
		"distribution": map[string]any{
			"skip_first_run_ui":   true,
			"make_chrome_default": false,
		},
		"browser": map[string]any{"check_default_browser": false},
	})
	return os.WriteFile(filepath.Join(layout.Data, "master_preferences"), prefs, 0o644)
}
