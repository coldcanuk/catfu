package youtube

import "testing"

func TestVideoID(t *testing.T) {
	cases := map[string]string{
		"https://www.youtube.com/watch?v=PfbFtY0aHbI": "PfbFtY0aHbI",
		"https://youtu.be/PfbFtY0aHbI":                 "PfbFtY0aHbI",
		"https://www.youtube.com/shorts/PfbFtY0aHbI":  "PfbFtY0aHbI",
		"PfbFtY0aHbI": "PfbFtY0aHbI",
		"https://example.com": "",
	}
	for in, want := range cases {
		if got := VideoID(in); got != want {
			t.Errorf("VideoID(%q)=%q want %q", in, got, want)
		}
	}
}

func TestParseChannelRef(t *testing.T) {
	ch := ParseChannelRef("https://www.youtube.com/@golang/videos")
	if ch == nil || ch.Handle != "@golang" {
		t.Fatalf("%+v", ch)
	}
	ch = ParseChannelRef("https://www.youtube.com/channel/UCO3LEtymiLrgvpb59cNsb8A")
	if ch == nil || ch.ChannelID != "UCO3LEtymiLrgvpb59cNsb8A" {
		t.Fatalf("%+v", ch)
	}
	ch = ParseChannelRef("@golang")
	if ch == nil || ch.Handle != "@golang" {
		t.Fatalf("%+v", ch)
	}
}

func TestExtractFromText(t *testing.T) {
	vids, chs := ExtractFromText(`see https://youtu.be/PfbFtY0aHbI and channel https://www.youtube.com/@golang`)
	if len(vids) != 1 || vids[0] != "PfbFtY0aHbI" {
		t.Fatalf("vids=%v", vids)
	}
	if len(chs) != 1 || chs[0].Handle != "@golang" {
		t.Fatalf("chs=%+v", chs)
	}
}
