// Package pac provides PAC (Proxy Auto-Config) file handling functionality.
package pac

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Fetcher handles fetching PAC files from URLs.
type Fetcher struct {
	client *http.Client
	cache  *pacCache
	mu     sync.RWMutex
}

// pacCache stores the cached PAC content.
type pacCache struct {
	content   string
	fetchedAt time.Time
	ttl       time.Duration
}

// NewFetcher creates a new PAC fetcher.
func NewFetcher(timeout time.Duration) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: timeout,
		},
		cache: &pacCache{
			ttl: 5 * time.Minute, // Default TTL
		},
	}
}

// Fetch retrieves the PAC file from the given URL.
// It returns cached content if available and not expired.
func (f *Fetcher) Fetch(ctx context.Context, url string) (string, error) {
	// Check cache first
	f.mu.RLock()
	cached := f.cache
	f.mu.RUnlock()

	if cached.content != "" && time.Since(cached.fetchedAt) < cached.ttl {
		return cached.content, nil
	}

	// Fetch from URL
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching PAC file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading PAC file: %w", err)
	}

	content := string(body)

	// Update cache
	f.mu.Lock()
	f.cache.content = content
	f.cache.fetchedAt = time.Now()
	f.mu.Unlock()

	return content, nil
}

// SetCacheTTL sets the cache TTL.
func (f *Fetcher) SetCacheTTL(ttl time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cache.ttl = ttl
}

// ClearCache clears the cached PAC content.
func (f *Fetcher) ClearCache() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cache.content = ""
	f.cache.fetchedAt = time.Time{}
}
