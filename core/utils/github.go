package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// githubRelease represents the minimal fields from GitHub Releases API response.
type githubRelease struct {
	TagName string `json:"tag_name"`
}

// GetLatestReleaseVersion queries the GitHub Releases API for the latest release
// of the given repository (e.g. "sensdata/idb") and returns the tag name.
// Returns an empty string if the request fails or no release is found.
func GetLatestReleaseVersion(githubRepo string) string {
	if githubRepo == "" {
		return ""
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
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
