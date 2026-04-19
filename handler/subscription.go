package handler

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout = 15 * time.Second
	userAgent      = "sub2api/1.0 (subscription converter)"
)

// FetchSubscription fetches raw subscription content from a remote URL.
// It follows redirects and returns the decoded body as a string.
func FetchSubscription(subURL string) (string, error) {
	if subURL == "" {
		return "", fmt.Errorf("subscription URL is empty")
	}

	parsed, err := url.ParseRequestURI(subURL)
	if err != nil {
		return "", fmt.Errorf("invalid subscription URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("unsupported URL scheme: %s", parsed.Scheme)
	}

	client := &http.Client{
		Timeout: defaultTimeout,
	}

	req, err := http.NewRequest(http.MethodGet, subURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch subscription: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	content := strings.TrimSpace(string(body))

	// Attempt base64 decode if content looks encoded
	if isBase64(content) {
		decoded, err := base64.StdEncoding.DecodeString(content)
		if err == nil {
			return strings.TrimSpace(string(decoded)), nil
		}
		// Try URL-safe base64
		decoded, err = base64.URLEncoding.DecodeString(content)
		if err == nil {
			return strings.TrimSpace(string(decoded)), nil
		}
	}

	return content, nil
}

// ParseNodes splits subscription content into individual proxy node lines.
func ParseNodes(content string) []string {
	lines := strings.Split(content, "\n")
	nodes := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		nodes = append(nodes, line)
	}
	return nodes
}

// isBase64 performs a naive check to determine if a string is likely base64-encoded.
// Note: this check allows newlines to be stripped before testing, which helps with
// some subscription providers that wrap lines at 76 characters.
func isBase64(s string) bool {
	// Strip newlines before checking, some providers wrap base64 output
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	if len(s) == 0 || len(s)%4 != 0 {
		return false
	}
	for _, c := range s {
		if !strings.ContainsRune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=_-", c) {
			return false
		}
	}
	return true
}
