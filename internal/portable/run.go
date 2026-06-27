package portable

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shzygq/chrome-portable/internal/bundle"
)

// Run starts the bundled Chrome next to Chrome.exe.
func Run() error {
	root, err := bundle.Root()
	if err != nil {
		return err
	}

	layout := bundle.NewLayout(root)
	browser, err := filepath.Abs(layout.BrowserExe())
	if err != nil {
		return err
	}
	if _, err := os.Stat(browser); err != nil {
		return fmt.Errorf("Chrome not found in %s", layout.Root)
	}
	if err := layout.EnsureDataDirs(); err != nil {
		return err
	}

	cmd := exec.Command(browser, layout.ChromeArgs()...)
	cmd.Dir = filepath.Dir(browser)
	return cmd.Start()
}
