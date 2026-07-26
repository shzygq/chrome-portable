//go:build windows

package portable

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// Host Chrome policy keys applied by IT via registry (Platform source).
var hostPolicyKeys = []string{
	`HKCU\Software\Policies\Google\Chrome`,
	`HKLM\SOFTWARE\Policies\Google\Chrome`,
}

type policyBackup struct {
	key  string
	path string
}

type policyGuard struct {
	backups []policyBackup
}

func newPolicyGuard() *policyGuard {
	return &policyGuard{}
}

// Isolate backs up host policy keys and removes them for this Chrome session.
// PortableApps Google Chrome Portable does the same for HKCU; HKLM is attempted
// and succeeds only when the launcher runs with administrator rights.
func (g *policyGuard) Isolate() {
	for _, key := range hostPolicyKeys {
		g.isolateKey(key)
	}
}

func (g *policyGuard) Restore() {
	for i := len(g.backups) - 1; i >= 0; i-- {
		b := g.backups[i]
		_ = regImport(b.path)
		_ = os.Remove(b.path)
	}
	g.backups = nil
}

func (g *policyGuard) isolateKey(key string) {
	if !regKeyExists(key) {
		return
	}

	f, err := os.CreateTemp("", "chrome-portable-policy-*.reg")
	if err != nil {
		return
	}
	path := f.Name()
	_ = f.Close()

	if err := regExport(key, path); err != nil {
		_ = os.Remove(path)
		return
	}
	if err := regDelete(key); err != nil {
		_ = os.Remove(path)
		return
	}

	g.backups = append(g.backups, policyBackup{key: key, path: path})
}

func regKeyExists(key string) bool {
	parts := strings.SplitN(key, `\`, 2)
	if len(parts) != 2 {
		return false
	}
	subkey := parts[1]
	switch parts[0] {
	case "HKCU":
		out, err := regQuery(`HKCU`, subkey)
		return err == nil && out
	case "HKLM":
		out, err := regQuery(`HKLM`, subkey)
		return err == nil && out
	default:
		return false
	}
}

func regQuery(root, subkey string) (bool, error) {
	cmd := exec.Command("reg", "query", root+`\`+subkey)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Run()
	return err == nil, err
}

func regExport(key, path string) error {
	cmd := exec.Command("reg", "export", key, path, "/y")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("reg export %s: %w: %s", key, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func regDelete(key string) error {
	cmd := exec.Command("reg", "delete", key, "/f")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("reg delete %s: %w: %s", key, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func regImport(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	cmd := exec.Command("reg", "import", abs)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("reg import: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
