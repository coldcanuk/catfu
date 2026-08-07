package store

import (
	"context"
	"testing"
)

func TestSearchEmptyQueryLists(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	_ = s.UpsertChannel(ctx, Channel{ID: "UC1", Title: "A"})
	_ = s.UpsertVideo(ctx, Video{ID: "v1", ChannelID: "UC1", Title: "Alpha", UploadDate: "2023-01-01"})
	_ = s.UpsertVideo(ctx, Video{ID: "v2", ChannelID: "UC1", Title: "Beta", UploadDate: "2023-02-01"})
	hits, err := s.SearchVideos(ctx, SearchParams{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 {
		t.Fatalf("got %d", len(hits))
	}
}
