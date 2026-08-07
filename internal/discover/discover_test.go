package discover

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coldcanuk/catfu/internal/backends/brave"
	"github.com/coldcanuk/catfu/internal/store"
)

func TestDiscoverExtractsYouTube(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/videos/search") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type": "videos",
				"results": []map[string]any{
					{
						"title":       "Concurrency in Go",
						"url":         "https://www.youtube.com/watch?v=PfbFtY0aHbI",
						"description": "Talk from https://www.youtube.com/@golang",
					},
				},
			})
			return
		}
		// web fallback
		_ = json.NewEncoder(w).Encode(map[string]any{
			"web": map[string]any{"results": []any{}},
		})
	}))
	defer srv.Close()

	st, err := store.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	svc := &Service{
		Brave: &brave.Client{APIKey: "k", BaseHost: srv.URL, HTTPClient: srv.Client()},
		Store: st,
	}
	rep, err := svc.Discover(context.Background(), "golang concurrency", Options{Limit: 5, Kind: "video"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Videos) < 1 || rep.Videos[0].ID != "PfbFtY0aHbI" {
		t.Fatalf("videos=%+v", rep.Videos)
	}
	if len(rep.Channels) < 1 || rep.Channels[0].Handle != "@golang" {
		t.Fatalf("channels=%+v", rep.Channels)
	}
	if rep.Videos[0].InCatalogue {
		t.Fatal("expected not in catalogue")
	}
}
