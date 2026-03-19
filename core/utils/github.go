package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// githubRelease represents the minimal fields from GitHub Releases API response.
type githubRelease struct {
	TagName string `json:"tag_name"`
}

const defaultProxy = "https://dl.idb.net"

// GetLatestReleaseVersion queries the GitHub Releases API for the latest release
// of the given repository (e.g. "sensdata/idb") and returns the tag name.
// It first tries GitHub directly; if that fails, it falls back to the dl.idb.net proxy.
// An optional proxyURL can be provided to override the default proxy.
// Returns an empty string if all attempts fail.
func GetLatestReleaseVersion(githubRepo string, proxyURL ...string) string {
	if githubRepo == "" {
		return ""
	}

	// Try GitHub directly first
	directURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	if tag := fetchReleaseTag(directURL); tag != "" {
		return tag
	}

	// Fallback to proxy
	proxy := defaultProxy
	if len(proxyURL) > 0 && proxyURL[0] != "" {
		proxy = strings.TrimRight(proxyURL[0], "/")
	}
	proxyAPIURL := fmt.Sprintf("%s/github-api/repos/%s/releases/latest", proxy, githubRepo)
	return fetchReleaseTag(proxyAPIURL)
}

func fetchReleaseTag(url string) string {
	client := &http.Client{Timeout: 10 * time.Second}

	requestURL := addCacheBuster(url)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("User-Agent", "idb-update-checker")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var release githubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return ""
	}

	return strings.TrimSpace(release.TagName)
}

func addCacheBuster(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	query := parsed.Query()
	query.Set("ts", fmt.Sprintf("%d", time.Now().UnixNano()))
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
