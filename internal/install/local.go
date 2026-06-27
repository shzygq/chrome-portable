package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shzygq/chrome-portable/internal/bundle"
)

// FromLocal copies an installed Chrome into the bundle.
func FromLocal(chromeExe string, layout bundle.Layout) error {
	exe, err := FindInstalled(chromeExe)
	if err != nil {
		return err
	}

	appDir, err := applicationDir(exe)
	if err != nil {
		return err
	}

	fmt.Printf("Copying Chrome from %s\n", appDir)
	if err := os.RemoveAll(layout.Browser); err != nil {
		return err
	}
	if err := os.MkdirAll(layout.Browser, 0o755); err != nil {
		return err
	}
	if err := CopyDir(appDir, layout.Browser); err != nil {
		return fmt.Errorf("copy Chrome: %w", err)
	}
	if _, err := os.Stat(layout.BrowserExe()); err != nil {
		return fmt.Errorf("chrome.exe missing after copy")
	}
	return nil
}

// FindInstalled locates a system Chrome install.
func FindInstalled(explicit string) (string, error) {
	if explicit == "" {
		explicit = os.Getenv("CHROME_EXE")
	}
	if explicit != "" {
		if _, err := os.Stat(explicit); err != nil {
			return "", fmt.Errorf("not found: %q", explicit)
		}
		return explicit, nil
	}

	for _, p := range chromeCandidates() {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("Chrome not installed")
}

func chromeCandidates() []string {
	var paths []string
	for _, base := range []string{
		os.Getenv("ProgramFiles"),
		os.Getenv("ProgramFiles(x86)"),
		os.Getenv("LOCALAPPDATA"),
	} {
		if base == "" {
			continue
		}
		paths = append(paths, filepath.Join(base, "Google", "Chrome", "Application", "chrome.exe"))
	}
	return paths
}

func applicationDir(chromeExe string) (string, error) {
	dir := filepath.Dir(chromeExe)
	if strings.EqualFold(filepath.Base(dir), "Application") {
		return dir, nil
	}
	parent := filepath.Dir(dir)
	if strings.EqualFold(filepath.Base(parent), "Application") {
		return parent, nil
	}
	return "", fmt.Errorf("invalid chrome path: %q", chromeExe)
}
