package httpclient

import (
	"io"
	"net/http"
	"os"
)

// UserAgent is a common Chrome-on-Windows user agent for HTTP downloads.
const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

// Get performs an HTTP GET with the common user agent.
func Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	return http.DefaultClient.Do(req)
}

// Download saves url to dest using the common user agent.
func Download(url, dest string) error {
	resp, err := Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &HTTPError{URL: url, Status: resp.Status}
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// HTTPError is returned for non-200 download responses.
type HTTPError struct {
	URL    string
	Status string
}

func (e *HTTPError) Error() string {
	return "download " + e.URL + ": HTTP " + e.Status
}
