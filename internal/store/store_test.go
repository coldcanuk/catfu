package store

import (
	"context"
	"testing"
	"time"
)

func TestUpsertAndSearch(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	ch := Channel{ID: "UCtest", Title: "Test Channel", CustomURL: "@test"}
	if err := s.UpsertChannel(ctx, ch); err != nil {
		t.Fatal(err)
	}
	v1 := Video{ID: "vid1", ChannelID: "UCtest", Title: "Hello World News", Description: "breaking news", UploadDate: "2024-01-15", Duration: 120, WebpageURL: "https://www.youtube.com/watch?v=vid1"}
	v2 := Video{ID: "vid2", ChannelID: "UCtest", Title: "Cooking Pasta", Description: "recipe", UploadDate: "2024-06-01", Duration: 300}
	if err := s.UpsertVideo(ctx, v1); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertVideo(ctx, v2); err != nil {
		t.Fatal(err)
	}
	if err := s.TouchChannelCatalogued(ctx, "UCtest", time.Now()); err != nil {
		t.Fatal(err)
	}

	hits, err := s.SearchVideos(ctx, SearchParams{Query: "hello news", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].ID != "vid1" {
		t.Fatalf("expected vid1, got %+v", hits)
	}

	hits, err = s.SearchVideos(ctx, SearchParams{After: "2024-05-01", Before: "2024-12-31", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].ID != "vid2" {
		t.Fatalf("date filter expected vid2, got %+v", hits)
	}

	// Idempotent upsert
	v1.Title = "Hello World News Updated"
	if err := s.UpsertVideo(ctx, v1); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetVideo(ctx, "vid1")
	if err != nil || got == nil || got.Title != "Hello World News Updated" {
		t.Fatalf("upsert update failed: %+v %v", got, err)
	}

	st, err := s.Stats(ctx)
	if err != nil || st.Channels != 1 || st.Videos != 2 {
		t.Fatalf("stats: %+v %v", st, err)
	}
}

func TestDeleteChannelCascade(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	_ = s.UpsertChannel(ctx, Channel{ID: "UC1", Title: "C"})
	_ = s.UpsertVideo(ctx, Video{ID: "v1", ChannelID: "UC1", Title: "T"})
	if err := s.DeleteChannel(ctx, "UC1"); err != nil {
		t.Fatal(err)
	}
	v, err := s.GetVideo(ctx, "v1")
	if err != nil {
		t.Fatal(err)
	}
	if v != nil {
		t.Fatal("expected cascade delete")
	}
}
