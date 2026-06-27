// Generate Windows .syso with icon from Google's official Chrome logo.
//
//	go run ./tools/genicon
package main

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"github.com/tc-hib/winres"
	"github.com/tc-hib/winres/version"
	"github.com/shzygq/chrome-portable/internal/httpclient"
)

const logoURL = "https://www.google.com/chrome/static/images/chrome-logo.svg"

func main() {
	root := moduleRoot()
	outDir := filepath.Join(root, "cmd", "chrome")

	img, err := fetchLogo()
	if err != nil {
		fail(err)
	}

	icon, err := winres.NewIconFromResizedImage(img, nil)
	if err != nil {
		fail(err)
	}

	for _, spec := range []struct {
		file string
		arch winres.Arch
	}{
		{"resource_windows_amd64.syso", winres.ArchAMD64},
		{"resource_windows_arm64.syso", winres.ArchARM64},
	} {
		path := filepath.Join(outDir, spec.file)
		fmt.Println("write", path)
		if err := writeSyso(path, spec.arch, icon); err != nil {
			fail(err)
		}
	}
}

func fetchLogo() (image.Image, error) {
	fmt.Println("download", logoURL)

	resp, err := httpclient.Get(logoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %s", resp.Status)
	}

	svg, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	const n = 256
	icon, err := oksvg.ReadIconStream(bytes.NewReader(svg), oksvg.WarnErrorMode)
	if err != nil {
		return nil, err
	}

	icon.SetTarget(0, 0, n, n)
	rgba := image.NewRGBA(image.Rect(0, 0, n, n))
	scanner := rasterx.NewScannerGV(n, n, rgba, rgba.Bounds())
	raster := rasterx.NewDasher(n, n, scanner)
	icon.Draw(raster, 1)
	return rgba, nil
}

func writeSyso(path string, arch winres.Arch, icon *winres.Icon) error {
	rs := winres.ResourceSet{}
	if err := rs.SetIcon(winres.ID(1), icon); err != nil {
		return err
	}

	vi := version.Info{
		FileVersion:    [4]uint16{1, 0, 0, 0},
		ProductVersion: [4]uint16{1, 0, 0, 0},
	}
	vi.Set(version.LangDefault, version.FileDescription, "Portable Chrome")
	vi.Set(version.LangDefault, version.ProductName, "Portable Chrome")
	vi.Set(version.LangDefault, version.OriginalFilename, "Chrome.exe")
	vi.Set(version.LangDefault, version.FileVersion, "1.0.0.0")
	vi.Set(version.LangDefault, version.ProductVersion, "1.0.0.0")
	rs.SetVersionInfo(vi)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return rs.WriteObject(f, arch)
}

func moduleRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
