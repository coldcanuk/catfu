// Package brave implements a thin Brave Search API client (Search plan).
//
// Auth: X-Subscription-Token with a Search product subscription key.
// Endpoints: web, news, videos under https://api.search.brave.com/res/v1/...
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

const (
	defaultHost      = "https://api.search.brave.com"
	pathWebSearch    = "/res/v1/web/search"
	pathNewsSearch   = "/res/v1/news/search"
	pathVideoSearch  = "/res/v1/videos/search"
	defaultUserAgent = "catfu/dev (+https://github.com/coldcanuk/catfu)"
)

// Client is a Brave Search backend (web / news / video).
type Client struct {
	APIKey     string
	BaseHost   string // optional override, default api.search.brave.com
	HTTPClient *http.Client
	UserAgent  string
}

// Name implements backends.Searcher.
func (c *Client) Name() string { return "brave" }

// Search implements backends.Searcher.
func (c *Client) Search(ctx context.Context, q backends.SearchQuery) ([]backends.Result, error) {
	if strings.TrimSpace(c.APIKey) == "" {
		return nil, fmt.Errorf("brave API key not configured (set BRAVE_API_KEY or --brave-api-key; Search plan subscription token)")
	}
	if strings.TrimSpace(q.Query) == "" {
		return nil, fmt.Errorf("query is required")
	}
	if len(q.Query) > 400 {
		return nil, fmt.Errorf("query too long (max 400 characters)")
	}

	kind := q.Kind
	if kind == "" {
		kind = backends.SearchKindWeb
	}
	path, maxCount, err := endpointForKind(kind)
	if err != nil {
		return nil, err
	}

	limit := q.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > maxCount {
		limit = maxCount
	}
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}
	if offset > 9 {
		offset = 9
	}

	host := c.BaseHost
	if host == "" {
		host = defaultHost
	}
	u, err := url.Parse(strings.TrimRight(host, "/") + path)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("q", q.Query)
	params.Set("count", strconv.Itoa(limit))
	params.Set("offset", strconv.Itoa(offset))
	if q.Country != "" {
		params.Set("country", strings.ToUpper(q.Country))
	}
	if q.SearchLang != "" {
		params.Set("search_lang", q.SearchLang)
	}
	if q.SafeSearch != "" {
		params.Set("safesearch", strings.ToLower(q.SafeSearch))
	}
	if f := freshnessParam(q); f != "" {
		params.Set("freshness", f)
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	// Do NOT set Accept-Encoding manually: net/http Transport auto-decompresses
	// gzip only when the header is unset by the caller.
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", c.APIKey)
	ua := c.UserAgent
	if ua == "" {
		ua = defaultUserAgent
	}
	req.Header.Set("User-Agent", ua)

	hc := c.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 30 * time.Second}
	}
	res, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return nil, err
	}

	if err := mapHTTPError(res, body); err != nil {
		return nil, err
	}

	return decodeResults(kind, body)
}

func endpointForKind(kind backends.SearchKind) (path string, maxCount int, err error) {
	switch kind {
	case backends.SearchKindWeb, "":
		return pathWebSearch, 20, nil
	case backends.SearchKindNews:
		return pathNewsSearch, 50, nil
	case backends.SearchKindVideo:
		return pathVideoSearch, 50, nil
	default:
		return "", 0, fmt.Errorf("unsupported brave search kind %q (use web, news, video)", kind)
	}
}

func freshnessParam(q backends.SearchQuery) string {
	if strings.TrimSpace(q.Freshness) != "" {
		return strings.TrimSpace(q.Freshness)
	}
	if q.After == nil && q.Before == nil {
		return ""
	}
	start := "1970-01-01"
	end := time.Now().UTC().Format("2006-01-02")
	if q.After != nil {
		start = q.After.UTC().Format("2006-01-02")
	}
	if q.Before != nil {
		end = q.Before.UTC().Format("2006-01-02")
	}
	return start + "to" + end
}

func mapHTTPError(res *http.Response, body []byte) error {
	code := res.StatusCode
	if code >= 200 && code < 300 {
		return nil
	}
	reset := res.Header.Get("X-RateLimit-Reset")
	remaining := res.Header.Get("X-RateLimit-Remaining")
	limit := res.Header.Get("X-RateLimit-Limit")
	apiCode, apiDetail := parseBraveAPIError(body)
	switch code {
	case http.StatusTooManyRequests:
		msg := "brave rate limited (429)"
		if reset != "" {
			msg += fmt.Sprintf("; retry after ~%ss", reset)
		}
		if remaining != "" || limit != "" {
			msg += fmt.Sprintf(" (remaining=%s limit=%s)", remaining, limit)
		}
		return fmt.Errorf("%s", msg)
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusPaymentRequired:
		return fmt.Errorf("brave auth/plan failed (%d): check Search plan subscription token (BRAVE_API_KEY)", code)
	case http.StatusUnprocessableEntity:
		if apiCode == "SUBSCRIPTION_TOKEN_INVALID" || strings.Contains(strings.ToLower(apiDetail), "subscription token") {
			return fmt.Errorf("brave subscription token invalid (422): use a Search plan key from https://api-dashboard.search.brave.com/ (not Answers); set BRAVE_API_KEY or --brave-api-key")
		}
		if apiDetail != "" {
			return fmt.Errorf("brave rejected request (422 %s): %s", apiCode, apiDetail)
		}
		return fmt.Errorf("brave rejected request (422): %s", truncate(string(body), 240))
	default:
		if apiDetail != "" {
			return fmt.Errorf("brave HTTP %d (%s): %s", code, apiCode, apiDetail)
		}
		return fmt.Errorf("brave HTTP %d: %s", code, truncate(string(body), 240))
	}
}

func parseBraveAPIError(body []byte) (code, detail string) {
	var wrap struct {
		Error struct {
			Code   string `json:"code"`
			Detail string `json:"detail"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil {
		return "", ""
	}
	return wrap.Error.Code, wrap.Error.Detail
}

func decodeResults(kind backends.SearchKind, body []byte) ([]backends.Result, error) {
	switch kind {
	case backends.SearchKindNews:
		return decodeFlatResults(body, "brave-news", string(kind))
	case backends.SearchKindVideo:
		return decodeFlatResults(body, "brave-video", string(kind))
	default:
		return decodeWebResults(body)
	}
}

func decodeWebResults(body []byte) ([]backends.Result, error) {
	var parsed struct {
		Web struct {
			Results []resultItem `json:"results"`
		} `json:"web"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode brave web response: %w", err)
	}
	out := make([]backends.Result, 0, len(parsed.Web.Results))
	for _, r := range parsed.Web.Results {
		out = append(out, r.toResult("brave", "web"))
	}
	return out, nil
}

func decodeFlatResults(body []byte, source, kind string) ([]backends.Result, error) {
	var top struct {
		Results []resultItem `json:"results"`
		News    *struct {
			Results []resultItem `json:"results"`
		} `json:"news"`
		Videos *struct {
			Results []resultItem `json:"results"`
		} `json:"videos"`
	}
	if err := json.Unmarshal(body, &top); err != nil {
		return nil, fmt.Errorf("decode brave %s response: %w", kind, err)
	}
	items := top.Results
	if len(items) == 0 && top.News != nil {
		items = top.News.Results
	}
	if len(items) == 0 && top.Videos != nil {
		items = top.Videos.Results
	}
	out := make([]backends.Result, 0, len(items))
	for _, r := range items {
		out = append(out, r.toResult(source, kind))
	}
	return out, nil
}

type resultItem struct {
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
	Age         string   `json:"age"`
	PageAge     string   `json:"page_age"`
	MetaURL     *metaURL `json:"meta_url"`
}

type metaURL struct {
	Hostname string `json:"hostname"`
}

func (r resultItem) toResult(source, kind string) backends.Result {
	age := r.Age
	if age == "" {
		age = r.PageAge
	}
	urlStr := r.URL
	if urlStr == "" && r.MetaURL != nil {
		urlStr = r.MetaURL.Hostname
	}
	return backends.Result{
		Title:       r.Title,
		URL:         urlStr,
		Description: r.Description,
		Source:      source,
		Kind:        kind,
		Age:         age,
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

var _ backends.Searcher = (*Client)(nil)
