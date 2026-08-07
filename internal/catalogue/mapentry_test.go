package catalogue

import (
	"encoding/json"
	"testing"
)

func TestMapEntryEngagementAndCaptions(t *testing.T) {
	raw := []byte(`{
		"id": "PfbFtY0aHbI",
		"title": "How to achieve concurrency",
		"description": "desc",
		"upload_date": "20240322",
		"duration": 118,
		"view_count": 3364,
		"like_count": 165,
		"comment_count": 5,
		"language": "en",
		"channel_id": "UCO3LEtymiLrgvpb59cNsb8A",
		"webpage_url": "https://www.youtube.com/watch?v=PfbFtY0aHbI",
		"subtitles": {"en-j3PyPqV-e1s": [{"url": "x"}]},
		"automatic_captions": {"en-orig": [], "fr": []}
	}`)
	var entry RawEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		t.Fatal(err)
	}
	m := MapEntry(entry, "UCfallback")
	if m.ID != "PfbFtY0aHbI" || m.Duration != 118 {
		t.Fatalf("%+v", m)
	}
	if m.ViewCount == nil || *m.ViewCount != 3364 {
		t.Fatalf("views %+v", m.ViewCount)
	}
	if m.LikeCount == nil || *m.LikeCount != 165 {
		t.Fatalf("likes %+v", m.LikeCount)
	}
	if m.CommentCount == nil || *m.CommentCount != 5 {
		t.Fatalf("comments %+v", m.CommentCount)
	}
	if m.Language != "en" {
		t.Fatalf("lang %q", m.Language)
	}
	if !m.HasSubtitles || !m.HasAutoCaptions || !m.HasTranscript {
		t.Fatalf("caption flags %+v", m)
	}
	// en from manual + auto, fr from auto
	want := map[string]bool{"en": true, "fr": true}
	for _, l := range m.Languages {
		delete(want, l)
	}
	if len(want) != 0 {
		t.Fatalf("langs %v leftover %v", m.Languages, want)
	}
}
