package brave

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coldcanuk/catfu/internal/backends"
)

func TestSearchWebOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Subscription-Token") != "test-key" {
			t.Errorf("missing token")
		}
		if r.Header.Get("Accept-Encoding") != "" {
			t.Errorf("Accept-Encoding should be unset by client, got %q", r.Header.Get("Accept-Encoding"))
		}
		if !strings.HasSuffix(r.URL.Path, "/res/v1/web/search") {
			t.Errorf("path %s", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "golang" {
			t.Errorf("q=%s", r.URL.Query().Get("q"))
		}
		if r.URL.Query().Get("country") != "CA" {
			t.Errorf("country=%s", r.URL.Query().Get("country"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"web": map[string]any{
				"results": []map[string]any{
					{"title": "Go", "url": "https://go.dev", "description": "The Go language", "age": "2 days ago"},
				},
			},
		})
	}))
	defer srv.Close()

	c := &Client{APIKey: "test-key", BaseHost: srv.URL, HTTPClient: srv.Client()}
	out, err := c.Search(context.Background(), backends.SearchQuery{
		Query:   "golang",
		Country: "ca",
		Limit:   5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Title != "Go" || out[0].Kind != "web" {
		t.Fatalf("%+v", out)
	}
}

func TestSearchVideoPathAndCountCap(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/res/v1/videos/search") {
			t.Errorf("path %s", r.URL.Path)
		}
		if r.URL.Query().Get("count") != "50" {
			t.Errorf("count=%s", r.URL.Query().Get("count"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type": "videos",
			"results": []map[string]any{
				{"title": "Talk", "url": "https://youtube.com/watch?v=x", "description": "conf"},
			},
		})
	}))
	defer srv.Close()
	c := &Client{APIKey: "k", BaseHost: srv.URL, HTTPClient: srv.Client()}
	out, err := c.Search(context.Background(), backends.SearchQuery{
		Query: "talk",
		Kind:  backends.SearchKindVideo,
		Limit: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Source != "brave-video" {
		t.Fatalf("%+v", out)
	}
}

func TestSearchNews(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/res/v1/news/search") {
			t.Errorf("path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]any{
				{"title": "Headline", "url": "https://news.example/a", "description": "story"},
			},
		})
	}))
	defer srv.Close()
	c := &Client{APIKey: "k", BaseHost: srv.URL, HTTPClient: srv.Client()}
	out, err := c.Search(context.Background(), backends.SearchQuery{
		Query: "ottawa",
		Kind:  backends.SearchKindNews,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Kind != "news" {
		t.Fatalf("%+v", out)
	}
}

func TestFreshnessFromDates(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query().Get("freshness")
		_ = json.NewEncoder(w).Encode(map[string]any{"web": map[string]any{"results": []any{}}})
	}))
	defer srv.Close()
	after := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	before := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	c := &Client{APIKey: "k", BaseHost: srv.URL, HTTPClient: srv.Client()}
	_, err := c.Search(context.Background(), backends.SearchQuery{Query: "x", After: &after, Before: &before})
	if err != nil {
		t.Fatal(err)
	}
	if got != "2024-01-01to2024-06-01" {
		t.Fatalf("freshness=%q", got)
	}
}

func TestRateLimitError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Reset", "2")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Limit", "1")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"type":"ErrorResponse"}`))
	}))
	defer srv.Close()
	c := &Client{APIKey: "k", BaseHost: srv.URL, HTTPClient: srv.Client()}
	_, err := c.Search(context.Background(), backends.SearchQuery{Query: "x"})
	if err == nil || !strings.Contains(err.Error(), "429") || !strings.Contains(err.Error(), "retry") {
		t.Fatalf("err=%v", err)
	}
}

func TestMissingKey(t *testing.T) {
	c := &Client{}
	_, err := c.Search(context.Background(), backends.SearchQuery{Query: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}
