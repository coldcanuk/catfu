// Package brave implements a thin Brave Search API client.
package brave

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/coldcanuk/catfu/internal/backends"
)

const defaultBase = "https://api.search.brave.com/res/v1/web/search"

// Client is a Brave Web Search backend.
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// Name implements backends.Searcher.
func (c *Client) Name() string { return "brave" }

// Search implements backends.Searcher.
func (c *Client) Search(ctx context.Context, q backends.SearchQuery) ([]backends.Result, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("brave API key not configured (set BRAVE_API_KEY or --brave-api-key)")
	}
	if strings.TrimSpace(q.Query) == "" {
		return nil, fmt.Errorf("query is required")
	}
	base := c.BaseURL
	if base == "" {
		base = defaultBase
	}
	limit := q.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}
	if offset > 9 {
		offset = 9
	}

	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Set("q", q.Query)
	params.Set("count", strconv.Itoa(limit))
	params.Set("offset", strconv.Itoa(offset))
	if q.After != nil || q.Before != nil {
		// freshness custom range
		start := "1970-01-01"
		end := time.Now().UTC().Format("2006-01-02")
		if q.After != nil {
			start = q.After.UTC().Format("2006-01-02")
		}
		if q.Before != nil {
			end = q.Before.UTC().Format("2006-01-02")
		}
		params.Set("freshness", start+"to"+end)
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("X-Subscription-Token", c.APIKey)

	hc := c.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 30 * time.Second}
	}
	res, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("brave rate limited (429)")
	}
	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("brave auth failed (%d): check API key", res.StatusCode)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("brave HTTP %d: %s", res.StatusCode, truncate(string(body), 200))
	}

	var parsed braveResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode brave response: %w", err)
	}
	out := make([]backends.Result, 0, len(parsed.Web.Results))
	for _, r := range parsed.Web.Results {
		out = append(out, backends.Result{
			Title:       r.Title,
			URL:         r.URL,
			Description: r.Description,
			Source:      "brave",
		})
	}
	return out, nil
}

type braveResponse struct {
	Web struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		} `json:"results"`
	} `json:"web"`
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

var _ backends.Searcher = (*Client)(nil)
