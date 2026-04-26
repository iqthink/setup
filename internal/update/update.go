package update

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const repo = "iqthink/setup"

// Run downloads the latest release binary from GitHub and replaces the
// currently running iqdev executable atomically.
func Run(currentVersion string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("only darwin is supported for now (you are on %s)", runtime.GOOS)
	}
	if runtime.GOARCH != "arm64" && runtime.GOARCH != "amd64" {
		return fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	fmt.Println("Looking up latest release on GitHub...")
	latest, err := latestTag()
	if err != nil {
		return fmt.Errorf("fetching latest release: %w", err)
	}
	fmt.Printf("Latest: %s (current: %s)\n", latest, currentVersion)

	if normalizeVersion(currentVersion) == normalizeVersion(latest) {
		fmt.Println("You are already on the latest version.")
		return nil
	}

	asset := fmt.Sprintf("iqdev-darwin-%s", runtime.GOARCH)
	sumFile := fmt.Sprintf("SHA256SUMS-darwin-%s.txt", runtime.GOARCH)
	base := fmt.Sprintf("https://github.com/%s/releases/download/%s", repo, latest)

	fmt.Printf("Downloading %s...\n", asset)
	binData, err := download(base + "/" + asset)
	if err != nil {
		return err
	}
	sumData, err := download(base + "/" + sumFile)
	if err != nil {
		return err
	}

	fields := strings.Fields(strings.TrimSpace(string(sumData)))
	if len(fields) == 0 {
		return fmt.Errorf("checksum file is empty")
	}
	expected := fields[0]
	actual := hex.EncodeToString(func() []byte { s := sha256.Sum256(binData); return s[:] }())
	if expected != actual {
		return fmt.Errorf("checksum mismatch (expected %s, got %s)", expected, actual)
	}
	fmt.Println("Checksum verified.")

	self, err := os.Executable()
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(self); err == nil {
		self = resolved
	}
	tmp := self + ".new"
	if err := os.WriteFile(tmp, binData, 0o755); err != nil {
		return err
	}
	if err := os.Rename(tmp, self); err != nil {
		return fmt.Errorf("replacing %s: %w", self, err)
	}
	fmt.Printf("Updated to %s.\n", latest)
	return nil
}

func latestTag() (string, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github API returned %d", resp.StatusCode)
	}
	var body struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	if body.TagName == "" {
		return "", fmt.Errorf("github response missing tag_name")
	}
	return body.TagName, nil
}

func download(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed (%d): %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

func normalizeVersion(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}
