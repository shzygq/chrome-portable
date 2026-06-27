package bundle

import (
	"fmt"
	"os"
	"path/filepath"
)

const Launcher = "Chrome.exe"

// Root is always the directory containing Chrome.exe (the portable folder).
func Root() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir, err := filepath.Abs(filepath.Dir(exe))
	if err != nil {
		return "", err
	}
	return dir, nil
}

// Layout describes files under the portable folder.
type Layout struct {
	Root       string
	Data       string
	Browser    string
	Extensions string
}

// NewLayout builds paths for the portable folder. Root is always absolute.
func NewLayout(root string) Layout {
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = root
	}
	root = abs
	return Layout{
		Root:       root,
		Data:       filepath.Join(root, "Data"),
		Browser:    filepath.Join(root, "Chrome"),
		Extensions: filepath.Join(root, "Extensions"),
	}
}

func (l Layout) ExtensionDir(id string) string {
	return filepath.Join(l.Extensions, id)
}

func (l Layout) BrowserExe() string {
	return filepath.Join(l.Browser, "chrome.exe")
}

func (l Layout) EnsureDirs() error {
	for _, dir := range []string{l.Root, l.Data, l.Browser, l.Extensions} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}
	return nil
}

func (l Layout) EnsureDataDirs() error {
	if err := os.MkdirAll(l.Data, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", l.Data, err)
	}
	return nil
}
