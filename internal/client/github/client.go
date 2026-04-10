package github

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

type RepoResponse struct {
	FullName string `json:"full_name"`
}

type ReleaseResponse struct {
	TagName string `json:"tag_name"`
}

func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://api.github.com",
		token:      token,
	}
}

func (c *Client) RepositoryExists(owner, repo string) (bool, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, owner, repo)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "github-release-notification-api")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusOK {
		var result RepoResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return false, err
		}
		return true, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusForbidden {
		return false, fmt.Errorf("github rate limit exceeded")
	}

	return false, fmt.Errorf("unexpected github status: %d", resp.StatusCode)
}

func (c *Client) GetLatestReleaseTag(owner, repo string) (*string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.baseURL, owner, repo)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "github-release-notification-api")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close resp.Body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusOK {
		var result ReleaseResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}
		return &result.TagName, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("github rate limit exceeded")
	}

	return nil, fmt.Errorf("unexpected github status: %d", resp.StatusCode)
}
