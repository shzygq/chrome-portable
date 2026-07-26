package portable

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shzygq/chrome-portable/internal/bundle"
)

// Run starts the bundled Chrome next to Chrome.exe.
// Host Platform policies are temporarily isolated for the session (PortableApps-style).
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

	guard := newPolicyGuard()
	guard.Isolate()
	defer guard.Restore()

	cmd := exec.Command(browser, layout.ChromeArgs()...)
	cmd.Dir = filepath.Dir(browser)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}
