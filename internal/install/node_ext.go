package install

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shzygq/chrome-portable/internal/bundle"
)

type nodeInstallConfig struct {
	Chrome     string          `json:"chrome"`
	ChromeDir  string          `json:"chromeDir"`
	LogDir     string          `json:"logDir"`
	ChromeArgs []string        `json:"chromeArgs"`
	Extensions []nodeExtension `json:"extensions"`
}

type nodeExtension struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// InstallExtensionsViaNode registers extensions with CDP Extensions.loadUnpacked.
// Node chrome-launcher can pass CRT pipe fds on Windows; Go os/exec ExtraFiles cannot.
func InstallExtensionsViaNode(layout bundle.Layout) error {
	toolDir, err := installExtToolDir()
	if err != nil {
		return err
	}

	browser, err := filepath.Abs(layout.BrowserExe())
	if err != nil {
		return fmt.Errorf("node install: resolve chrome.exe: %w", err)
	}
	if _, err := os.Stat(browser); err != nil {
		return fmt.Errorf("node install: chrome.exe not found: %w", err)
	}

	args := layout.PortableFlags()
	args = append(args,
		"--disable-gpu",
	)

	var exts []nodeExtension
	for _, ext := range bundle.BundledExtensions {
		if extensionInstalled(layout, ext.ID) {
			fmt.Printf("Extension %s (%s) already in profile\n", ext.Name, ext.ID)
			continue
		}
		path, err := filepath.Abs(layout.ExtensionDir(ext.ID))
		if err != nil {
			return err
		}
		if _, err := os.Stat(filepath.Join(path, "manifest.json")); err != nil {
			return fmt.Errorf("extension %s missing manifest: %w", ext.Name, err)
		}
		exts = append(exts, nodeExtension{
			ID:   ext.ID,
			Name: ext.Name,
			Path: filepath.ToSlash(path),
		})
	}
	if len(exts) == 0 {
		return nil
	}

	cfgFile, err := os.CreateTemp("", "chrome-portable-install-*.json")
	if err != nil {
		return err
	}
	cfgPath := cfgFile.Name()
	defer os.Remove(cfgPath)

	cfg := nodeInstallConfig{
		Chrome:     browser,
		ChromeDir:  filepath.Dir(browser),
		LogDir:     layout.Data,
		ChromeArgs: args,
		Extensions: exts,
	}
	if err := json.NewEncoder(cfgFile).Encode(cfg); err != nil {
		cfgFile.Close()
		return err
	}
	if err := cfgFile.Close(); err != nil {
		return err
	}

	cmd := exec.Command("node", "install.mjs", cfgPath)
	cmd.Dir = toolDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("node install: %w", err)
	}
	return nil
}

func installExtToolDir() (string, error) {
	for _, root := range moduleRootCandidates() {
		dir := filepath.Join(root, "tools", "install-ext")
		script := filepath.Join(dir, "install.mjs")
		if _, err := os.Stat(script); err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, "node_modules", "chrome-launcher")); err != nil {
			return "", fmt.Errorf("chrome-launcher not installed: run npm ci in tools/install-ext")
		}
		return dir, nil
	}
	return "", fmt.Errorf("tools/install-ext not found")
}

func moduleRootCandidates() []string {
	var roots []string
	wd, err := os.Getwd()
	if err != nil {
		return roots
	}
	for dir := wd; ; dir = filepath.Dir(dir) {
		roots = append(roots, dir)
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		if filepath.Dir(dir) == dir {
			break
		}
	}
	return roots
}
