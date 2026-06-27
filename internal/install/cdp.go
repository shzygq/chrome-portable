package install

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/shzygq/chrome-portable/internal/bundle"
)

const cdpTimeout = 90 * time.Second

type cdpRequest struct {
	ID     int            `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params,omitempty"`
}

type cdpResponse struct {
	ID     int             `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *cdpError       `json:"error,omitempty"`
}

type cdpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type cdpPipe struct {
	out io.Writer
	in  io.Reader

	mu       sync.Mutex
	nextID   int
	pending  map[int]chan cdpResponse
	readOnce sync.Once
}

func newCDPPipe(out io.Writer, in io.Reader) *cdpPipe {
	p := &cdpPipe{
		out:     out,
		in:      in,
		pending: make(map[int]chan cdpResponse),
	}
	p.readOnce.Do(func() { go p.readLoop() })
	return p
}

func (p *cdpPipe) readLoop() {
	scanner := bufio.NewScanner(p.in)
	scanner.Split(splitNullDelimited)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var resp cdpResponse
		if err := json.Unmarshal(line, &resp); err != nil || resp.ID == 0 {
			continue
		}
		p.mu.Lock()
		ch := p.pending[resp.ID]
		delete(p.pending, resp.ID)
		p.mu.Unlock()
		if ch != nil {
			ch <- resp
		}
	}
}

func splitNullDelimited(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if i := bytes.IndexByte(data, 0); i >= 0 {
		return i + 1, data[:i], nil
	}
	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func (p *cdpPipe) call(method string, params map[string]any) (cdpResponse, error) {
	p.mu.Lock()
	p.nextID++
	id := p.nextID
	ch := make(chan cdpResponse, 1)
	p.pending[id] = ch
	p.mu.Unlock()

	req, err := json.Marshal(cdpRequest{ID: id, Method: method, Params: params})
	if err != nil {
		return cdpResponse{}, err
	}
	payload := append(req, 0)
	if _, err := p.out.Write(payload); err != nil {
		return cdpResponse{}, fmt.Errorf("cdp write: %w", err)
	}

	select {
	case resp := <-ch:
		if resp.Error != nil {
			return resp, fmt.Errorf("cdp %s: %s", method, resp.Error.Message)
		}
		return resp, nil
	case <-time.After(cdpTimeout):
		return cdpResponse{}, fmt.Errorf("cdp %s: timed out", method)
	}
}

// InstallExtensionsViaCDP registers bundled extensions in the profile using
// Extensions.loadUnpacked. Required for Chrome 137+ where --load-extension was removed.
func InstallExtensionsViaCDP(layout bundle.Layout) error {
	browser, err := filepath.Abs(layout.BrowserExe())
	if err != nil {
		return fmt.Errorf("cdp: resolve chrome.exe: %w", err)
	}
	if _, err := os.Stat(browser); err != nil {
		return fmt.Errorf("cdp: chrome.exe not found: %w", err)
	}

	args := layout.PortableFlags()
	args = append(args,
		"--headless=new",
		"--disable-gpu",
		"--remote-debugging-pipe",
		"--enable-unsafe-extension-debugging",
		"about:blank",
	)

	toChromeR, toChromeW, err := os.Pipe()
	if err != nil {
		return err
	}
	fromChromeR, fromChromeW, err := os.Pipe()
	if err != nil {
		toChromeR.Close()
		toChromeW.Close()
		return err
	}

	cmd := exec.Command(browser, args...)
	cmd.Dir = filepath.Dir(browser)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	cmd.ExtraFiles = []*os.File{toChromeR, fromChromeW}

	if err := cmd.Start(); err != nil {
		toChromeR.Close()
		toChromeW.Close()
		fromChromeR.Close()
		fromChromeW.Close()
		return fmt.Errorf("cdp start: %w", err)
	}
	toChromeR.Close()
	fromChromeW.Close()

	pipe := newCDPPipe(toChromeW, fromChromeR)
	defer func() {
		toChromeW.Close()
		fromChromeR.Close()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	}()

	if _, err := pipe.call("Browser.getVersion", nil); err != nil {
		return fmt.Errorf("cdp connect: %w", err)
	}

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

		fmt.Printf("Installing extension %s (%s) via CDP\n", ext.Name, ext.ID)
		resp, err := pipe.call("Extensions.loadUnpacked", map[string]any{
			"path": filepath.ToSlash(path),
		})
		if err != nil {
			return fmt.Errorf("install %s: %w", ext.Name, err)
		}

		var result struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(resp.Result, &result); err != nil {
			return fmt.Errorf("install %s: parse result: %w", ext.Name, err)
		}
		if result.ID != ext.ID {
			return fmt.Errorf("install %s: expected id %s, got %s", ext.Name, ext.ID, result.ID)
		}
	}

	time.Sleep(3 * time.Second)
	return nil
}

func extensionInstalled(layout bundle.Layout, id string) bool {
	dir := filepath.Join(layout.Data, "Default", "Extensions", id)
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			if _, err := os.Stat(filepath.Join(dir, e.Name(), "manifest.json")); err == nil {
				return true
			}
		}
	}
	return false
}
