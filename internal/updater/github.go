package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitzen/internal/logger"
)

// GitHubClient handles interactions with GitHub API
type GitHubClient struct {
	repo   string
	client *http.Client
	logger *logger.Logger
}

// Release represents a GitHub release
type Release struct {
	TagName         string         `json:"tag_name"`
	Name            string         `json:"name"`
	Body            string         `json:"body"`
	Draft           bool           `json:"draft"`
	Prerelease      bool           `json:"prerelease"`
	CreatedAt       string         `json:"created_at"`
	PublishedAt     string         `json:"published_at"`
	Assets          []ReleaseAsset `json:"assets"`
	TarballURL      string         `json:"tarball_url"`
	ZipballURL      string         `json:"zipball_url"`
	ID              int64          `json:"id"`
	NodeID          string         `json:"node_id"`
	URL             string         `json:"url"`
	AssetsURL       string         `json:"assets_url"`
	UploadURL       string         `json:"upload_url"`
	HTMLURL         string         `json:"html_url"`
}

// ReleaseAsset represents a GitHub release asset (binary file)
type ReleaseAsset struct {
	Name               string `json:"name"`
	Label              string `json:"label"`
	State              string `json:"state"`
	ContentType        string `json:"content_type"`
	Size               int64  `json:"size"`
	DownloadCount      int64  `json:"download_count"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
	BrowserDownloadURL string `json:"browser_download_url"`
	ID                 int64  `json:"id"`
	NodeID             string `json:"node_id"`
	URL                string `json:"url"`
}

// NewGitHubClient creates a new GitHub API client
func NewGitHubClient(repo string) *GitHubClient {
	return &GitHubClient{
		repo: repo,
		client: &http.Client{
			Timeout: HTTPTimeout,
		},
		logger: logger.Get(),
	}
}

// GetLatestRelease fetches the latest release from GitHub
func (g *GitHubClient) GetLatestRelease() (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", g.repo)
	
	g.logger.Debug("Fetching latest release from: %s", url)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers for GitHub API
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", UserAgent)
	
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, resp.Status)
	}
	
	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release data: %w", err)
	}
	
	// Filter out drafts and prereleases by default
	if release.Draft {
		return nil, fmt.Errorf("latest release is a draft")
	}
	
	if release.Prerelease {
		return nil, fmt.Errorf("latest release is a prerelease")
	}
	
	g.logger.Debug("Found release: %s (created: %s)", release.TagName, release.CreatedAt)
	
	return &release, nil
}

// GetReleaseByTag fetches a specific release by tag name
func (g *GitHubClient) GetReleaseByTag(tag string) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", g.repo, tag)
	
	g.logger.Debug("Fetching release by tag from: %s", url)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", UserAgent)
	
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release not found: %s", tag)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, resp.Status)
	}
	
	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release data: %w", err)
	}
	
	return &release, nil
}

// ListReleases fetches all releases from GitHub (with pagination support)
func (g *GitHubClient) ListReleases(page, perPage int) ([]Release, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 30 // GitHub default
	}
	
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases?page=%d&per_page=%d", 
		g.repo, page, perPage)
	
	g.logger.Debug("Fetching releases from: %s", url)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", UserAgent)
	
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, resp.Status)
	}
	
	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode releases data: %w", err)
	}
	
	g.logger.Debug("Fetched %d releases", len(releases))
	
	return releases, nil
}

// CheckRateLimit checks the current rate limit status
func (g *GitHubClient) CheckRateLimit() (*RateLimit, error) {
	url := "https://api.github.com/rate_limit"
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", UserAgent)
	
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, resp.Status)
	}
	
	var rateLimit struct {
		Rate RateLimit `json:"rate"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&rateLimit); err != nil {
		return nil, fmt.Errorf("failed to decode rate limit data: %w", err)
	}
	
	return &rateLimit.Rate, nil
}

// RateLimit represents GitHub API rate limit information
type RateLimit struct {
	Limit     int   `json:"limit"`
	Used      int   `json:"used"`
	Remaining int   `json:"remaining"`
	Reset     int64 `json:"reset"`
}

// ResetTime returns the time when the rate limit resets
func (rl *RateLimit) ResetTime() time.Time {
	return time.Unix(rl.Reset, 0)
}

// TimeUntilReset returns the duration until the rate limit resets
func (rl *RateLimit) TimeUntilReset() time.Duration {
	return time.Until(rl.ResetTime())
}