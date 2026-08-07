package search

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coldcanuk/catfu/internal/backends"
	"github.com/coldcanuk/catfu/internal/backends/brave"
	"github.com/coldcanuk/catfu/internal/store"
)

func TestHybridMergesRemoteYouTube(t *testing.T) {
	st, err := store.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	_ = st.UpsertChannel(context.Background(), store.Channel{ID: "UCtest", Title: "T"})
	_ = st.UpsertVideo(context.Background(), store.Video{ID: "AAAAAAAAAAA", ChannelID: "UCtest", Title: "local only concurrency talk"})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]any{
				{"title": "remote concurrency", "url": "https://www.youtube.com/watch?v=BBBBBBBBBBB", "description": "x"},
				{"title": "already local", "url": "https://www.youtube.com/watch?v=AAAAAAAAAAA", "description": "x"},
			},
		})
	}))
	defer srv.Close()

	h := &Hybrid{
		Local: &CatalogueSearcher{Store: st},
		Brave: &brave.Client{APIKey: "k", BaseHost: srv.URL, HTTPClient: srv.Client()},
	}
	out, err := h.Search(context.Background(), backends.SearchQuery{Query: "concurrency", Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	nLocal := 0
	for _, r := range out {
		ids[r.ID] = true
		if r.ID == "AAAAAAAAAAA" {
			nLocal++
		}
	}
	if !ids["AAAAAAAAAAA"] || !ids["BBBBBBBBBBB"] {
		t.Fatalf("ids=%v", ids)
	}
	if nLocal != 1 {
		t.Fatalf("dup local: %+v", out)
	}
}
