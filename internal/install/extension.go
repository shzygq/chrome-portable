package install

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/shzygq/chrome-portable/internal/bundle"
	"github.com/shzygq/chrome-portable/internal/httpclient"
)

// BundleExtensions downloads extensions into Extensions/{id}/ (build-time, requires network).
func BundleExtensions(layout bundle.Layout) error {
	prodVersion := chromeProdVersion(layout.Browser)
	for _, ext := range bundle.BundledExtensions {
		if err := bundleExtension(layout, ext, prodVersion); err != nil {
			return err
		}
	}
	return nil
}

func bundleExtension(layout bundle.Layout, ext bundle.BundledExtension, prodVersion string) error {
	dir := layout.ExtensionDir(ext.ID)

	fmt.Printf("Downloading extension %s (%s)\n", ext.Name, ext.ID)
	crx, err := downloadCRX(ext.ID, prodVersion)
	if err != nil {
		return fmt.Errorf("download %s: %w", ext.Name, err)
	}

	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	if err := unpackCRX(crx, dir); err != nil {
		return fmt.Errorf("unpack %s: %w", ext.Name, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "extension.crx"), crx, 0o644); err != nil {
		return fmt.Errorf("write %s crx: %w", ext.Name, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "manifest.json")); err != nil {
		return fmt.Errorf("%s manifest.json missing after unpack", ext.Name)
	}

	fmt.Printf("Bundled extension into %s\n", dir)
	return nil
}

func chromeProdVersion(browserDir string) string {
	data, err := os.ReadFile(filepath.Join(browserDir, "VERSION"))
	if err != nil {
		return "131.0.0.0"
	}
	v := strings.TrimSpace(string(data))
	if v == "" {
		return "131.0.0.0"
	}
	return v
}

func crxURL(extensionID, prodVersion string) string {
	x := fmt.Sprintf("id=%s&installsource=ondemand&uc", extensionID)
	return fmt.Sprintf(
		"https://clients2.google.com/service/update2/crx?response=redirect&prodversion=%s&acceptformat=crx2,crx3&x=%s",
		prodVersion,
		urlQueryEscape(x),
	)
}

func urlQueryEscape(s string) string {
	replacer := strings.NewReplacer("%", "%25", "=", "%3D", "&", "%26")
	return replacer.Replace(s)
}

func downloadCRX(extensionID, prodVersion string) ([]byte, error) {
	url := crxURL(extensionID, prodVersion)
	resp, err := httpclient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &httpclient.HTTPError{URL: url, Status: resp.Status}
	}

	return io.ReadAll(resp.Body)
}

func unpackCRX(crx []byte, destDir string) error {
	zipStart := bytes.Index(crx, []byte("PK\x03\x04"))
	if zipStart < 0 {
		return fmt.Errorf("crx: zip payload not found")
	}

	zipBytes := crx[zipStart:]
	r := bytes.NewReader(zipBytes)
	zr, err := zip.NewReader(r, int64(len(zipBytes)))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	for _, f := range zr.File {
		target := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("crx: invalid path %q", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := extractZipFile(f, target); err != nil {
			return err
		}
	}
	return nil
}

func extractZipFile(f *zip.File, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode().Perm())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}
