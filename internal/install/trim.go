package install

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Locales kept under Chrome .../locales/ (others removed to save space).
var keptLocales = map[string]bool{
	"en-US": true,
	"zh-CN": true,
	"zh-TW": true,
}

var chromeVersionDir = regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)

// TrimChrome removes non-essential files from a copied Chrome Application tree.
func TrimChrome(browserDir string) error {
	var total int64
	n, err := trimOldVersions(browserDir)
	if err != nil {
		return err
	}
	total += n

	n, err = trimLocalePaks(browserDir)
	if err != nil {
		return err
	}
	total += n

	n, err = removeExtraFiles(browserDir)
	if err != nil {
		return err
	}
	total += n

	if total > 0 {
		fmt.Printf("Trimmed %s from Chrome bundle\n", formatBytes(total))
	}
	return nil
}

func trimOldVersions(browserDir string) (int64, error) {
	entries, err := os.ReadDir(browserDir)
	if err != nil {
		return 0, err
	}

	var versions []string
	for _, e := range entries {
		if e.IsDir() && chromeVersionDir.MatchString(e.Name()) {
			versions = append(versions, e.Name())
		}
	}
	if len(versions) <= 1 {
		return 0, nil
	}

	keep := newestVersion(versions)
	var freed int64
	for _, v := range versions {
		if v == keep {
			continue
		}
		dir := filepath.Join(browserDir, v)
		n, err := dirSize(dir)
		if err != nil {
			return freed, err
		}
		if err := os.RemoveAll(dir); err != nil {
			return freed, err
		}
		fmt.Printf("Removed old Chrome version %s\n", v)
		freed += n
	}
	return freed, nil
}

func trimLocalePaks(browserDir string) (int64, error) {
	var freed int64
	err := filepath.WalkDir(browserDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() || !strings.EqualFold(d.Name(), "locales") {
			return nil
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.HasSuffix(strings.ToLower(name), ".pak") {
				continue
			}
			locale := strings.TrimSuffix(name, filepath.Ext(name))
			if keptLocales[locale] {
				continue
			}
			target := filepath.Join(path, name)
			info, err := e.Info()
			if err != nil {
				return err
			}
			if err := os.Remove(target); err != nil {
				return err
			}
			freed += info.Size()
		}
		return nil
	})
	return freed, err
}

func removeExtraFiles(browserDir string) (int64, error) {
	var freed int64
	for _, name := range []string{"setup.exe", "chrome_installer.log", "debug.log"} {
		path := filepath.Join(browserDir, name)
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if err := os.Remove(path); err != nil {
			return freed, err
		}
		freed += info.Size()
	}
	return freed, nil
}

func newestVersion(versions []string) string {
	best := versions[0]
	for _, v := range versions[1:] {
		if compareVersion(v, best) > 0 {
			best = v
		}
	}
	return best
}

func compareVersion(a, b string) int {
	ap := parseVersionParts(a)
	bp := parseVersionParts(b)
	n := len(ap)
	if len(bp) > n {
		n = len(bp)
	}
	for i := 0; i < n; i++ {
		av, bv := 0, 0
		if i < len(ap) {
			av = ap[i]
		}
		if i < len(bp) {
			bv = bp[i]
		}
		if av != bv {
			return av - bv
		}
	}
	return 0
}

func parseVersionParts(v string) []int {
	parts := strings.Split(v, ".")
	out := make([]int, len(parts))
	for i, p := range parts {
		out[i], _ = strconv.Atoi(p)
	}
	return out
}

func dirSize(root string) (int64, error) {
	var size int64
	err := filepath.WalkDir(root, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		size += info.Size()
		return nil
	})
	return size, err
}

func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
